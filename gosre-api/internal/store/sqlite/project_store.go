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

var errProjectNotFound = errors.New("project not found")

// ProjectStore implements store.ProjectStore for SQLite.
type ProjectStore struct {
	db *sql.DB
}

// ProjectStore returns a ProjectStore backed by the same database connection.
func (s *Store) ProjectStore() *ProjectStore {
	return &ProjectStore{db: s.db}
}

// Save inserts or replaces a Project in the database.
func (s *ProjectStore) Save(ctx context.Context, p domain.Project) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO projects (id, organization_id, team_id, name, slug, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.OrganizationID, p.TeamID, p.Name, p.Slug, p.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save project %q: %w", p.ID, err)
	}
	return nil
}

// Get retrieves a Project by ID. Returns errProjectNotFound if not present.
func (s *ProjectStore) Get(ctx context.Context, id string) (domain.Project, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects WHERE id = ?`, id)
	return scanProject(row)
}

// GetBySlug retrieves a Project by organization_id + slug. Returns errProjectNotFound if not present.
func (s *ProjectStore) GetBySlug(ctx context.Context, orgID, slug string) (domain.Project, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects
		 WHERE organization_id = ? AND slug = ?`, orgID, slug)
	return scanProject(row)
}

// ListByOrg returns all projects for the given organization, ordered by name.
func (s *ProjectStore) ListByOrg(ctx context.Context, orgID string) ([]domain.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects
		 WHERE organization_id = ? ORDER BY name`, orgID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list projects for org %q: %w", orgID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectProjects(rows)
}

// ListByTeam returns all projects for the given team, ordered by name.
func (s *ProjectStore) ListByTeam(ctx context.Context, teamID string) ([]domain.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, organization_id, team_id, name, slug, created_at FROM projects
		 WHERE team_id = ? ORDER BY name`, teamID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list projects for team %q: %w", teamID, err)
	}
	defer func() { _ = rows.Close() }()
	return collectProjects(rows)
}

// Delete removes a Project by ID. Returns errProjectNotFound if not present.
func (s *ProjectStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete project %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
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
		return nil, fmt.Errorf("sqlite: iterate projects: %w", err)
	}
	return out, nil
}

func scanProject(s scanner) (domain.Project, error) {
	var (
		p         domain.Project
		createdAt string
	)
	err := s.Scan(&p.ID, &p.OrganizationID, &p.TeamID, &p.Name, &p.Slug, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Project{}, errProjectNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("sqlite: scan project: %w", err)
	}
	p.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("sqlite: parse created_at %q: %w", createdAt, err)
	}
	return p, nil
}
