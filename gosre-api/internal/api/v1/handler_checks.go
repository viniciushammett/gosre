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

// CheckHandler handles HTTP requests for CheckConfig resources.
type CheckHandler struct {
	svc *service.CheckService
}

// NewCheckHandler constructs a CheckHandler.
func NewCheckHandler(svc *service.CheckService) *CheckHandler {
	return &CheckHandler{svc: svc}
}

// ListChecks handles GET /api/v1/checks.
func (h *CheckHandler) ListChecks(c *gin.Context) {
	checks, err := h.svc.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, checks)
}

// CreateCheck handles POST /api/v1/checks.
func (h *CheckHandler) CreateCheck(c *gin.Context) {
	var body domain.CheckConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	check, err := h.svc.Create(c.Request.Context(), body)
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, check)
}

// RunCheck handles POST /api/v1/checks/:id/run.
func (h *CheckHandler) RunCheck(c *gin.Context) {
	r, err := h.svc.Run(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "check_not_found", "check not found")
		return
	}
	if errors.Is(err, domain.ErrTargetNotFound) {
		Fail(c, http.StatusNotFound, "target_not_found", err.Error())
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, r)
}
