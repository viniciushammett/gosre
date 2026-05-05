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

// ProjectHandler handles HTTP requests for Project resources.
type ProjectHandler struct {
	svc *service.ProjectService
}

// NewProjectHandler constructs a ProjectHandler.
func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type createProjectRequest struct {
	Name   string `json:"name"`
	Slug   string `json:"slug,omitempty"`
	TeamID string `json:"team_id,omitempty"`
}

// ListByOrg handles GET /api/v1/organizations/:org_id/projects.
func (h *ProjectHandler) ListByOrg(c *gin.Context) {
	projects, err := h.svc.ListByOrg(c.Request.Context(), c.Param("org_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, projects)
}

// Create handles POST /api/v1/organizations/:org_id/projects.
func (h *ProjectHandler) Create(c *gin.Context) {
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	p, err := h.svc.Create(c.Request.Context(), domain.Project{
		OrganizationID: c.Param("org_id"),
		TeamID:         req.TeamID,
		Name:           req.Name,
		Slug:           req.Slug,
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, p)
}

// Get handles GET /api/v1/organizations/:org_id/projects/:id.
func (h *ProjectHandler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrProjectNotFound) {
		Fail(c, http.StatusNotFound, "project_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, p)
}

// Delete handles DELETE /api/v1/organizations/:org_id/projects/:id.
func (h *ProjectHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrProjectNotFound) {
		Fail(c, http.StatusNotFound, "project_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
