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

// TeamHandler handles HTTP requests for Team resources.
type TeamHandler struct {
	svc *service.TeamService
}

// NewTeamHandler constructs a TeamHandler.
func NewTeamHandler(svc *service.TeamService) *TeamHandler {
	return &TeamHandler{svc: svc}
}

type createTeamRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// ListByOrg handles GET /api/v1/organizations/:org_id/teams.
func (h *TeamHandler) ListByOrg(c *gin.Context) {
	teams, err := h.svc.ListByOrg(c.Request.Context(), c.Param("org_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, teams)
}

// Create handles POST /api/v1/organizations/:org_id/teams.
func (h *TeamHandler) Create(c *gin.Context) {
	var req createTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	t, err := h.svc.Create(c.Request.Context(), domain.Team{
		OrganizationID: c.Param("org_id"),
		Name:           req.Name,
		Slug:           req.Slug,
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, t)
}

// Get handles GET /api/v1/organizations/:org_id/teams/:id.
func (h *TeamHandler) Get(c *gin.Context) {
	t, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrTeamNotFound) {
		Fail(c, http.StatusNotFound, "team_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, t)
}

// Delete handles DELETE /api/v1/organizations/:org_id/teams/:id.
func (h *TeamHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrTeamNotFound) {
		Fail(c, http.StatusNotFound, "team_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
