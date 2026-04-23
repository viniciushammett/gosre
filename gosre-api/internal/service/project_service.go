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
)

// ProjectStore is the persistence contract required by ProjectService.
type ProjectStore interface {
	Save(ctx context.Context, p domain.Project) error
	Get(ctx context.Context, id string) (domain.Project, error)
	GetBySlug(ctx context.Context, orgID, slug string) (domain.Project, error)
	ListByOrg(ctx context.Context, orgID string) ([]domain.Project, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.Project, error)
	Delete(ctx context.Context, id string) error
}

// ErrProjectNotFound is returned when a project does not exist.
var ErrProjectNotFound = errors.New("project not found")

// ProjectService handles business logic for Project entities.
type ProjectService struct {
	store ProjectStore
}

// NewProjectService constructs a ProjectService backed by the given store.
func NewProjectService(store ProjectStore) *ProjectService {
	return &ProjectService{store: store}
}

// Create validates, generates ID and slug, and persists a new Project.
func (svc *ProjectService) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	if p.Name == "" {
		return domain.Project{}, fmt.Errorf("name is required")
	}
	if p.OrganizationID == "" {
		return domain.Project{}, fmt.Errorf("organization_id is required")
	}
	if p.Slug == "" {
		p.Slug = slugify(p.Name)
	}
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now().UTC()
	if err := svc.store.Save(ctx, p); err != nil {
		return domain.Project{}, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

// Get retrieves a Project by ID.
func (svc *ProjectService) Get(ctx context.Context, id string) (domain.Project, error) {
	p, err := svc.store.Get(ctx, id)
	if err != nil {
		if err.Error() == ErrProjectNotFound.Error() {
			return domain.Project{}, ErrProjectNotFound
		}
		return domain.Project{}, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

// GetBySlug retrieves a Project by organization_id and slug.
func (svc *ProjectService) GetBySlug(ctx context.Context, orgID, slug string) (domain.Project, error) {
	p, err := svc.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		if err.Error() == ErrProjectNotFound.Error() {
			return domain.Project{}, ErrProjectNotFound
		}
		return domain.Project{}, fmt.Errorf("get project by slug: %w", err)
	}
	return p, nil
}

// ListByOrg returns all projects for the given organization.
func (svc *ProjectService) ListByOrg(ctx context.Context, orgID string) ([]domain.Project, error) {
	return svc.store.ListByOrg(ctx, orgID)
}

// ListByTeam returns all projects for the given team.
func (svc *ProjectService) ListByTeam(ctx context.Context, teamID string) ([]domain.Project, error) {
	return svc.store.ListByTeam(ctx, teamID)
}

// Delete removes a Project by ID.
func (svc *ProjectService) Delete(ctx context.Context, id string) error {
	if err := svc.store.Delete(ctx, id); err != nil {
		if err.Error() == ErrProjectNotFound.Error() {
			return ErrProjectNotFound
		}
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}
