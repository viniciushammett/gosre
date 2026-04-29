// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosre/gosre-sdk/domain"

	"github.com/viniciushammett/gosre/gosre-catalog/internal/service"
)

// EnvironmentHandler handles HTTP requests for Environment resources.
type EnvironmentHandler struct {
	svc *service.EnvironmentService
}

// NewEnvironmentHandler constructs an EnvironmentHandler.
func NewEnvironmentHandler(svc *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{svc: svc}
}

type createEnvironmentRequest struct {
	Name      string                 `json:"name"`
	ProjectID string                 `json:"project_id"`
	Kind      domain.EnvironmentKind `json:"kind"`
}

// ListByProject handles GET /api/v1/projects/:project_id/environments.
func (h *EnvironmentHandler) ListByProject(c *gin.Context) {
	envs, err := h.svc.ListByProject(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, envs)
}

// Create handles POST /api/v1/projects/:project_id/environments.
func (h *EnvironmentHandler) Create(c *gin.Context) {
	var req createEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	e, err := h.svc.Create(c.Request.Context(), domain.Environment{
		Name:      req.Name,
		ProjectID: c.Param("project_id"),
		Kind:      req.Kind,
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, e)
}

// Get handles GET /api/v1/environments/:id.
func (h *EnvironmentHandler) Get(c *gin.Context) {
	e, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrEnvironmentNotFound) {
		Fail(c, http.StatusNotFound, "environment_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, e)
}

// Delete handles DELETE /api/v1/environments/:id.
func (h *EnvironmentHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrEnvironmentNotFound) {
		Fail(c, http.StatusNotFound, "environment_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
