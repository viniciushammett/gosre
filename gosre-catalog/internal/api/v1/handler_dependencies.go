// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"

	"github.com/viniciushammett/gosre/gosre-catalog/internal/service"
)

// DependencyHandler handles HTTP requests for Dependency resources.
type DependencyHandler struct {
	svc *service.DependencyService
}

// NewDependencyHandler constructs a DependencyHandler.
func NewDependencyHandler(svc *service.DependencyService) *DependencyHandler {
	return &DependencyHandler{svc: svc}
}

type createDependencyRequest struct {
	SourceServiceID string                `json:"source_service_id"`
	TargetServiceID string                `json:"target_service_id"`
	Kind            domain.DependencyKind `json:"kind"`
}

// List handles GET /api/v1/dependencies.
// Accepts ?source= or ?target= to filter by service ID.
func (h *DependencyHandler) List(c *gin.Context) {
	if src := c.Query("source"); src != "" {
		deps, err := h.svc.ListBySource(c.Request.Context(), src)
		if err != nil {
			Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		OK(c, http.StatusOK, deps)
		return
	}
	if tgt := c.Query("target"); tgt != "" {
		deps, err := h.svc.ListByTarget(c.Request.Context(), tgt)
		if err != nil {
			Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		OK(c, http.StatusOK, deps)
		return
	}
	Fail(c, http.StatusBadRequest, "filter_required", "provide ?source=<id> or ?target=<id>")
}

// Create handles POST /api/v1/dependencies.
func (h *DependencyHandler) Create(c *gin.Context) {
	var req createDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	d, err := h.svc.Create(c.Request.Context(), domain.Dependency{
		SourceServiceID: req.SourceServiceID,
		TargetServiceID: req.TargetServiceID,
		Kind:            req.Kind,
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, d)
}

// Get handles GET /api/v1/dependencies/:id.
func (h *DependencyHandler) Get(c *gin.Context) {
	d, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrDependencyNotFound) {
		Fail(c, http.StatusNotFound, "dependency_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, d)
}

// Delete handles DELETE /api/v1/dependencies/:id.
func (h *DependencyHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrDependencyNotFound) {
		Fail(c, http.StatusNotFound, "dependency_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
