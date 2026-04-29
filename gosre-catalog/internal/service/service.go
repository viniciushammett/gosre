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

// ErrServiceNotFound is returned when a Service does not exist.
var ErrServiceNotFound = errors.New("service not found")

// CatalogService handles business logic for Service catalog entries.
type CatalogService struct {
	store store.ServiceStore
}

// NewCatalogService constructs a CatalogService.
func NewCatalogService(s store.ServiceStore) *CatalogService {
	return &CatalogService{store: s}
}

// Create validates and persists a new Service.
func (svc *CatalogService) Create(ctx context.Context, s domain.Service) (domain.Service, error) {
	if s.Name == "" {
		return domain.Service{}, fmt.Errorf("name is required")
	}
	if s.Owner == "" {
		return domain.Service{}, fmt.Errorf("owner is required")
	}
	if s.Criticality == "" {
		s.Criticality = domain.CriticalityMedium
	}
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	s.CreatedAt = time.Now().UTC()
	if err := svc.store.Save(ctx, s); err != nil {
		return domain.Service{}, fmt.Errorf("create service: %w", err)
	}
	return s, nil
}

// Get retrieves a Service by ID.
func (svc *CatalogService) Get(ctx context.Context, id string) (domain.Service, error) {
	s, err := svc.store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrServiceNotFound) || err.Error() == "service not found" {
			return domain.Service{}, ErrServiceNotFound
		}
		return domain.Service{}, fmt.Errorf("get service %q: %w", id, err)
	}
	return s, nil
}

// List returns all services.
func (svc *CatalogService) List(ctx context.Context) ([]domain.Service, error) {
	return svc.store.List(ctx)
}

// ListByProject returns services scoped to a project.
func (svc *CatalogService) ListByProject(ctx context.Context, projectID string) ([]domain.Service, error) {
	return svc.store.ListByProject(ctx, projectID)
}

// Delete removes a Service by ID.
func (svc *CatalogService) Delete(ctx context.Context, id string) error {
	err := svc.store.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrServiceNotFound) || err.Error() == "service not found" {
			return ErrServiceNotFound
		}
		return fmt.Errorf("delete service %q: %w", id, err)
	}
	return nil
}
