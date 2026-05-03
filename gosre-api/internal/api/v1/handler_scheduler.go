// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	agentDownThreshold    = 60 * time.Second
	schedulerVirtualNodes = 150
)

// SchedulerStatusResponse is the response body for GET /api/v1/scheduler/status.
type SchedulerStatusResponse struct {
	AgentsActive        int            `json:"agents_active"`
	AgentsDown          int            `json:"agents_down"`
	TotalChecks         int            `json:"total_checks"`
	AssignmentsPerAgent map[string]int `json:"assignments_per_agent"`
}

// SchedulerHandler handles scheduler observability endpoints.
type SchedulerHandler struct {
	agents agentStorer
	checks checkStorer
}

// NewSchedulerHandler constructs a SchedulerHandler.
func NewSchedulerHandler(agents agentStorer, checks checkStorer) *SchedulerHandler {
	return &SchedulerHandler{agents: agents, checks: checks}
}

// Status handles GET /api/v1/scheduler/status.
// It derives the current assignment distribution on-demand using the same
// consistent hash algorithm as gosre-scheduler (FNV-1a, 150 virtual nodes).
func (h *SchedulerHandler) Status(c *gin.Context) {
	agents, err := h.agents.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	checks, err := h.checks.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	now := time.Now().UTC()
	activeIDs := make([]string, 0, len(agents))
	down := 0
	for _, a := range agents {
		if now.Sub(a.LastSeen) <= agentDownThreshold {
			activeIDs = append(activeIDs, a.ID)
		} else {
			down++
		}
	}

	dist := make(map[string]int, len(activeIDs))
	for _, id := range activeIDs {
		dist[id] = 0
	}
	if len(activeIDs) > 0 {
		ring := buildStatusRing(activeIDs)
		for _, ch := range checks {
			if id := statusRingGet(ring, ch.ID); id != "" {
				dist[id]++
			}
		}
	}

	OK(c, http.StatusOK, SchedulerStatusResponse{
		AgentsActive:        len(activeIDs),
		AgentsDown:          down,
		TotalChecks:         len(checks),
		AssignmentsPerAgent: dist,
	})
}

// ── inline consistent hash (mirrors gosre-scheduler/internal/hash) ─────────

type statusRingNode struct {
	hash    uint32
	agentID string
}

func buildStatusRing(agentIDs []string) []statusRingNode {
	nodes := make([]statusRingNode, 0, len(agentIDs)*schedulerVirtualNodes)
	for _, id := range agentIDs {
		for i := range schedulerVirtualNodes {
			nodes = append(nodes, statusRingNode{
				hash:    schedulerFNV32(fmt.Sprintf("%s-%d", id, i)),
				agentID: id,
			})
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].hash < nodes[j].hash
	})
	return nodes
}

func statusRingGet(nodes []statusRingNode, key string) string {
	if len(nodes) == 0 {
		return ""
	}
	h := schedulerFNV32(key)
	idx := sort.Search(len(nodes), func(i int) bool {
		return nodes[i].hash >= h
	})
	if idx == len(nodes) {
		idx = 0
	}
	return nodes[idx].agentID
}

func schedulerFNV32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
