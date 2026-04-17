// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/gosre/gosre-sdk/store"
)

// TargetService handles business logic for Target entities.
type TargetService struct {
	store store.TargetStore
}

// NewTargetService constructs a TargetService backed by the given store.
func NewTargetService(s store.TargetStore) *TargetService {
	return &TargetService{store: s}
}

// Create validates, assigns an ID if absent, and persists a new Target.
func (svc *TargetService) Create(ctx context.Context, t domain.Target) (domain.Target, error) {
	if t.Name == "" {
		return domain.Target{}, fmt.Errorf("name is required")
	}
	if t.Address == "" {
		return domain.Target{}, fmt.Errorf("address is required")
	}
	if t.ID == "" {
		t.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if err := svc.store.Save(ctx, t); err != nil {
		return domain.Target{}, fmt.Errorf("create target: %w", err)
	}
	return t, nil
}

// Get retrieves a Target by ID.
func (svc *TargetService) Get(ctx context.Context, id string) (domain.Target, error) {
	return svc.store.Get(ctx, id)
}

// List returns all stored targets.
func (svc *TargetService) List(ctx context.Context) ([]domain.Target, error) {
	return svc.store.List(ctx)
}

// Delete removes a Target by ID.
func (svc *TargetService) Delete(ctx context.Context, id string) error {
	return svc.store.Delete(ctx, id)
}
