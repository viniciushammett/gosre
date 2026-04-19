// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gosre/gosre-api/internal/service"
	"github.com/gosre/gosre-sdk/domain"
)

// ResultHandler handles HTTP requests for Result resources.
type ResultHandler struct {
	svc *service.ResultService
}

// NewResultHandler constructs a ResultHandler.
func NewResultHandler(svc *service.ResultService) *ResultHandler {
	return &ResultHandler{svc: svc}
}

type postResultRequest struct {
	ID         string            `json:"id"`
	CheckID    string            `json:"check_id"`
	TargetID   string            `json:"target_id"`
	AgentID    string            `json:"agent_id"`
	Status     domain.CheckStatus `json:"status"`
	DurationMs int64             `json:"duration_ms"`
	Error      string            `json:"error"`
	Timestamp  time.Time         `json:"timestamp"`
}

// PostResult handles POST /api/v1/results — used by gosre-agent to report check results.
func (h *ResultHandler) PostResult(c *gin.Context) {
	var body postResultRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	r := domain.Result{
		ID:        body.ID,
		CheckID:   body.CheckID,
		TargetID:  body.TargetID,
		AgentID:   body.AgentID,
		Status:    body.Status,
		Duration:  time.Duration(body.DurationMs) * time.Millisecond,
		Error:     body.Error,
		Timestamp: body.Timestamp,
	}

	saved, err := h.svc.Save(c.Request.Context(), r)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, saved)
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
