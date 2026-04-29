// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/gosre/gosre-sdk/store"
	"github.com/viniciushammett/gosre/gosre-events/events"
	gosrejs "github.com/viniciushammett/gosre/gosre-events/jetstream"
)

// ResultService handles business logic for Result entities.
type ResultService struct {
	store     store.ResultStore
	publisher *gosrejs.Publisher
}

// NewResultService constructs a ResultService backed by the given store.
// pub may be nil — publishing is skipped when NATS is not configured.
func NewResultService(s store.ResultStore, pub *gosrejs.Publisher) *ResultService {
	return &ResultService{store: s, publisher: pub}
}

// Save assigns an ID if absent and persists a Result.
func (svc *ResultService) Save(ctx context.Context, r domain.Result) (domain.Result, error) {
	if r.ID == "" {
		r.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now().UTC()
	}
	if err := svc.store.Save(ctx, r); err != nil {
		return domain.Result{}, fmt.Errorf("save result: %w", err)
	}
	if svc.publisher != nil {
		_ = svc.publisher.Publish(ctx, events.SubjectResultsCreated, events.ResultCreatedPayload{
			EventEnvelope: events.NewEnvelope(),
			ResultID:      r.ID,
			CheckID:       r.CheckID,
			TargetID:      r.TargetID,
			TargetName:    r.TargetName,
			AgentID:       r.AgentID,
			Status:        string(r.Status),
			DurationMS:    r.Duration.Milliseconds(),
			Error:         r.Error,
			Timestamp:     r.Timestamp,
			Metadata:      r.Metadata,
		})
	}
	return r, nil
}

// Get retrieves a Result by ID.
func (svc *ResultService) Get(ctx context.Context, id string) (domain.Result, error) {
	return svc.store.Get(ctx, id)
}

// List returns all Results, scoped to targetID when non-empty.
func (svc *ResultService) List(ctx context.Context, targetID string) ([]domain.Result, error) {
	if targetID != "" {
		return svc.store.ListByTarget(ctx, targetID)
	}
	return svc.store.List(ctx)
}
