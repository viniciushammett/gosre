// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gosre/gosre-sdk/domain"
	"github.com/gosre/gosre-sdk/store"
)

// ErrDependencyNotFound is returned when a Dependency does not exist.
var ErrDependencyNotFound = errors.New("dependency not found")

// DependencyService handles business logic for inter-service Dependency edges.
type DependencyService struct {
	store store.DependencyStore
}

// NewDependencyService constructs a DependencyService.
func NewDependencyService(s store.DependencyStore) *DependencyService {
	return &DependencyService{store: s}
}

// Create validates and persists a new Dependency.
func (svc *DependencyService) Create(ctx context.Context, d domain.Dependency) (domain.Dependency, error) {
	if d.SourceServiceID == "" {
		return domain.Dependency{}, fmt.Errorf("source_service_id is required")
	}
	if d.TargetServiceID == "" {
		return domain.Dependency{}, fmt.Errorf("target_service_id is required")
	}
	if d.SourceServiceID == d.TargetServiceID {
		return domain.Dependency{}, fmt.Errorf("source and target service must differ")
	}
	if d.Kind == "" {
		d.Kind = domain.DependencyKindGeneric
	}
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	d.CreatedAt = time.Now().UTC()
	if err := svc.store.Save(ctx, d); err != nil {
		return domain.Dependency{}, fmt.Errorf("create dependency: %w", err)
	}
	return d, nil
}

// Get retrieves a Dependency by ID.
func (svc *DependencyService) Get(ctx context.Context, id string) (domain.Dependency, error) {
	d, err := svc.store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDependencyNotFound) || err.Error() == "dependency not found" {
			return domain.Dependency{}, ErrDependencyNotFound
		}
		return domain.Dependency{}, fmt.Errorf("get dependency %q: %w", id, err)
	}
	return d, nil
}

// ListBySource returns all outgoing dependencies from a service.
func (svc *DependencyService) ListBySource(ctx context.Context, sourceServiceID string) ([]domain.Dependency, error) {
	return svc.store.ListBySource(ctx, sourceServiceID)
}

// ListByTarget returns all incoming dependencies to a service.
func (svc *DependencyService) ListByTarget(ctx context.Context, targetServiceID string) ([]domain.Dependency, error) {
	return svc.store.ListByTarget(ctx, targetServiceID)
}

// Delete removes a Dependency by ID.
func (svc *DependencyService) Delete(ctx context.Context, id string) error {
	err := svc.store.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDependencyNotFound) || err.Error() == "dependency not found" {
			return ErrDependencyNotFound
		}
		return fmt.Errorf("delete dependency %q: %w", id, err)
	}
	return nil
}
