// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"
	"github.com/viniciushammett/gosre/gosre-sdk/store"
)

// ErrEnvironmentNotFound is returned when an Environment does not exist.
var ErrEnvironmentNotFound = errors.New("environment not found")

// EnvironmentService handles business logic for Environment entities.
type EnvironmentService struct {
	store store.EnvironmentStore
}

// NewEnvironmentService constructs an EnvironmentService.
func NewEnvironmentService(s store.EnvironmentStore) *EnvironmentService {
	return &EnvironmentService{store: s}
}

// Create validates and persists a new Environment.
func (svc *EnvironmentService) Create(ctx context.Context, e domain.Environment) (domain.Environment, error) {
	if e.Name == "" {
		return domain.Environment{}, fmt.Errorf("name is required")
	}
	if e.ProjectID == "" {
		return domain.Environment{}, fmt.Errorf("project_id is required")
	}
	if e.Kind == "" {
		e.Kind = domain.EnvironmentKindDev
	}
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	e.CreatedAt = time.Now().UTC()
	if err := svc.store.Save(ctx, e); err != nil {
		return domain.Environment{}, fmt.Errorf("create environment: %w", err)
	}
	return e, nil
}

// Get retrieves an Environment by ID.
func (svc *EnvironmentService) Get(ctx context.Context, id string) (domain.Environment, error) {
	e, err := svc.store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrEnvironmentNotFound) || err.Error() == "environment not found" {
			return domain.Environment{}, ErrEnvironmentNotFound
		}
		return domain.Environment{}, fmt.Errorf("get environment %q: %w", id, err)
	}
	return e, nil
}

// ListByProject returns all environments scoped to a project.
func (svc *EnvironmentService) ListByProject(ctx context.Context, projectID string) ([]domain.Environment, error) {
	return svc.store.ListByProject(ctx, projectID)
}

// Delete removes an Environment by ID.
func (svc *EnvironmentService) Delete(ctx context.Context, id string) error {
	err := svc.store.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrEnvironmentNotFound) || err.Error() == "environment not found" {
			return ErrEnvironmentNotFound
		}
		return fmt.Errorf("delete environment %q: %w", id, err)
	}
	return nil
}
