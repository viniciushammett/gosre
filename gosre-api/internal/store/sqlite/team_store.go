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

var errTeamNotFound = errors.New("team not found")

// TeamStore implements store.TeamStore for SQLite.
type TeamStore struct {
	db *sql.DB
}

// TeamStore returns a TeamStore backed by the same database connection.
func (s *Store) TeamStore() *TeamStore {
	return &TeamStore{db: s.db}
}

// Save inserts or replaces a Team in the database.
func (s *TeamStore) Save(ctx context.Context, t domain.Team) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO teams (id, organization_id, name, slug, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		t.ID, t.OrganizationID, t.Name, t.Slug, t.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save team %q: %w", t.ID, err)
	}
	return nil
}

// Get retrieves a Team by ID. Returns errTeamNotFound if not present.
func (s *TeamStore) Get(ctx context.Context, id string) (domain.Team, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, name, slug, created_at FROM teams WHERE id = ?`, id)
	return scanTeam(row)
}

// ListByOrg returns all teams for the given organization, ordered by name.
func (s *TeamStore) ListByOrg(ctx context.Context, orgID string) ([]domain.Team, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, organization_id, name, slug, created_at FROM teams
		 WHERE organization_id = ? ORDER BY name`, orgID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list teams for org %q: %w", orgID, err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.Team, 0)
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate teams: %w", err)
	}
	return out, nil
}

// Delete removes a Team by ID. Returns errTeamNotFound if not present.
func (s *TeamStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM teams WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete team %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
	}
	if n == 0 {
		return errTeamNotFound
	}
	return nil
}

func scanTeam(s scanner) (domain.Team, error) {
	var (
		t         domain.Team
		createdAt string
	)
	err := s.Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Slug, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Team{}, errTeamNotFound
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("sqlite: scan team: %w", err)
	}
	t.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Team{}, fmt.Errorf("sqlite: parse created_at %q: %w", createdAt, err)
	}
	return t, nil
}
