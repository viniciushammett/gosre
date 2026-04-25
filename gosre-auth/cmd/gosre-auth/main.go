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
	"github.com/gosre/gosre-auth/internal/domain"
	"github.com/gosre/gosre-auth/internal/middleware"
	"github.com/gosre/gosre-auth/internal/service"
	"github.com/gosre/gosre-auth/internal/store"
)

type registerRequest struct {
	Email    string      `json:"email"    binding:"required"`
	Password string      `json:"password" binding:"required"`
	Role     domain.Role `json:"role"     binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	cfg := config.Load()
	svc := service.New(store.NewMemoryStore(), cfg.JWTSecret)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gosre-auth"})
	})

	auth := router.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) {
			var req registerRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			u, err := svc.Register(c.Request.Context(), req.Email, req.Password, req.Role)
			if err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"id":         u.ID,
				"email":      u.Email,
				"role":       u.Role,
				"created_at": u.CreatedAt,
			})
		})

		auth.POST("/login", func(c *gin.Context) {
			var req loginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			token, err := svc.Login(c.Request.Context(), req.Email, req.Password)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"token": token})
		})

		me := auth.Group("/")
		me.Use(middleware.JWT(svc))
		me.GET("me", func(c *gin.Context) {
			claims, ok := middleware.GetClaims(c)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "claims not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"user_id": claims.UserID,
				"email":   claims.Email,
				"role":    claims.Role,
			})
		})
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
