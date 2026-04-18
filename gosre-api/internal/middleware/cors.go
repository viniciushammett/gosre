// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that allows cross-origin requests.
// Origins are read from GOSRE_CORS_ORIGINS (comma-separated).
// If the env var is empty, all origins are allowed — suitable for development.
func CORS() gin.HandlerFunc {
	origins := os.Getenv("GOSRE_CORS_ORIGINS")

	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	if origins == "" {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = strings.Split(origins, ",")
	}

	return cors.New(cfg)
}
