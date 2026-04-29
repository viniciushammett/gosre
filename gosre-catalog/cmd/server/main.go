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

	v1 "github.com/viniciushammett/gosre/gosre-catalog/internal/api/v1"
	"github.com/viniciushammett/gosre/gosre-catalog/internal/service"
	"github.com/viniciushammett/gosre/gosre-catalog/internal/store/azuresql"
	"github.com/viniciushammett/gosre/gosre-catalog/internal/store/sqlite"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	var (
		catalogSvc *service.CatalogService
		depSvc     *service.DependencyService
		envSvc     *service.EnvironmentService
	)

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		logger.Info("using azuresql store", zap.String("url", dbURL))
		az, err := azuresql.New(dbURL)
		if err != nil {
			logger.Fatal("open azuresql store", zap.Error(err))
		}
		defer func() { _ = az.Close() }()
		catalogSvc = service.NewCatalogService(az.ServiceStore())
		depSvc = service.NewDependencyService(az.DependencyStore())
		envSvc = service.NewEnvironmentService(az.EnvironmentStore())
	} else {
		logger.Info("using sqlite store", zap.String("path", "catalog.db"))
		lite, err := sqlite.New("catalog.db")
		if err != nil {
			logger.Fatal("open sqlite store", zap.Error(err))
		}
		defer func() { _ = lite.Close() }()
		catalogSvc = service.NewCatalogService(lite.ServiceStore())
		depSvc = service.NewDependencyService(lite.DependencyStore())
		envSvc = service.NewEnvironmentService(lite.EnvironmentStore())
	}

	serviceH := v1.NewServiceHandler(catalogSvc)
	depH := v1.NewDependencyHandler(depSvc)
	envH := v1.NewEnvironmentHandler(envSvc)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")

	api.GET("/services", serviceH.List)
	api.POST("/services", serviceH.Create)
	api.GET("/services/:id", serviceH.Get)
	api.DELETE("/services/:id", serviceH.Delete)

	api.GET("/dependencies", depH.List)
	api.POST("/dependencies", depH.Create)
	api.GET("/dependencies/:id", depH.Get)
	api.DELETE("/dependencies/:id", depH.Delete)

	api.GET("/projects/:project_id/environments", envH.ListByProject)
	api.POST("/projects/:project_id/environments", envH.Create)
	api.GET("/environments/:id", envH.Get)
	api.DELETE("/environments/:id", envH.Delete)

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
