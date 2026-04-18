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

// IncidentHandler handles HTTP requests for Incident resources.
type IncidentHandler struct {
	svc *service.IncidentService
}

// NewIncidentHandler constructs an IncidentHandler.
func NewIncidentHandler(svc *service.IncidentService) *IncidentHandler {
	return &IncidentHandler{svc: svc}
}

// ListIncidents handles GET /api/v1/incidents.
// Accepts optional query param: state (open|acknowledged|resolved).
func (h *IncidentHandler) ListIncidents(c *gin.Context) {
	state := domain.IncidentState(c.Query("state"))
	incidents, err := h.svc.List(c.Request.Context(), state)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, incidents)
}

// PatchIncident handles PATCH /api/v1/incidents/:id.
// Body: {"state": "acknowledged"|"resolved"}
func (h *IncidentHandler) PatchIncident(c *gin.Context) {
	var body struct {
		State domain.IncidentState `json:"state" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	inc, err := h.svc.UpdateState(c.Request.Context(), c.Param("id"), body.State)
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "incident_not_found", "incident not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusBadRequest, "transition_error", err.Error())
		return
	}
	OK(c, http.StatusOK, inc)
}
