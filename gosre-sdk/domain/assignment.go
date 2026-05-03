// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import "time"

// Assignment represents a single check-to-agent binding managed by the scheduler.
type Assignment struct {
	ID         string    `json:"id"`
	CheckID    string    `json:"check_id"`
	AgentID    string    `json:"agent_id"`
	AssignedAt time.Time `json:"assigned_at"`
}
