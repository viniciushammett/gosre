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

var errDependencyNotFound = errors.New("dependency not found")

// DependencyStore implements store.DependencyStore for SQLite.
type DependencyStore struct {
	db *sql.DB
}

// DependencyStore returns a DependencyStore backed by the same database connection.
func (s *Store) DependencyStore() *DependencyStore {
	return &DependencyStore{db: s.db}
}

// Save inserts or replaces a Dependency in the database.
func (s *DependencyStore) Save(ctx context.Context, d domain.Dependency) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO dependencies
		 (id, source_service_id, target_service_id, kind, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		d.ID, d.SourceServiceID, d.TargetServiceID, string(d.Kind),
		d.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save dependency %q: %w", d.ID, err)
	}
	return nil
}

// Get retrieves a Dependency by ID.
func (s *DependencyStore) Get(ctx context.Context, id string) (domain.Dependency, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, source_service_id, target_service_id, kind, created_at
		 FROM dependencies WHERE id = ?`, id)
	return scanDependency(row)
}

// ListBySource returns all dependencies where source_service_id matches.
func (s *DependencyStore) ListBySource(ctx context.Context, sourceServiceID string) ([]domain.Dependency, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source_service_id, target_service_id, kind, created_at
		 FROM dependencies WHERE source_service_id = ? ORDER BY created_at`, sourceServiceID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list dependencies by source %q: %w", sourceServiceID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectDependencies(rows)
}

// ListByTarget returns all dependencies where target_service_id matches.
func (s *DependencyStore) ListByTarget(ctx context.Context, targetServiceID string) ([]domain.Dependency, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source_service_id, target_service_id, kind, created_at
		 FROM dependencies WHERE target_service_id = ? ORDER BY created_at`, targetServiceID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list dependencies by target %q: %w", targetServiceID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectDependencies(rows)
}

// Delete removes a Dependency by ID.
func (s *DependencyStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM dependencies WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete dependency %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
	}
	if n == 0 {
		return errDependencyNotFound
	}
	return nil
}

func collectDependencies(rows *sql.Rows) ([]domain.Dependency, error) {
	out := make([]domain.Dependency, 0)
	for rows.Next() {
		d, err := scanDependency(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate dependencies: %w", err)
	}
	return out, nil
}

func scanDependency(s scanner) (domain.Dependency, error) {
	var (
		d         domain.Dependency
		kind      string
		createdAt string
	)
	err := s.Scan(&d.ID, &d.SourceServiceID, &d.TargetServiceID, &kind, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Dependency{}, errDependencyNotFound
	}
	if err != nil {
		return domain.Dependency{}, fmt.Errorf("sqlite: scan dependency: %w", err)
	}
	d.Kind = domain.DependencyKind(kind)
	d.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Dependency{}, fmt.Errorf("sqlite: parse created_at %q: %w", createdAt, err)
	}
	return d, nil
}
