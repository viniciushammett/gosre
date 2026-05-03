// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosre/gosre-sdk/domain"

	"github.com/gosre/gosre-api/internal/service"
)

// SLOHandler handles HTTP requests for SLO resources.
type SLOHandler struct {
	svc *service.SLOService
}

// NewSLOHandler constructs a SLOHandler.
func NewSLOHandler(svc *service.SLOService) *SLOHandler {
	return &SLOHandler{svc: svc}
}

// List handles GET /api/v1/slos?target_id=xxx.
func (h *SLOHandler) List(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		Fail(c, http.StatusBadRequest, "target_id_required", "target_id query parameter is required")
		return
	}
	slos, err := h.svc.ListByTarget(c.Request.Context(), targetID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, slos)
}

// Create handles POST /api/v1/slos.
func (h *SLOHandler) Create(c *gin.Context) {
	var body domain.SLO
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	slo, err := h.svc.Create(c.Request.Context(), body)
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, slo)
}

// Get handles GET /api/v1/slos/:id.
func (h *SLOHandler) Get(c *gin.Context) {
	slo, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "slo_not_found", "slo not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, slo)
}

// Delete handles DELETE /api/v1/slos/:id.
func (h *SLOHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "slo_not_found", "slo not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusNoContent, nil)
}

// Budget handles GET /api/v1/slos/:id/budget.
func (h *SLOHandler) Budget(c *gin.Context) {
	result, err := h.svc.Budget(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "slo_not_found", "slo not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, result)
}
