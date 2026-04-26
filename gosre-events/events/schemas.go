// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package events

import (
	"crypto/rand"
	"fmt"
	"time"
)

// EventEnvelope is embedded in every event payload.
// EventID enables idempotent processing and downstream dedup.
// OccurredAt is the moment the event was emitted (distinct from any payload timestamp).
type EventEnvelope struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

// NewEnvelope returns an EventEnvelope with a fresh UUIDv4 and the current UTC time.
// All event constructors should call this to ensure the envelope is always populated.
func NewEnvelope() EventEnvelope {
	return EventEnvelope{
		EventID:    newEventID(),
		OccurredAt: time.Now().UTC(),
	}
}

// newEventID returns a random UUID v4 using crypto/rand.
// crypto/rand.Read cannot return an error on any supported Go platform (Go 1.20+).
func newEventID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Unreachable on supported platforms; included for correctness.
		panic("gosre-events: crypto/rand unavailable: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ResultCreatedPayload is the payload for SubjectResultsCreated.
// Published by gosre-api after a check result is persisted.
// Consumed by the incident detection service and gosre-notifier.
type ResultCreatedPayload struct {
	EventEnvelope
	ResultID   string            `json:"result_id"`
	CheckID    string            `json:"check_id"`
	TargetID   string            `json:"target_id"`
	TargetName string            `json:"target_name,omitempty"`
	AgentID    string            `json:"agent_id,omitempty"`
	Status     string            `json:"status"`
	DurationMS int64             `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	ProjectID  string            `json:"project_id,omitempty"`
}

// IncidentPayload is the shared structure for incident lifecycle events.
// The subject name — SubjectIncidentsOpened or SubjectIncidentsResolved — is
// what distinguishes an open event from a resolve event.
type IncidentPayload struct {
	EventEnvelope
	IncidentID string    `json:"incident_id"`
	TargetID   string    `json:"target_id"`
	State      string    `json:"state"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	ResultIDs  []string  `json:"result_ids,omitempty"`
	ProjectID  string    `json:"project_id,omitempty"`
}

// IncidentOpenedPayload is the payload for SubjectIncidentsOpened.
// Published by gosre-api when 3 consecutive check failures open an incident.
// Consumed by gosre-notifier for alert delivery.
type IncidentOpenedPayload = IncidentPayload

// IncidentResolvedPayload is the payload for SubjectIncidentsResolved.
// Published by gosre-api on auto-resolve (1 success) or manual state transition.
// Consumed by gosre-notifier for resolution notification.
type IncidentResolvedPayload = IncidentPayload

// AgentHeartbeatPayload is the payload for SubjectAgentsHeartbeat.
// Published by gosre-agent alongside the existing HTTP POST heartbeat.
// Consumed by gosre-scheduler to track agent liveness for check reassignment.
type AgentHeartbeatPayload struct {
	EventEnvelope
	AgentID  string `json:"agent_id"`
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

// CheckAssignedPayload is the payload for SubjectChecksAssigned.
// Published by gosre-scheduler when a check is assigned or reassigned to an agent.
// Consumed by gosre-agent as a push alternative to the HTTP poll assignments endpoint.
type CheckAssignedPayload struct {
	EventEnvelope
	AgentID   string `json:"agent_id"`
	CheckID   string `json:"check_id"`
	TargetID  string `json:"target_id"`
	CheckType string `json:"check_type"`
}
