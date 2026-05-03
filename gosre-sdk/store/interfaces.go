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

// ServiceStore defines persistence operations for Service catalog entries.
type ServiceStore interface {
	Save(ctx context.Context, s domain.Service) error
	Get(ctx context.Context, id string) (domain.Service, error)
	List(ctx context.Context) ([]domain.Service, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.Service, error)
	Delete(ctx context.Context, id string) error
}

// DependencyStore defines persistence operations for inter-service Dependency edges.
type DependencyStore interface {
	Save(ctx context.Context, d domain.Dependency) error
	Get(ctx context.Context, id string) (domain.Dependency, error)
	ListBySource(ctx context.Context, sourceServiceID string) ([]domain.Dependency, error)
	ListByTarget(ctx context.Context, targetServiceID string) ([]domain.Dependency, error)
	Delete(ctx context.Context, id string) error
}

// EnvironmentStore defines persistence operations for Environment entities.
type EnvironmentStore interface {
	Save(ctx context.Context, e domain.Environment) error
	Get(ctx context.Context, id string) (domain.Environment, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.Environment, error)
	Delete(ctx context.Context, id string) error
}

// AssignmentStore defines persistence operations for check-to-agent Assignment records.
// Implementations are expected to use Redis — not SQL — for low-latency reads by the scheduler.
type AssignmentStore interface {
	Save(ctx context.Context, a domain.Assignment) error
	Get(ctx context.Context, id string) (domain.Assignment, error)
	ListByAgent(ctx context.Context, agentID string) ([]domain.Assignment, error)
	DeleteByAgent(ctx context.Context, agentID string) error
}

// SLOStore defines persistence operations for SLO entities.
type SLOStore interface {
	Save(ctx context.Context, s domain.SLO) error
	Get(ctx context.Context, id string) (domain.SLO, error)
	ListByTarget(ctx context.Context, targetID string) ([]domain.SLO, error)
	Delete(ctx context.Context, id string) error
}

// NotificationChannelStore defines persistence operations for NotificationChannel entities.
type NotificationChannelStore interface {
	Save(ctx context.Context, c domain.NotificationChannel) error
	Get(ctx context.Context, id string) (domain.NotificationChannel, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.NotificationChannel, error)
	Delete(ctx context.Context, id string) error
}

// NotificationRuleStore defines persistence operations for NotificationRule entities.
type NotificationRuleStore interface {
	Save(ctx context.Context, r domain.NotificationRule) error
	Get(ctx context.Context, id string) (domain.NotificationRule, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.NotificationRule, error)
	Delete(ctx context.Context, id string) error
}
