// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/microsoft/go-mssqldb"
	"go.uber.org/zap"

	"github.com/viniciushammett/gosre/gosre-auth/internal/config"
	"github.com/viniciushammett/gosre/gosre-auth/internal/handler"
	"github.com/viniciushammett/gosre/gosre-auth/internal/middleware"
	authoidc "github.com/viniciushammett/gosre/gosre-auth/internal/oidc"
	"github.com/viniciushammett/gosre/gosre-auth/internal/service"
	"github.com/viniciushammett/gosre/gosre-auth/internal/store"
	"github.com/viniciushammett/gosre/gosre-auth/internal/store/sqlstore"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	cfg := config.Load()

	sessions := resolveSessionStore(cfg.RedisURL, logger)
	svc := service.New(resolveUserStore(cfg.DatabaseURL, logger), sessions, cfg.JWTSecret)

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

// resolveUserStore opens a SQL user store when DATABASE_URL is set; falls back to in-memory.
func resolveUserStore(databaseURL string, logger *zap.Logger) store.UserStore {
	if databaseURL == "" {
		logger.Warn("DATABASE_URL not set; using in-memory user store")
		return store.NewMemoryStore()
	}
	db, err := sql.Open("sqlserver", databaseURL)
	if err != nil {
		logger.Fatal("open database", zap.Error(err))
	}
	if err := sqlstore.RunMigrations(db); err != nil {
		logger.Fatal("run migrations", zap.Error(err))
	}
	logger.Info("using sql user store")
	return sqlstore.NewSQLUserStore(db)
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
