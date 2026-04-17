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

	v1 "github.com/gosre/gosre-api/internal/api/v1"
	"github.com/gosre/gosre-api/internal/service"
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

	db, err := sqlite.New("gosre.db")
	if err != nil {
		logger.Fatal("open store", zap.Error(err))
	}

	targetSvc := service.NewTargetService(db)
	targetHandler := v1.NewTargetHandler(targetSvc)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/healthz", v1.HealthHandler)

	api := router.Group("/api/v1")
	api.GET("/targets", targetHandler.ListTargets)
	api.POST("/targets", targetHandler.CreateTarget)
	api.GET("/targets/:id", targetHandler.GetTarget)
	api.DELETE("/targets/:id", targetHandler.DeleteTarget)

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
