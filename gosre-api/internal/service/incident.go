// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/gosre/gosre-sdk/store"
)

const consecutiveFailuresThreshold = 3

// IncidentService handles incident detection and state management.
type IncidentService struct {
	store   store.IncidentStore
	results store.ResultStore
}

// NewIncidentService constructs an IncidentService.
func NewIncidentService(s store.IncidentStore, results store.ResultStore) *IncidentService {
	return &IncidentService{store: s, results: results}
}

// Process evaluates a saved Result and opens or resolves an Incident as needed.
// It must be called after each result is persisted.
func (svc *IncidentService) Process(ctx context.Context, r domain.Result) error {
	if r.Status == domain.StatusOK {
		return svc.resolveIfOpen(ctx, r)
	}
	return svc.openIfThresholdReached(ctx, r)
}

// List returns incidents filtered by state. An empty state returns all incidents.
func (svc *IncidentService) List(ctx context.Context, state domain.IncidentState) ([]domain.Incident, error) {
	return svc.store.ListByState(ctx, state)
}

// Get returns an Incident by ID. Returns sql.ErrNoRows if not found.
func (svc *IncidentService) Get(ctx context.Context, id string) (domain.Incident, error) {
	return svc.store.Get(ctx, id)
}

// UpdateState applies a state transition to the incident identified by id.
// Valid transitions: open→acknowledged, open→resolved, acknowledged→resolved.
func (svc *IncidentService) UpdateState(ctx context.Context, id string, state domain.IncidentState) (domain.Incident, error) {
	inc, err := svc.store.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Incident{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.Incident{}, fmt.Errorf("get incident %q: %w", id, err)
	}

	if !validTransition(inc.State, state) {
		return domain.Incident{}, fmt.Errorf("invalid transition %q → %q", inc.State, state)
	}

	inc.State = state
	inc.LastSeen = time.Now().UTC()

	if err := svc.store.Update(ctx, inc); err != nil {
		return domain.Incident{}, fmt.Errorf("update incident %q: %w", id, err)
	}
	return inc, nil
}

func (svc *IncidentService) openIfThresholdReached(ctx context.Context, r domain.Result) error {
	recent, err := svc.results.ListByTarget(ctx, r.TargetID)
	if err != nil {
		return fmt.Errorf("incident: list results for target %q: %w", r.TargetID, err)
	}

	if len(recent) < consecutiveFailuresThreshold {
		return nil
	}

	for _, res := range recent[:consecutiveFailuresThreshold] {
		if res.Status == domain.StatusOK {
			return nil
		}
	}

	open, err := svc.store.ListByState(ctx, domain.IncidentStateOpen)
	if err != nil {
		return fmt.Errorf("incident: list open incidents: %w", err)
	}
	for _, inc := range open {
		if inc.TargetID == r.TargetID {
			return nil // already open
		}
	}

	ids := make([]string, consecutiveFailuresThreshold)
	for i, res := range recent[:consecutiveFailuresThreshold] {
		ids[i] = res.ID
	}

	now := time.Now().UTC()
	incident := domain.Incident{
		ID:        fmt.Sprintf("%d", now.UnixNano()),
		TargetID:  r.TargetID,
		State:     domain.IncidentStateOpen,
		FirstSeen: now,
		LastSeen:  now,
		ResultIDs: ids,
	}
	return svc.store.Save(ctx, incident)
}

func (svc *IncidentService) resolveIfOpen(ctx context.Context, r domain.Result) error {
	open, err := svc.store.ListByState(ctx, domain.IncidentStateOpen)
	if err != nil {
		return fmt.Errorf("incident: list open incidents: %w", err)
	}

	for _, inc := range open {
		if inc.TargetID != r.TargetID {
			continue
		}
		inc.State = domain.IncidentStateResolved
		inc.LastSeen = time.Now().UTC()
		inc.ResultIDs = append(inc.ResultIDs, r.ID)
		return svc.store.Update(ctx, inc)
	}
	return nil
}

func validTransition(from, to domain.IncidentState) bool {
	switch from {
	case domain.IncidentStateOpen:
		return to == domain.IncidentStateAcknowledged || to == domain.IncidentStateResolved
	case domain.IncidentStateAcknowledged:
		return to == domain.IncidentStateResolved
	default:
		return false
	}
}
