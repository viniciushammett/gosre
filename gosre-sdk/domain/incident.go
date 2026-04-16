// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import "time"

// IncidentState represents the current operational state of an incident.
type IncidentState string

const (
	IncidentStateOpen         IncidentState = "open"
	IncidentStateAcknowledged IncidentState = "acknowledged"
	IncidentStateResolved     IncidentState = "resolved"
)

// Incident is the derived operational state from repeated Result failures on a Target.
type Incident struct {
	ID        string        `json:"id"`
	TargetID  string        `json:"target_id"`
	State     IncidentState `json:"state"`
	FirstSeen time.Time     `json:"first_seen"`
	LastSeen  time.Time     `json:"last_seen"`
	ResultIDs []string      `json:"result_ids,omitempty"`
}
