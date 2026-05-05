// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package azuresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

var errProjectNotFound = errors.New("project not found")

// ProjectStore implements store.ProjectStore for Azure SQL.
type ProjectStore struct {
	db *sql.DB
}

// ProjectStore returns a ProjectStore backed by the same database connection.
func (s *Store) ProjectStore() *ProjectStore {
	return &ProjectStore{db: s.db}
}

// Save inserts or updates a Project using a MERGE statement.
func (s *ProjectStore) Save(ctx context.Context, p domain.Project) error {
	_, err := s.db.ExecContext(ctx, `
		MERGE projects WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5, @p6))
			AS S(id, organization_id, team_id, name, slug, created_at)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.organization_id=S.organization_id, T.team_id=S.team_id,
			           T.name=S.name, T.slug=S.slug
		WHEN NOT MATCHED THEN
			INSERT (id, organization_id, team_id, name, slug, created_at)
			VALUES (S.id, S.organization_id, S.team_id, S.name, S.slug, S.created_at);`,
		p.ID, p.OrganizationID, p.TeamID, p.Name, p.Slug, p.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save project %q: %w", p.ID, err)
	}
	return nil
}

// Get retrieves a Project by ID. Returns errProjectNotFound if not present.
func (s *ProjectStore) Get(ctx context.Context, id string) (domain.Project, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects WHERE id = @p1`, id)
	return scanProject(row)
}

// GetBySlug retrieves a Project by organization_id + slug. Returns errProjectNotFound if not present.
func (s *ProjectStore) GetBySlug(ctx context.Context, orgID, slug string) (domain.Project, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects
		 WHERE organization_id = @p1 AND slug = @p2`, orgID, slug)
	return scanProject(row)
}

// ListByOrg returns all projects for the given organization, ordered by name.
func (s *ProjectStore) ListByOrg(ctx context.Context, orgID string) ([]domain.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects
		 WHERE organization_id = @p1 ORDER BY name`, orgID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list projects for org %q: %w", orgID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectProjects(rows)
}

// ListByTeam returns all projects for the given team, ordered by name.
func (s *ProjectStore) ListByTeam(ctx context.Context, teamID string) ([]domain.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects
		 WHERE team_id = @p1 ORDER BY name`, teamID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list projects for team %q: %w", teamID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectProjects(rows)
}

// Delete removes a Project by ID. Returns errProjectNotFound if not present.
func (s *ProjectStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete project %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("azuresql: rows affected: %w", err)
	}
	if n == 0 {
		return errProjectNotFound
	}
	return nil
}

func collectProjects(rows *sql.Rows) ([]domain.Project, error) {
	out := make([]domain.Project, 0)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azuresql: iterate projects: %w", err)
	}
	return out, nil
}

func scanProject(s scanner) (domain.Project, error) {
	var p domain.Project
	err := s.Scan(&p.ID, &p.OrganizationID, &p.TeamID, &p.Name, &p.Slug, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Project{}, errProjectNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("azuresql: scan project: %w", err)
	}
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}
