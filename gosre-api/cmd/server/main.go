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
	"go.uber.org/zap"

	"github.com/gosre/gosre-sdk/domain"

	v1 "github.com/gosre/gosre-api/internal/api/v1"
	"github.com/gosre/gosre-api/internal/check"
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

	checkers := map[domain.CheckType]domain.Checker{
		domain.CheckTypeHTTP: check.NewHTTPChecker(),
		domain.CheckTypeTCP:  check.NewTCPChecker(),
		domain.CheckTypeDNS:  check.NewDNSChecker(),
		domain.CheckTypeTLS:  check.NewTLSChecker(),
	}

	var (
		targetSvc    *service.TargetService
		resultSvc    *service.ResultService
		incidentSvc  *service.IncidentService
		checkSvc     *service.CheckService
		agentHandler *v1.AgentHandler
		orgSvc       *service.OrgService
		teamSvc      *service.TeamService
		projectSvc   *service.ProjectService
	)

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		logger.Info("using azuresql store", zap.String("url", dbURL))
		az, err := azuresql.New(dbURL)
		if err != nil {
			logger.Fatal("open azuresql store", zap.Error(err))
		}
		defer func() { _ = az.Close() }()
		targetSvc = service.NewTargetService(az.TargetStore(), az.CheckStore(), az.ResultStore(), az.IncidentStore())
		resultSvc = service.NewResultService(az.ResultStore())
		incidentSvc = service.NewIncidentService(az.IncidentStore(), az.ResultStore())
		checkSvc = service.NewCheckService(az.CheckStore(), az.TargetStore(), resultSvc, incidentSvc, checkers)
		agentHandler = v1.NewAgentHandler(az.AgentStore(), az.CheckStore())
	} else {
		logger.Info("using sqlite store", zap.String("path", "gosre.db"))
		lite, err := sqlite.New("gosre.db")
		if err != nil {
			logger.Fatal("open sqlite store", zap.Error(err))
		}
		targetSvc = service.NewTargetService(lite, lite.CheckStore(), lite.ResultStore(), lite.IncidentStore())
		resultSvc = service.NewResultService(lite.ResultStore())
		incidentSvc = service.NewIncidentService(lite.IncidentStore(), lite.ResultStore())
		checkSvc = service.NewCheckService(lite.CheckStore(), lite, resultSvc, incidentSvc, checkers)
		agentHandler = v1.NewAgentHandler(lite.AgentStore(), lite.CheckStore())
	}

	targetHandler := v1.NewTargetHandler(targetSvc)
	resultHandler := v1.NewResultHandler(resultSvc)
	incidentHandler := v1.NewIncidentHandler(incidentSvc)
	checkHandler := v1.NewCheckHandler(checkSvc)
	orgHandler := v1.NewOrgHandler(orgSvc)
	teamHandler := v1.NewTeamHandler(teamSvc)
	projectHandler := v1.NewProjectHandler(projectSvc)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	router.GET("/healthz", v1.HealthHandler)

	api := router.Group("/api/v1")
	api.Use(middleware.APIKey())
	api.GET("/targets", targetHandler.ListTargets)
	api.POST("/targets", targetHandler.CreateTarget)
	api.GET("/targets/:id", targetHandler.GetTarget)
	api.PUT("/targets/:id", targetHandler.UpdateTarget)
	api.DELETE("/targets/:id", targetHandler.DeleteTarget)

	api.GET("/checks", checkHandler.ListChecks)
	api.POST("/checks", checkHandler.CreateCheck)
	api.POST("/checks/:id/run", checkHandler.RunCheck)

	api.GET("/results", resultHandler.ListResults)
	api.GET("/results/:id", resultHandler.GetResult)
	api.POST("/results", resultHandler.PostResult)

	api.GET("/incidents", incidentHandler.ListIncidents)
	api.PATCH("/incidents/:id", incidentHandler.PatchIncident)

	api.GET("/agents", agentHandler.List)
	api.POST("/agents/register", agentHandler.Register)
	api.POST("/agents/:id/heartbeat", agentHandler.Heartbeat)
	api.GET("/agents/:id/assignments", agentHandler.Assignments)

	orgs := api.Group("/organizations")
	orgs.GET("", orgHandler.List)
	orgs.POST("", orgHandler.Create)
	orgs.GET("/:id", orgHandler.Get)
	orgs.DELETE("/:id", orgHandler.Delete)

	orgTeams := api.Group("/organizations/:org_id/teams")
	orgTeams.GET("", teamHandler.ListByOrg)
	orgTeams.POST("", teamHandler.Create)
	orgTeams.GET("/:id", teamHandler.Get)
	orgTeams.DELETE("/:id", teamHandler.Delete)

	orgProjects := api.Group("/organizations/:org_id/projects")
	orgProjects.GET("", projectHandler.ListByOrg)
	orgProjects.POST("", projectHandler.Create)
	orgProjects.GET("/:id", projectHandler.Get)
	orgProjects.DELETE("/:id", projectHandler.Delete)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
