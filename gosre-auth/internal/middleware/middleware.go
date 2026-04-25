// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package middleware provides JWT validation middleware for Gin routes.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gosre/gosre-auth/internal/domain"
	"github.com/gosre/gosre-auth/internal/service"
)

// claimsKey is the gin.Context key used to store validated JWT claims.
const claimsKey = "claims"

// JWT returns a Gin middleware that validates Bearer tokens from the Authorization header.
// On success, it stores domain.Claims under the "claims" key in the request context.
func JWT(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := svc.ValidateToken(c.Request.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

// GetClaims retrieves validated domain.Claims from the Gin context.
// Returns false if the middleware did not run or the token was rejected.
func GetClaims(c *gin.Context) (domain.Claims, bool) {
	v, ok := c.Get(claimsKey)
	if !ok {
		return domain.Claims{}, false
	}
	claims, ok := v.(domain.Claims)
	return claims, ok
}
