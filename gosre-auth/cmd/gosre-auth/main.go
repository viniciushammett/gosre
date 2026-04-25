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

	"github.com/gosre/gosre-auth/internal/config"
	"github.com/gosre/gosre-auth/internal/handler"
	"github.com/gosre/gosre-auth/internal/middleware"
	authoidc "github.com/gosre/gosre-auth/internal/oidc"
	"github.com/gosre/gosre-auth/internal/service"
	"github.com/gosre/gosre-auth/internal/store"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	cfg := config.Load()

	sessions := resolveSessionStore(cfg.RedisURL, logger)
	svc := service.New(store.NewMemoryStore(), sessions, cfg.JWTSecret)

	var ghProvider *authoidc.GitHubProvider
	if cfg.GitHubClientID != "" {
		ghProvider = authoidc.NewGitHubProvider(
			cfg.GitHubClientID,
			cfg.GitHubClientSecret,
			cfg.GitHubRedirectURL,
		)
	}

	h := handler.New(svc, ghProvider)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gosre-auth"})
	})

	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
		auth.GET("/github/login", h.GitHubLogin)
		auth.GET("/github/callback", h.GitHubCallback)

		protected := auth.Group("/")
		protected.Use(middleware.JWT(svc))
		protected.GET("me", h.Me)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("gosre-auth started", zap.String("port", cfg.Port))
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
	logger.Info("gosre-auth stopped")
}

// resolveSessionStore tries to connect to Redis; falls back to in-memory on failure.
func resolveSessionStore(redisURL string, logger *zap.Logger) store.SessionStore {
	rs, err := store.NewRedisSessionStore(redisURL)
	if err != nil {
		logger.Warn("redis url invalid; using in-memory session store", zap.Error(err))
		return store.NewMemorySessionStore()
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rs.Ping(pingCtx); err != nil {
		logger.Warn("redis unavailable; using in-memory session store", zap.Error(err))
		return store.NewMemorySessionStore()
	}

	logger.Info("using redis session store", zap.String("url", redisURL))
	return rs
}
