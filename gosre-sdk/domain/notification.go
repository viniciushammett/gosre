// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

// NotificationChannelKind identifies the delivery mechanism for a channel.
type NotificationChannelKind string

const (
	ChannelKindSlack   NotificationChannelKind = "slack"
	ChannelKindEmail   NotificationChannelKind = "email"
	ChannelKindWebhook NotificationChannelKind = "webhook"
)

// NotificationChannel defines a delivery endpoint for incident alerts.
// Config holds channel-specific settings (e.g. webhook_url, smtp_host, to).
type NotificationChannel struct {
	ID        string                  `json:"id"`
	ProjectID string                  `json:"project_id"`
	Name      string                  `json:"name"`
	Kind      NotificationChannelKind `json:"kind"`
	Config    map[string]string       `json:"config"`
}

// NotificationRule routes incident events to a channel.
// TagFilter matches against target tags — empty slice matches all events.
// EventKind is the NATS subject name (e.g. "gosre.incidents.opened").
type NotificationRule struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	ChannelID string   `json:"channel_id"`
	EventKind string   `json:"event_kind"`
	TagFilter []string `json:"tag_filter,omitempty"`
}
