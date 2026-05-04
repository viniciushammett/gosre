// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	natsjets "github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/gosre/gosre-sdk/domain"
	gosrejs "github.com/viniciushammett/gosre/gosre-events/jetstream"

	v1 "github.com/gosre/gosre-api/internal/api/v1"
	"github.com/gosre/gosre-api/internal/check"
	"github.com/gosre/gosre-api/internal/consumer"
	"github.com/gosre/gosre-api/internal/middleware"
	"github.com/gosre/gosre-api/internal/service"
	"github.com/gosre/gosre-api/internal/store/azuresql"
	"github.com/gosre/gosre-api/internal/store/sqlite"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("GOSRE_JWT_SECRET")
	if jwtSecret == "" {
		logger.Warn("GOSRE_JWT_SECRET not set; using insecure dev default — do not use in production")
		jwtSecret = "dev-secret-change-in-production"
	}

	checkers := map[domain.CheckType]domain.Checker{
		domain.CheckTypeHTTP: check.NewHTTPChecker(),
		domain.CheckTypeTCP:  check.NewTCPChecker(),
		domain.CheckTypeDNS:  check.NewDNSChecker(),
		domain.CheckTypeTLS:  check.NewTLSChecker(),
	}

	var (
		targetSvc   *service.TargetService
		resultSvc   *service.ResultService
		incidentSvc *service.IncidentService
		checkSvc    *service.CheckService
		agentH      *v1.AgentHandler
		schedulerH  *v1.SchedulerHandler
		sloSvc      *service.SLOService
		sloH        *v1.SLOHandler
		orgSvc      *service.OrgService
		teamSvc     *service.TeamService
		projectSvc  *service.ProjectService
		notifSvc    *service.NotificationService
		notifH      *v1.NotificationHandler
	)

	var (
		pub *gosrejs.Publisher
		js  natsjets.JetStream
	)
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		nc, err := nats.Connect(natsURL)
		if err != nil {
			logger.Fatal("connect nats", zap.Error(err))
		}
		defer func() { _ = nc.Drain() }()
		js, err = natsjets.New(nc)
		if err != nil {
			logger.Fatal("jetstream client", zap.Error(err))
		}
		initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer initCancel()
		if err := gosrejs.EnsureStream(initCtx, js); err != nil {
			logger.Fatal("ensure nats stream", zap.Error(err))
		}
		pub = gosrejs.NewPublisher(js)
		logger.Info("nats connected", zap.String("url", natsURL))
	} else {
		logger.Warn("NATS_URL not set; event publishing disabled")
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		logger.Info("using azuresql store", zap.String("url", dbURL))
		az, err := azuresql.New(dbURL)
		if err != nil {
			logger.Fatal("open azuresql store", zap.Error(err))
		}
		defer func() { _ = az.Close() }()
		targetSvc = service.NewTargetService(az.TargetStore(), az.CheckStore(), az.ResultStore(), az.IncidentStore())
		resultSvc = service.NewResultService(az.ResultStore(), pub)
		incidentSvc = service.NewIncidentService(az.IncidentStore(), az.ResultStore(), pub)
		checkSvc = service.NewCheckService(az.CheckStore(), az.TargetStore(), resultSvc, checkers)
		agentH = v1.NewAgentHandler(az.AgentStore(), az.CheckStore())
		schedulerH = v1.NewSchedulerHandler(az.AgentStore(), az.CheckStore())
		sloSvc = service.NewSLOService(az.SLOStore(), az.ResultStore())
		sloH = v1.NewSLOHandler(sloSvc)
		orgSvc = service.NewOrgService(az.OrgStore())
		teamSvc = service.NewTeamService(az.TeamStore())
		projectSvc = service.NewProjectService(az.ProjectStore())
		notifSvc = service.NewNotificationService(az.NotificationChannelStore(), az.NotificationRuleStore())
		notifH = v1.NewNotificationHandler(notifSvc)
	} else {
		logger.Info("using sqlite store", zap.String("path", "gosre.db"))
		lite, err := sqlite.New("gosre.db")
		if err != nil {
			logger.Fatal("open sqlite store", zap.Error(err))
		}
		targetSvc = service.NewTargetService(lite, lite.CheckStore(), lite.ResultStore(), lite.IncidentStore())
		resultSvc = service.NewResultService(lite.ResultStore(), pub)
		incidentSvc = service.NewIncidentService(lite.IncidentStore(), lite.ResultStore(), pub)
		checkSvc = service.NewCheckService(lite.CheckStore(), lite, resultSvc, checkers)
		agentH = v1.NewAgentHandler(lite.AgentStore(), lite.CheckStore())
		schedulerH = v1.NewSchedulerHandler(lite.AgentStore(), lite.CheckStore())
		sloSvc = service.NewSLOService(lite.SLOStore(), lite.ResultStore())
		sloH = v1.NewSLOHandler(sloSvc)
		orgSvc = service.NewOrgService(lite.OrgStore())
		teamSvc = service.NewTeamService(lite.TeamStore())
		projectSvc = service.NewProjectService(lite.ProjectStore())
		notifSvc = service.NewNotificationService(lite.NotificationChannelStore(), lite.NotificationRuleStore())
		notifH = v1.NewNotificationHandler(notifSvc)
	}

	targetH := v1.NewTargetHandler(targetSvc)
	resultH := v1.NewResultHandler(resultSvc)
	incidentH := v1.NewIncidentHandler(incidentSvc)
	checkH := v1.NewCheckHandler(checkSvc)
	orgH := v1.NewOrgHandler(orgSvc)
	teamH := v1.NewTeamHandler(teamSvc)
	projectH := v1.NewProjectHandler(projectSvc)
	catalogProxy := v1.NewCatalogProxy()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Public — no authentication required
	router.GET("/healthz", v1.HealthHandler)
	router.POST("/api/v1/agents/register", agentH.Register)

	// Agent operational — API key only (used by gosre-agent process)
	agentOps := router.Group("/api/v1/agents")
	agentOps.Use(middleware.APIKey())
	agentOps.POST("/:id/heartbeat", agentH.Heartbeat)
	agentOps.GET("/:id/assignments", agentH.Assignments)

	// JWT-protected — viewer+ (any authenticated user)
	api := router.Group("/api/v1")
	api.Use(middleware.JWT(jwtSecret))

	api.GET("/targets", targetH.ListTargets)
	api.GET("/targets/:id", targetH.GetTarget)
	api.GET("/checks", checkH.ListChecks)
	api.GET("/results", resultH.ListResults)
	api.GET("/results/:id", resultH.GetResult)
	api.GET("/incidents", incidentH.ListIncidents)
	api.GET("/agents", agentH.List)
	api.GET("/scheduler/status", schedulerH.Status)
	api.GET("/slos", sloH.List)
	api.GET("/slos/:id", sloH.Get)
	api.GET("/slos/:id/budget", sloH.Budget)
	api.GET("/organizations", orgH.List)
	api.GET("/organizations/:org_id", orgH.Get)
	api.GET("/organizations/:org_id/teams", teamH.ListByOrg)
	api.GET("/organizations/:org_id/teams/:id", teamH.Get)
	api.GET("/organizations/:org_id/projects", projectH.ListByOrg)
	api.GET("/organizations/:org_id/projects/:id", projectH.Get)
	api.GET("/notification/channels", notifH.ListChannels)
	api.GET("/notification/channels/:id", notifH.GetChannel)
	api.GET("/notification/rules", notifH.ListRules)
	api.GET("/notification/rules/:id", notifH.GetRule)
	api.GET("/catalog/services", catalogProxy.Proxy)
	api.GET("/catalog/services/:id", catalogProxy.Proxy)
	api.GET("/catalog/dependencies", catalogProxy.Proxy)
	api.GET("/catalog/environments", catalogProxy.Proxy)

	// JWT-protected — operator+ (write access)
	op := api.Group("/")
	op.Use(middleware.RequireRole("operator", "admin", "owner"))
	op.POST("/slos", sloH.Create)
	op.POST("/notification/channels", notifH.CreateChannel)
	op.POST("/notification/rules", notifH.CreateRule)
	op.POST("/catalog/services", catalogProxy.Proxy)
	op.POST("/catalog/dependencies", catalogProxy.Proxy)
	op.POST("/catalog/environments", catalogProxy.Proxy)
	op.POST("/targets", targetH.CreateTarget)
	op.PUT("/targets/:id", targetH.UpdateTarget)
	op.POST("/checks", checkH.CreateCheck)
	op.POST("/checks/:id/run", checkH.RunCheck)
	op.POST("/results", resultH.PostResult)
	op.PATCH("/incidents/:id", incidentH.PatchIncident)
	op.POST("/organizations", orgH.Create)
	op.POST("/organizations/:org_id/teams", teamH.Create)
	op.POST("/organizations/:org_id/projects", projectH.Create)

	// JWT-protected — admin+ (destructive operations)
	adm := api.Group("/")
	adm.Use(middleware.RequireRole("admin", "owner"))
	adm.DELETE("/slos/:id", sloH.Delete)
	adm.DELETE("/catalog/services/:id", catalogProxy.Proxy)
	adm.DELETE("/catalog/dependencies/:id", catalogProxy.Proxy)
	adm.DELETE("/catalog/environments/:id", catalogProxy.Proxy)
	adm.DELETE("/targets/:id", targetH.DeleteTarget)
	adm.DELETE("/organizations/:org_id", orgH.Delete)
	adm.DELETE("/organizations/:org_id/teams/:id", teamH.Delete)
	adm.DELETE("/organizations/:org_id/projects/:id", projectH.Delete)
	adm.DELETE("/notification/channels/:id", notifH.DeleteChannel)
	adm.DELETE("/notification/rules/:id", notifH.DeleteRule)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if js != nil {
		if err := consumer.StartIncidentDetector(ctx, js, incidentSvc, logger); err != nil {
			logger.Fatal("start incident detector", zap.Error(err))
		}
	}

	go func() {
		logger.Info("server started", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
	logger.Info("server stopped")
}
