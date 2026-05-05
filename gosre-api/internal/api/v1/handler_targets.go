// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"

	"github.com/viniciushammett/gosre/gosre-api/internal/service"
)

// TargetHandler handles HTTP requests for Target resources.
type TargetHandler struct {
	svc *service.TargetService
}

// NewTargetHandler constructs a TargetHandler.
func NewTargetHandler(svc *service.TargetService) *TargetHandler {
	return &TargetHandler{svc: svc}
}

// ListTargets handles GET /api/v1/targets.
func (h *TargetHandler) ListTargets(c *gin.Context) {
	targets, err := h.svc.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, targets)
}

// CreateTarget handles POST /api/v1/targets.
func (h *TargetHandler) CreateTarget(c *gin.Context) {
	var body domain.Target
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	t, err := h.svc.Create(c.Request.Context(), body)
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, t)
}

// GetTarget handles GET /api/v1/targets/:id.
func (h *TargetHandler) GetTarget(c *gin.Context) {
	t, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, domain.ErrTargetNotFound) {
		Fail(c, http.StatusNotFound, "target_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, t)
}

// UpdateTarget handles PUT /api/v1/targets/:id.
func (h *TargetHandler) UpdateTarget(c *gin.Context) {
	var body domain.Target
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	t, err := h.svc.Update(c.Request.Context(), c.Param("id"), body)
	if errors.Is(err, domain.ErrTargetNotFound) {
		Fail(c, http.StatusNotFound, "target_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusOK, t)
}

// DeleteTarget handles DELETE /api/v1/targets/:id.
func (h *TargetHandler) DeleteTarget(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, domain.ErrTargetNotFound) {
		Fail(c, http.StatusNotFound, "target_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
