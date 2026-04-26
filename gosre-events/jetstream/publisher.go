// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package jetstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// Publisher publishes events to NATS JetStream.
// Each Publish call blocks until the server acknowledges the message or
// ctx is cancelled — JetStream at-least-once delivery is guaranteed.
type Publisher struct {
	js jetstream.JetStream
}

// NewPublisher constructs a Publisher backed by the given JetStream instance.
func NewPublisher(js jetstream.JetStream) *Publisher {
	return &Publisher{js: js}
}

// Publish serializes payload to JSON and publishes it to subject.
// The caller should use a subject constant from the events package (e.g. events.SubjectResultsCreated).
func (p *Publisher) Publish(ctx context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload for %s: %w", subject, err)
	}
	if _, err := p.js.Publish(ctx, subject, data); err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}
