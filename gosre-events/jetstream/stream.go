// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package jetstream provides a NATS JetStream wrapper for publishing and
// subscribing to GoSRE events. All event subjects live under the single
// GOSRE stream with subject filter "gosre.>".
package jetstream

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	// streamName is the JetStream stream that covers all gosre.* subjects.
	streamName = "GOSRE"

	// streamSubjectFilter matches every GoSRE event subject.
	streamSubjectFilter = "gosre.>"
)

// EnsureStream creates the GOSRE stream if it does not already exist.
// CreateOrUpdateStream is idempotent — safe to call on every service startup.
func EnsureStream(ctx context.Context, js jetstream.JetStream) error {
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamSubjectFilter},
	})
	if err != nil {
		return fmt.Errorf("ensure stream %s: %w", streamName, err)
	}
	return nil
}
