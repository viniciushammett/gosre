// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosre/gosre-sdk/domain"

	"github.com/gosre/gosre-api/internal/service"
)

// OrgHandler handles HTTP requests for Organization resources.
type OrgHandler struct {
	svc *service.OrgService
}

// NewOrgHandler constructs an OrgHandler.
func NewOrgHandler(svc *service.OrgService) *OrgHandler {
	return &OrgHandler{svc: svc}
}

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// List handles GET /api/v1/organizations.
func (h *OrgHandler) List(c *gin.Context) {
	orgs, err := h.svc.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, orgs)
}

// Create handles POST /api/v1/organizations.
func (h *OrgHandler) Create(c *gin.Context) {
	var req createOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	o, err := h.svc.Create(c.Request.Context(), domain.Organization{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, o)
}

// Get handles GET /api/v1/organizations/:org_id.
func (h *OrgHandler) Get(c *gin.Context) {
	o, err := h.svc.Get(c.Request.Context(), c.Param("org_id"))
	if errors.Is(err, service.ErrOrgNotFound) {
		Fail(c, http.StatusNotFound, "org_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, o)
}

// Delete handles DELETE /api/v1/organizations/:org_id.
func (h *OrgHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("org_id"))
	if errors.Is(err, service.ErrOrgNotFound) {
		Fail(c, http.StatusNotFound, "org_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
