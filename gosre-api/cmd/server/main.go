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
	"github.com/gosre/gosre-api/internal/store/postgres"
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
		targetSvc   *service.TargetService
		resultSvc   *service.ResultService
		incidentSvc *service.IncidentService
		checkSvc    *service.CheckService
	)

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		logger.Info("using postgres store", zap.String("url", dbURL))
		pg, err := postgres.New(dbURL)
		if err != nil {
			logger.Fatal("open postgres store", zap.Error(err))
		}
		targetSvc = service.NewTargetService(pg)
		resultSvc = service.NewResultService(pg.ResultStore())
		incidentSvc = service.NewIncidentService(pg.IncidentStore(), pg.ResultStore())
		checkSvc = service.NewCheckService(pg.CheckStore(), pg, resultSvc, incidentSvc, checkers)
	} else {
		logger.Info("using sqlite store", zap.String("path", "gosre.db"))
		lite, err := sqlite.New("gosre.db")
		if err != nil {
			logger.Fatal("open sqlite store", zap.Error(err))
		}
		targetSvc = service.NewTargetService(lite)
		resultSvc = service.NewResultService(lite.ResultStore())
		incidentSvc = service.NewIncidentService(lite.IncidentStore(), lite.ResultStore())
		checkSvc = service.NewCheckService(lite.CheckStore(), lite, resultSvc, incidentSvc, checkers)
	}

	targetHandler := v1.NewTargetHandler(targetSvc)
	resultHandler := v1.NewResultHandler(resultSvc)
	incidentHandler := v1.NewIncidentHandler(incidentSvc)
	checkHandler := v1.NewCheckHandler(checkSvc)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.APIKey())

	router.GET("/healthz", v1.HealthHandler)

	api := router.Group("/api/v1")
	api.GET("/targets", targetHandler.ListTargets)
	api.POST("/targets", targetHandler.CreateTarget)
	api.GET("/targets/:id", targetHandler.GetTarget)
	api.DELETE("/targets/:id", targetHandler.DeleteTarget)

	api.GET("/checks", checkHandler.ListChecks)
	api.POST("/checks", checkHandler.CreateCheck)
	api.POST("/checks/:id/run", checkHandler.RunCheck)

	api.GET("/results", resultHandler.ListResults)
	api.GET("/results/:id", resultHandler.GetResult)

	api.GET("/incidents", incidentHandler.ListIncidents)
	api.PATCH("/incidents/:id", incidentHandler.PatchIncident)

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
