// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

var errServiceNotFound = errors.New("service not found")

// ServiceStore implements store.ServiceStore for SQLite.
type ServiceStore struct {
	db *sql.DB
}

// ServiceStore returns a ServiceStore backed by the same database connection.
func (s *Store) ServiceStore() *ServiceStore {
	return &ServiceStore{db: s.db}
}

// Save inserts or replaces a Service in the database.
func (s *ServiceStore) Save(ctx context.Context, svc domain.Service) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO services
		 (id, name, owner, criticality, runbook_url, repo_url, project_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		svc.ID, svc.Name, svc.Owner, string(svc.Criticality),
		svc.RunbookURL, svc.RepoURL, svc.ProjectID,
		svc.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save service %q: %w", svc.ID, err)
	}
	return nil
}

// Get retrieves a Service by ID.
func (s *ServiceStore) Get(ctx context.Context, id string) (domain.Service, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, owner, criticality, runbook_url, repo_url, project_id, created_at
		 FROM services WHERE id = ?`, id)
	return scanService(row)
}

// List returns all services ordered by name.
func (s *ServiceStore) List(ctx context.Context) ([]domain.Service, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, owner, criticality, runbook_url, repo_url, project_id, created_at
		 FROM services ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list services: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return collectServices(rows)
}

// ListByProject returns all services for the given project, ordered by name.
func (s *ServiceStore) ListByProject(ctx context.Context, projectID string) ([]domain.Service, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, owner, criticality, runbook_url, repo_url, project_id, created_at
		 FROM services WHERE project_id = ? ORDER BY name`, projectID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list services for project %q: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectServices(rows)
}

// Delete removes a Service by ID.
func (s *ServiceStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM services WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete service %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
	}
	if n == 0 {
		return errServiceNotFound
	}
	return nil
}

func collectServices(rows *sql.Rows) ([]domain.Service, error) {
	out := make([]domain.Service, 0)
	for rows.Next() {
		svc, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, svc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate services: %w", err)
	}
	return out, nil
}

func scanService(s scanner) (domain.Service, error) {
	var (
		svc         domain.Service
		criticality string
		createdAt   string
	)
	err := s.Scan(&svc.ID, &svc.Name, &svc.Owner, &criticality,
		&svc.RunbookURL, &svc.RepoURL, &svc.ProjectID, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Service{}, errServiceNotFound
	}
	if err != nil {
		return domain.Service{}, fmt.Errorf("sqlite: scan service: %w", err)
	}
	svc.Criticality = domain.ServiceCriticality(criticality)
	svc.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Service{}, fmt.Errorf("sqlite: parse created_at %q: %w", createdAt, err)
	}
	return svc, nil
}
