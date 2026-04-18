// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gosre/gosre-api/internal/service"
)

// ResultHandler handles HTTP requests for Result resources.
type ResultHandler struct {
	svc *service.ResultService
}

// NewResultHandler constructs a ResultHandler.
func NewResultHandler(svc *service.ResultService) *ResultHandler {
	return &ResultHandler{svc: svc}
}

// ListResults handles GET /api/v1/results.
// Accepts optional query param: target_id.
func (h *ResultHandler) ListResults(c *gin.Context) {
	results, err := h.svc.List(c.Request.Context(), c.Query("target_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, results)
}

// GetResult handles GET /api/v1/results/:id.
func (h *ResultHandler) GetResult(c *gin.Context) {
	r, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "result_not_found", "result not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, r)
}
