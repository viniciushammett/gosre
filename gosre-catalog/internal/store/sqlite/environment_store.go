// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gosre/gosre-sdk/domain"
)

var errEnvironmentNotFound = errors.New("environment not found")

// EnvironmentStore implements store.EnvironmentStore for SQLite.
type EnvironmentStore struct {
	db *sql.DB
}

// EnvironmentStore returns an EnvironmentStore backed by the same database connection.
func (s *Store) EnvironmentStore() *EnvironmentStore {
	return &EnvironmentStore{db: s.db}
}

// Save inserts or replaces an Environment in the database.
func (s *EnvironmentStore) Save(ctx context.Context, e domain.Environment) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO environments
		 (id, name, project_id, kind, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		e.ID, e.Name, e.ProjectID, string(e.Kind),
		e.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save environment %q: %w", e.ID, err)
	}
	return nil
}

// Get retrieves an Environment by ID.
func (s *EnvironmentStore) Get(ctx context.Context, id string) (domain.Environment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, project_id, kind, created_at FROM environments WHERE id = ?`, id)
	return scanEnvironment(row)
}

// ListByProject returns all environments for the given project, ordered by kind.
func (s *EnvironmentStore) ListByProject(ctx context.Context, projectID string) ([]domain.Environment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, project_id, kind, created_at
		 FROM environments WHERE project_id = ? ORDER BY kind`, projectID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list environments for project %q: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectEnvironments(rows)
}

// Delete removes an Environment by ID.
func (s *EnvironmentStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM environments WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete environment %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
	}
	if n == 0 {
		return errEnvironmentNotFound
	}
	return nil
}

func collectEnvironments(rows *sql.Rows) ([]domain.Environment, error) {
	out := make([]domain.Environment, 0)
	for rows.Next() {
		e, err := scanEnvironment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate environments: %w", err)
	}
	return out, nil
}

func scanEnvironment(s scanner) (domain.Environment, error) {
	var (
		e         domain.Environment
		kind      string
		createdAt string
	)
	err := s.Scan(&e.ID, &e.Name, &e.ProjectID, &kind, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Environment{}, errEnvironmentNotFound
	}
	if err != nil {
		return domain.Environment{}, fmt.Errorf("sqlite: scan environment: %w", err)
	}
	e.Kind = domain.EnvironmentKind(kind)
	e.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Environment{}, fmt.Errorf("sqlite: parse created_at %q: %w", createdAt, err)
	}
	return e, nil
}
