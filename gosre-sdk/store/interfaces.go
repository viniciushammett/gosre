// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package store

import (
	"context"

	"github.com/gosre/gosre-sdk/domain"
)

// TargetStore defines persistence operations for Target entities.
type TargetStore interface {
	Save(ctx context.Context, t domain.Target) error
	Get(ctx context.Context, id string) (domain.Target, error)
	List(ctx context.Context) ([]domain.Target, error)
	Delete(ctx context.Context, id string) error
}

// ResultStore defines persistence operations for Result entities.
type ResultStore interface {
	Save(ctx context.Context, r domain.Result) error
	Get(ctx context.Context, id string) (domain.Result, error)
	List(ctx context.Context) ([]domain.Result, error)
	ListByTarget(ctx context.Context, targetID string) ([]domain.Result, error)
	DeleteByTargetID(ctx context.Context, targetID string) error
}

// CheckStore defines persistence operations for CheckConfig entities.
type CheckStore interface {
	Save(ctx context.Context, c domain.CheckConfig) error
	Get(ctx context.Context, id string) (domain.CheckConfig, error)
	List(ctx context.Context) ([]domain.CheckConfig, error)
	Delete(ctx context.Context, id string) error
	DeleteByTargetID(ctx context.Context, targetID string) error
}

// IncidentStore defines persistence operations for Incident entities.
type IncidentStore interface {
	Save(ctx context.Context, i domain.Incident) error
	Get(ctx context.Context, id string) (domain.Incident, error)
	ListByState(ctx context.Context, state domain.IncidentState) ([]domain.Incident, error)
	Update(ctx context.Context, i domain.Incident) error
	DeleteByTargetID(ctx context.Context, targetID string) error
}

// OrgStore defines persistence operations for Organization entities.
type OrgStore interface {
	Save(ctx context.Context, o domain.Organization) error
	Get(ctx context.Context, id string) (domain.Organization, error)
	GetBySlug(ctx context.Context, slug string) (domain.Organization, error)
	List(ctx context.Context) ([]domain.Organization, error)
	Delete(ctx context.Context, id string) error
}

// TeamStore defines persistence operations for Team entities.
type TeamStore interface {
	Save(ctx context.Context, t domain.Team) error
	Get(ctx context.Context, id string) (domain.Team, error)
	ListByOrg(ctx context.Context, orgID string) ([]domain.Team, error)
	Delete(ctx context.Context, id string) error
}

// ProjectStore defines persistence operations for Project entities.
type ProjectStore interface {
	Save(ctx context.Context, p domain.Project) error
	Get(ctx context.Context, id string) (domain.Project, error)
	GetBySlug(ctx context.Context, orgID, slug string) (domain.Project, error)
	ListByOrg(ctx context.Context, orgID string) ([]domain.Project, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.Project, error)
	Delete(ctx context.Context, id string) error
}
