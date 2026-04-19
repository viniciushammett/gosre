// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gosre/gosre-api/internal/store/sqlite"
	"github.com/gosre/gosre-sdk/domain"
)

// AgentHandler handles HTTP requests for agent lifecycle endpoints.
type AgentHandler struct {
	agents *sqlite.AgentStore
	checks *sqlite.CheckStore
}

// NewAgentHandler constructs an AgentHandler.
func NewAgentHandler(agents *sqlite.AgentStore, checks *sqlite.CheckStore) *AgentHandler {
	return &AgentHandler{agents: agents, checks: checks}
}

type registerAgentRequest struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

type agentResponse struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	Version  string    `json:"version"`
	LastSeen time.Time `json:"last_seen"`
}

// List handles GET /api/v1/agents.
func (h *AgentHandler) List(c *gin.Context) {
	recs, err := h.agents.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	out := make([]agentResponse, 0, len(recs))
	for _, r := range recs {
		out = append(out, agentResponse{
			ID:       r.ID,
			Hostname: r.Hostname,
			Version:  r.Version,
			LastSeen: r.LastSeen,
		})
	}
	OK(c, http.StatusOK, out)
}

// Register handles POST /api/v1/agents/register.
func (h *AgentHandler) Register(c *gin.Context) {
	var body registerAgentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	rec := sqlite.AgentRecord{
		ID:       uuid.New().String(),
		Hostname: body.Hostname,
		Version:  body.Version,
		LastSeen: time.Now().UTC(),
	}
	if err := h.agents.Register(c.Request.Context(), rec); err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	OK(c, http.StatusCreated, agentResponse{
		ID:       rec.ID,
		Hostname: rec.Hostname,
		Version:  rec.Version,
		LastSeen: rec.LastSeen,
	})
}

// Heartbeat handles POST /api/v1/agents/:id/heartbeat.
func (h *AgentHandler) Heartbeat(c *gin.Context) {
	id := c.Param("id")
	if err := h.agents.Heartbeat(c.Request.Context(), id); err != nil {
		Fail(c, http.StatusNotFound, "agent_not_found", "agent not found")
		return
	}
	c.Status(http.StatusNoContent)
}

type assignmentResponse struct {
	CheckID  string           `json:"check_id"`
	TargetID string           `json:"target_id"`
	Type     domain.CheckType `json:"type"`
}

// Assignments handles GET /api/v1/agents/:id/assignments.
func (h *AgentHandler) Assignments(c *gin.Context) {
	checks, err := h.checks.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	out := make([]assignmentResponse, 0, len(checks))
	for _, ch := range checks {
		out = append(out, assignmentResponse{
			CheckID:  ch.ID,
			TargetID: ch.TargetID,
			Type:     ch.Type,
		})
	}
	OK(c, http.StatusOK, out)
}
