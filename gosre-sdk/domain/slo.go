// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

// SLO represents a Service Level Objective defined for a monitored target.
// WindowSeconds is stored as int64 (not time.Duration) to avoid JSON nanosecond
// serialization per L-001. Convert with time.Duration(slo.WindowSeconds) * time.Second.
type SLO struct {
	ID            string  `json:"id"`
	TargetID      string  `json:"target_id"`
	Name          string  `json:"name"`
	Metric        string  `json:"metric"`
	Threshold     float64 `json:"threshold"`
	WindowSeconds int64   `json:"window_seconds"`
}
