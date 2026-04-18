// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKey returns a Gin middleware that enforces X-API-Key authentication.
// If GOSRE_API_KEY is not set the middleware is a no-op.
func APIKey() gin.HandlerFunc {
	key := os.Getenv("GOSRE_API_KEY")
	if key == "" {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if c.GetHeader("X-API-Key") != key {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"data": nil,
				"error": gin.H{
					"code":    "unauthorized",
					"message": "invalid or missing API key",
				},
			})
			return
		}
		c.Next()
	}
}
