// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import (
	"context"
	"time"
)

// CheckType identifies the kind of check to execute against a Target.
type CheckType string

const (
	CheckTypeHTTP CheckType = "http"
	CheckTypeTCP  CheckType = "tcp"
	CheckTypeDNS  CheckType = "dns"
	CheckTypeTLS  CheckType = "tls"
)

// CheckConfig defines a validation to run against a Target.
type CheckConfig struct {
	ID       string            `json:"id"`
	Type     CheckType         `json:"type"`
	TargetID string            `json:"target_id"`
	Interval time.Duration     `json:"interval"`
	Timeout  time.Duration     `json:"timeout"`
	Params    map[string]string `json:"params,omitempty"`
	ProjectID string            `json:"project_id,omitempty"`
}

// Checker executes a check against a Target and returns a Result.
// ctx must be respected for cancellation and timeout propagation (L-003, L-006).
type Checker interface {
	Execute(ctx context.Context, t Target, cfg CheckConfig) Result
}
