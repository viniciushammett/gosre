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
)

// TeamStore is the persistence contract required by TeamService.
type TeamStore interface {
	Save(ctx context.Context, t domain.Team) error
	Get(ctx context.Context, id string) (domain.Team, error)
	ListByOrg(ctx context.Context, orgID string) ([]domain.Team, error)
	Delete(ctx context.Context, id string) error
}

// ErrTeamNotFound is returned when a team does not exist.
var ErrTeamNotFound = errors.New("team not found")

// TeamService handles business logic for Team entities.
type TeamService struct {
	store TeamStore
}

// NewTeamService constructs a TeamService backed by the given store.
func NewTeamService(store TeamStore) *TeamService {
	return &TeamService{store: store}
}

// Create validates, generates ID and slug, and persists a new Team.
func (svc *TeamService) Create(ctx context.Context, t domain.Team) (domain.Team, error) {
	if t.Name == "" {
		return domain.Team{}, fmt.Errorf("name is required")
	}
	if t.OrganizationID == "" {
		return domain.Team{}, fmt.Errorf("organization_id is required")
	}
	if t.Slug == "" {
		t.Slug = slugify(t.Name)
	}
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	if err := svc.store.Save(ctx, t); err != nil {
		return domain.Team{}, fmt.Errorf("create team: %w", err)
	}
	return t, nil
}

// Get retrieves a Team by ID.
func (svc *TeamService) Get(ctx context.Context, id string) (domain.Team, error) {
	t, err := svc.store.Get(ctx, id)
	if err != nil {
		if err.Error() == ErrTeamNotFound.Error() {
			return domain.Team{}, ErrTeamNotFound
		}
		return domain.Team{}, fmt.Errorf("get team: %w", err)
	}
	return t, nil
}

// ListByOrg returns all teams for the given organization.
func (svc *TeamService) ListByOrg(ctx context.Context, orgID string) ([]domain.Team, error) {
	return svc.store.ListByOrg(ctx, orgID)
}

// Delete removes a Team by ID.
func (svc *TeamService) Delete(ctx context.Context, id string) error {
	if err := svc.store.Delete(ctx, id); err != nil {
		if err.Error() == ErrTeamNotFound.Error() {
			return ErrTeamNotFound
		}
		return fmt.Errorf("delete team: %w", err)
	}
	return nil
}
