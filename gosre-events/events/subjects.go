// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package events defines typed event payloads and subject name constants
// for the GoSRE NATS JetStream event bus.
package events

const (
	// SubjectResultsCreated is published by gosre-api when a check result is persisted.
	SubjectResultsCreated = "gosre.results.created"

	// SubjectIncidentsOpened is published by gosre-api when an incident is opened.
	SubjectIncidentsOpened = "gosre.incidents.opened"

	// SubjectIncidentsResolved is published by gosre-api when an incident is resolved.
	SubjectIncidentsResolved = "gosre.incidents.resolved"

	// SubjectAgentsHeartbeat is published by gosre-agent on every heartbeat cycle.
	SubjectAgentsHeartbeat = "gosre.agents.heartbeat"

	// SubjectChecksAssigned is published by gosre-scheduler when checks are (re)assigned to an agent.
	SubjectChecksAssigned = "gosre.checks.assigned"
)
