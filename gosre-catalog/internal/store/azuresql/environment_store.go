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

var errEnvironmentNotFound = errors.New("environment not found")

// EnvironmentStore implements store.EnvironmentStore for Azure SQL.
type EnvironmentStore struct {
	db *sql.DB
}

// EnvironmentStore returns an EnvironmentStore backed by the same database connection.
func (s *Store) EnvironmentStore() *EnvironmentStore {
	return &EnvironmentStore{db: s.db}
}

// Save inserts or updates an Environment using a MERGE statement.
func (s *EnvironmentStore) Save(ctx context.Context, e domain.Environment) error {
	_, err := s.db.ExecContext(ctx, `
		MERGE environments WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5))
			AS S(id, name, project_id, kind, created_at)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.name=S.name, T.project_id=S.project_id, T.kind=S.kind
		WHEN NOT MATCHED THEN
			INSERT (id, name, project_id, kind, created_at)
			VALUES (S.id, S.name, S.project_id, S.kind, S.created_at);`,
		e.ID, e.Name, e.ProjectID, string(e.Kind), e.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save environment %q: %w", e.ID, err)
	}
	return nil
}

// Get retrieves an Environment by ID.
func (s *EnvironmentStore) Get(ctx context.Context, id string) (domain.Environment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, project_id, kind, created_at FROM environments WHERE id = @p1`, id)
	return scanEnvironment(row)
}

// ListByProject returns all environments for the given project, ordered by kind.
func (s *EnvironmentStore) ListByProject(ctx context.Context, projectID string) ([]domain.Environment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, project_id, kind, created_at
		 FROM environments WHERE project_id = @p1 ORDER BY kind`, projectID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list environments for project %q: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectEnvironments(rows)
}

// Delete removes an Environment by ID.
func (s *EnvironmentStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM environments WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete environment %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("azuresql: rows affected: %w", err)
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
		return nil, fmt.Errorf("azuresql: iterate environments: %w", err)
	}
	return out, nil
}

func scanEnvironment(s scanner) (domain.Environment, error) {
	var (
		e    domain.Environment
		kind string
	)
	err := s.Scan(&e.ID, &e.Name, &e.ProjectID, &kind, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Environment{}, errEnvironmentNotFound
	}
	if err != nil {
		return domain.Environment{}, fmt.Errorf("azuresql: scan environment: %w", err)
	}
	e.Kind = domain.EnvironmentKind(kind)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}
