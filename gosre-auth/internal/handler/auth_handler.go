// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package handler implements the HTTP handlers for gosre-auth endpoints.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gosre/gosre-auth/internal/domain"
	"github.com/gosre/gosre-auth/internal/middleware"
	"github.com/gosre/gosre-auth/internal/service"
	"github.com/gosre/gosre-auth/internal/store"
)

// Handler holds the dependencies for the auth HTTP handlers.
type Handler struct {
	svc *service.AuthService
}

// New returns a Handler backed by the given AuthService.
func New(svc *service.AuthService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRequest is the body accepted by POST /auth/register.
type RegisterRequest struct {
	Email    string      `json:"email"    binding:"required"`
	Password string      `json:"password" binding:"required,min=8"`
	Role     domain.Role `json:"role"     binding:"required"`
}

// LoginRequest is the body accepted by POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is returned on a successful login.
type LoginResponse struct {
	Token string `json:"token"`
}

// MeResponse is returned by GET /auth/me.
type MeResponse struct {
	UserID string      `json:"user_id"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
}

// Register handles POST /auth/register.
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if errors.Is(err, store.ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         u.ID,
		"email":      u.Email,
		"role":       u.Role,
		"created_at": u.CreatedAt,
	})
}

// Login handles POST /auth/login.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// Me handles GET /auth/me (requires JWT middleware).
func (h *Handler) Me(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "claims not found"})
		return
	}
	c.JSON(http.StatusOK, MeResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	})
}
