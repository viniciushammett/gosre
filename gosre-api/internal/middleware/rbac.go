// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const claimsKey = "jwt_claims"

// Claims holds the identity fields extracted from a validated JWT.
type Claims struct {
	UserID string
	Email  string
	Role   string
}

// JWT returns a Gin middleware that validates HS256 Bearer tokens.
// On success, Claims are stored under the "jwt_claims" key in the context.
func JWT(secret string) gin.HandlerFunc {
	key := []byte(secret)

	return func(c *gin.Context) {
		if apiKey := os.Getenv("GOSRE_API_KEY"); apiKey != "" {
			if c.GetHeader("X-API-Key") == apiKey {
				c.Set(claimsKey, Claims{UserID: "exporter", Email: "", Role: "operator"})
				c.Next()
				return
			}
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abortUnauthorized(c, "missing bearer token")
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return key, nil
		}, jwt.WithValidMethods([]string{"HS256"}))

		if err != nil || !token.Valid {
			abortUnauthorized(c, "invalid or expired token")
			return
		}

		mc, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			abortUnauthorized(c, "malformed token claims")
			return
		}

		userID, ok1 := mc["user_id"].(string)
		email, ok2 := mc["email"].(string)
		role, ok3 := mc["role"].(string)

		if !ok1 || !ok2 || !ok3 || userID == "" || role == "" {
			abortUnauthorized(c, "incomplete token claims")
			return
		}

		c.Set(claimsKey, Claims{UserID: userID, Email: email, Role: role})
		c.Next()
	}
}

// RequireRole returns a middleware that allows access only to callers whose
// role is listed in the allowed set. Must run after JWT middleware.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			abortUnauthorized(c, "missing claims")
			return
		}
		if _, permitted := allowed[claims.Role]; !permitted {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"data": nil,
				"error": gin.H{
					"code":    "forbidden",
					"message": "insufficient role",
				},
			})
			return
		}
		c.Next()
	}
}

// GetClaims retrieves the validated Claims from the Gin context.
func GetClaims(c *gin.Context) (Claims, bool) {
	v, ok := c.Get(claimsKey)
	if !ok {
		return Claims{}, false
	}
	claims, ok := v.(Claims)
	return claims, ok
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"data": nil,
		"error": gin.H{
			"code":    "unauthorized",
			"message": message,
		},
	})
}
