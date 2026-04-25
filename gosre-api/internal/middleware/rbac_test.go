// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gosre/gosre-api/internal/middleware"
)

const testSecret = "test-secret-key"

func init() {
	gin.SetMode(gin.TestMode)
}

func signToken(t *testing.T, userID, email, role string, exp time.Duration) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(exp).Unix(),
		"iat":     time.Now().Unix(),
	})
	s, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func newRouter(roles ...string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.JWT(testSecret))
	if len(roles) > 0 {
		r.Use(middleware.RequireRole(roles...))
	}
	r.GET("/test", func(c *gin.Context) {
		claims, _ := middleware.GetClaims(c)
		c.JSON(http.StatusOK, gin.H{"role": claims.Role})
	})
	return r
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	r := newRouter()
	token := signToken(t, "uid-1", "alice@example.com", "viewer", time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	r := newRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	r := newRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	r := newRouter("operator", "admin", "owner")
	token := signToken(t, "uid-2", "bob@example.com", "operator", time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	r := newRouter("operator", "admin", "owner")
	token := signToken(t, "uid-3", "carol@example.com", "viewer", time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
