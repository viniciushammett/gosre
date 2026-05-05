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

// ServiceHandler handles HTTP requests for Service catalog resources.
type ServiceHandler struct {
	svc *service.CatalogService
}

// NewServiceHandler constructs a ServiceHandler.
func NewServiceHandler(svc *service.CatalogService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

type createServiceRequest struct {
	Name        string                    `json:"name"`
	Owner       string                    `json:"owner"`
	Criticality domain.ServiceCriticality `json:"criticality"`
	RunbookURL  string                    `json:"runbook_url"`
	RepoURL     string                    `json:"repo_url"`
	ProjectID   string                    `json:"project_id"`
}

// List handles GET /api/v1/services.
// Accepts optional ?project_id= query param to filter by project.
func (h *ServiceHandler) List(c *gin.Context) {
	if pid := c.Query("project_id"); pid != "" {
		svcs, err := h.svc.ListByProject(c.Request.Context(), pid)
		if err != nil {
			Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		OK(c, http.StatusOK, svcs)
		return
	}
	svcs, err := h.svc.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, svcs)
}

// Create handles POST /api/v1/services.
func (h *ServiceHandler) Create(c *gin.Context) {
	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	s, err := h.svc.Create(c.Request.Context(), domain.Service{
		Name:        req.Name,
		Owner:       req.Owner,
		Criticality: req.Criticality,
		RunbookURL:  req.RunbookURL,
		RepoURL:     req.RepoURL,
		ProjectID:   req.ProjectID,
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, s)
}

// Get handles GET /api/v1/services/:id.
func (h *ServiceHandler) Get(c *gin.Context) {
	s, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrServiceNotFound) {
		Fail(c, http.StatusNotFound, "service_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, s)
}

// Delete handles DELETE /api/v1/services/:id.
func (h *ServiceHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrServiceNotFound) {
		Fail(c, http.StatusNotFound, "service_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
