// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import "time"

// CheckStatus represents the outcome of a single check execution.
type CheckStatus string

const (
	StatusOK      CheckStatus = "ok"
	StatusFail    CheckStatus = "fail"
	StatusTimeout CheckStatus = "timeout"
	StatusUnknown CheckStatus = "unknown"
)

// Result is the structured output of a single check execution.
// Duration is stored as time.Duration internally; serialize as duration_ms (L-001).
type Result struct {
	ID        string            `json:"id"`
	CheckID   string            `json:"check_id"`
	TargetID  string            `json:"target_id"`
	AgentID   string            `json:"agent_id,omitempty"`
	Status    CheckStatus       `json:"status"`
	Duration  time.Duration     `json:"duration_ms"`
	Error     string            `json:"error,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	ProjectID string            `json:"project_id,omitempty"`
}
