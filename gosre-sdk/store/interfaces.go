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
}

// CheckStore defines persistence operations for CheckConfig entities.
type CheckStore interface {
	Save(ctx context.Context, c domain.CheckConfig) error
	Get(ctx context.Context, id string) (domain.CheckConfig, error)
	List(ctx context.Context) ([]domain.CheckConfig, error)
	Delete(ctx context.Context, id string) error
}

// IncidentStore defines persistence operations for Incident entities.
type IncidentStore interface {
	Save(ctx context.Context, i domain.Incident) error
	Get(ctx context.Context, id string) (domain.Incident, error)
	ListByState(ctx context.Context, state domain.IncidentState) ([]domain.Incident, error)
	Update(ctx context.Context, i domain.Incident) error
}
