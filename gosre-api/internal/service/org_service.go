// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosre/gosre-sdk/domain"
)

// OrgStore is the persistence contract required by OrgService.
type OrgStore interface {
	Save(ctx context.Context, o domain.Organization) error
	Get(ctx context.Context, id string) (domain.Organization, error)
	GetBySlug(ctx context.Context, slug string) (domain.Organization, error)
	List(ctx context.Context) ([]domain.Organization, error)
	Delete(ctx context.Context, id string) error
}

// ErrOrgNotFound is returned when an organization does not exist.
var ErrOrgNotFound = errors.New("organization not found")

// OrgService handles business logic for Organization entities.
type OrgService struct {
	store OrgStore
}

// NewOrgService constructs an OrgService backed by the given store.
func NewOrgService(store OrgStore) *OrgService {
	return &OrgService{store: store}
}

// Create validates, generates ID and slug, and persists a new Organization.
func (svc *OrgService) Create(ctx context.Context, o domain.Organization) (domain.Organization, error) {
	if o.Name == "" {
		return domain.Organization{}, fmt.Errorf("name is required")
	}
	if o.Slug == "" {
		o.Slug = slugify(o.Name)
	}
	o.ID = uuid.New().String()
	o.CreatedAt = time.Now().UTC()
	if err := svc.store.Save(ctx, o); err != nil {
		return domain.Organization{}, fmt.Errorf("create organization: %w", err)
	}
	return o, nil
}

// Get retrieves an Organization by ID.
func (svc *OrgService) Get(ctx context.Context, id string) (domain.Organization, error) {
	o, err := svc.store.Get(ctx, id)
	if err != nil {
		if err.Error() == ErrOrgNotFound.Error() {
			return domain.Organization{}, ErrOrgNotFound
		}
		return domain.Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return o, nil
}

// GetBySlug retrieves an Organization by slug.
func (svc *OrgService) GetBySlug(ctx context.Context, slug string) (domain.Organization, error) {
	o, err := svc.store.GetBySlug(ctx, slug)
	if err != nil {
		if err.Error() == ErrOrgNotFound.Error() {
			return domain.Organization{}, ErrOrgNotFound
		}
		return domain.Organization{}, fmt.Errorf("get organization by slug: %w", err)
	}
	return o, nil
}

// List returns all organizations.
func (svc *OrgService) List(ctx context.Context) ([]domain.Organization, error) {
	return svc.store.List(ctx)
}

// Delete removes an Organization by ID.
func (svc *OrgService) Delete(ctx context.Context, id string) error {
	if err := svc.store.Delete(ctx, id); err != nil {
		if err.Error() == ErrOrgNotFound.Error() {
			return ErrOrgNotFound
		}
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}

// slugify converts a display name to a URL-safe slug.
func slugify(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}
