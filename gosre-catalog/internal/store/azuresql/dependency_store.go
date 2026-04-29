// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package azuresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosre/gosre-sdk/domain"
)

var errDependencyNotFound = errors.New("dependency not found")

// DependencyStore implements store.DependencyStore for Azure SQL.
type DependencyStore struct {
	db *sql.DB
}

// DependencyStore returns a DependencyStore backed by the same database connection.
func (s *Store) DependencyStore() *DependencyStore {
	return &DependencyStore{db: s.db}
}

// Save inserts or updates a Dependency using a MERGE statement.
func (s *DependencyStore) Save(ctx context.Context, d domain.Dependency) error {
	_, err := s.db.ExecContext(ctx, `
		MERGE dependencies WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5))
			AS S(id, source_service_id, target_service_id, kind, created_at)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.source_service_id=S.source_service_id,
			           T.target_service_id=S.target_service_id, T.kind=S.kind
		WHEN NOT MATCHED THEN
			INSERT (id, source_service_id, target_service_id, kind, created_at)
			VALUES (S.id, S.source_service_id, S.target_service_id, S.kind, S.created_at);`,
		d.ID, d.SourceServiceID, d.TargetServiceID, string(d.Kind), d.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save dependency %q: %w", d.ID, err)
	}
	return nil
}

// Get retrieves a Dependency by ID.
func (s *DependencyStore) Get(ctx context.Context, id string) (domain.Dependency, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, source_service_id, target_service_id, kind, created_at
		 FROM dependencies WHERE id = @p1`, id)
	return scanDependency(row)
}

// ListBySource returns all dependencies where source_service_id matches.
func (s *DependencyStore) ListBySource(ctx context.Context, sourceServiceID string) ([]domain.Dependency, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source_service_id, target_service_id, kind, created_at
		 FROM dependencies WHERE source_service_id = @p1 ORDER BY created_at`, sourceServiceID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list dependencies by source %q: %w", sourceServiceID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectDependencies(rows)
}

// ListByTarget returns all dependencies where target_service_id matches.
func (s *DependencyStore) ListByTarget(ctx context.Context, targetServiceID string) ([]domain.Dependency, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source_service_id, target_service_id, kind, created_at
		 FROM dependencies WHERE target_service_id = @p1 ORDER BY created_at`, targetServiceID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list dependencies by target %q: %w", targetServiceID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectDependencies(rows)
}

// Delete removes a Dependency by ID.
func (s *DependencyStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM dependencies WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete dependency %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("azuresql: rows affected: %w", err)
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
		return nil, fmt.Errorf("azuresql: iterate dependencies: %w", err)
	}
	return out, nil
}

func scanDependency(s scanner) (domain.Dependency, error) {
	var (
		d    domain.Dependency
		kind string
	)
	err := s.Scan(&d.ID, &d.SourceServiceID, &d.TargetServiceID, &kind, &d.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Dependency{}, errDependencyNotFound
	}
	if err != nil {
		return domain.Dependency{}, fmt.Errorf("azuresql: scan dependency: %w", err)
	}
	d.Kind = domain.DependencyKind(kind)
	d.CreatedAt = d.CreatedAt.UTC()
	return d, nil
}
