// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/gosre/gosre-sdk/store"
)

const defaultCheckTimeout = 10 * time.Second

// CheckService handles business logic for CheckConfig entities.
type CheckService struct {
	store     store.CheckStore
	targets   store.TargetStore
	results   *ResultService
	incidents *IncidentService
	checkers  map[domain.CheckType]domain.Checker
}

// NewCheckService constructs a CheckService with its runtime dependencies.
func NewCheckService(
	s store.CheckStore,
	targets store.TargetStore,
	results *ResultService,
	incidents *IncidentService,
	checkers map[domain.CheckType]domain.Checker,
) *CheckService {
	return &CheckService{
		store:     s,
		targets:   targets,
		results:   results,
		incidents: incidents,
		checkers:  checkers,
	}
}

// Create validates, assigns an ID if absent, and persists a new CheckConfig.
func (svc *CheckService) Create(ctx context.Context, c domain.CheckConfig) (domain.CheckConfig, error) {
	if c.TargetID == "" {
		return domain.CheckConfig{}, fmt.Errorf("target_id is required")
	}
	if c.Type == "" {
		return domain.CheckConfig{}, fmt.Errorf("type is required")
	}
	if c.ID == "" {
		c.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if err := svc.store.Save(ctx, c); err != nil {
		return domain.CheckConfig{}, fmt.Errorf("create check: %w", err)
	}
	return c, nil
}

// Get retrieves a CheckConfig by ID.
func (svc *CheckService) Get(ctx context.Context, id string) (domain.CheckConfig, error) {
	return svc.store.Get(ctx, id)
}

// List returns all stored CheckConfigs.
func (svc *CheckService) List(ctx context.Context) ([]domain.CheckConfig, error) {
	return svc.store.List(ctx)
}

// Delete removes a CheckConfig by ID.
func (svc *CheckService) Delete(ctx context.Context, id string) error {
	return svc.store.Delete(ctx, id)
}

// Run executes the check identified by id immediately and persists the Result.
func (svc *CheckService) Run(ctx context.Context, id string) (domain.Result, error) {
	cfg, err := svc.store.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Result{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.Result{}, fmt.Errorf("get check %q: %w", id, err)
	}

	t, err := svc.targets.Get(ctx, cfg.TargetID)
	if errors.Is(err, domain.ErrTargetNotFound) {
		return domain.Result{}, domain.ErrTargetNotFound
	}
	if err != nil {
		return domain.Result{}, fmt.Errorf("get target %q: %w", cfg.TargetID, err)
	}

	checker, ok := svc.checkers[cfg.Type]
	if !ok {
		return domain.Result{}, fmt.Errorf("unsupported check type %q", cfg.Type)
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = defaultCheckTimeout
	}

	r := checker.Execute(ctx, t, cfg)

	saved, err := svc.results.Save(ctx, r)
	if err != nil {
		return domain.Result{}, fmt.Errorf("save result: %w", err)
	}

	if err := svc.incidents.Process(ctx, saved); err != nil {
		return domain.Result{}, fmt.Errorf("process incident: %w", err)
	}
	return saved, nil
}
