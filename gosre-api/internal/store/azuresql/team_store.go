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

var errTeamNotFound = errors.New("team not found")

// TeamStore implements store.TeamStore for Azure SQL.
type TeamStore struct {
	db *sql.DB
}

// TeamStore returns a TeamStore backed by the same database connection.
func (s *Store) TeamStore() *TeamStore {
	return &TeamStore{db: s.db}
}

// Save inserts or updates a Team using a MERGE statement.
func (s *TeamStore) Save(ctx context.Context, t domain.Team) error {
	_, err := s.db.ExecContext(ctx, `
		MERGE teams WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5))
			AS S(id, organization_id, name, slug, created_at)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.organization_id=S.organization_id, T.name=S.name, T.slug=S.slug
		WHEN NOT MATCHED THEN
			INSERT (id, organization_id, name, slug, created_at)
			VALUES (S.id, S.organization_id, S.name, S.slug, S.created_at);`,
		t.ID, t.OrganizationID, t.Name, t.Slug, t.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save team %q: %w", t.ID, err)
	}
	return nil
}

// Get retrieves a Team by ID. Returns errTeamNotFound if not present.
func (s *TeamStore) Get(ctx context.Context, id string) (domain.Team, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, name, slug, created_at FROM teams WHERE id = @p1`, id)
	return scanTeam(row)
}

// ListByOrg returns all teams for the given organization, ordered by name.
func (s *TeamStore) ListByOrg(ctx context.Context, orgID string) ([]domain.Team, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, organization_id, name, slug, created_at FROM teams
		 WHERE organization_id = @p1 ORDER BY name`, orgID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list teams for org %q: %w", orgID, err)
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
		return nil, fmt.Errorf("azuresql: iterate teams: %w", err)
	}
	return out, nil
}

// Delete removes a Team by ID. Returns errTeamNotFound if not present.
func (s *TeamStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM teams WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete team %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("azuresql: rows affected: %w", err)
	}
	if n == 0 {
		return errTeamNotFound
	}
	return nil
}

func scanTeam(s scanner) (domain.Team, error) {
	var t domain.Team
	err := s.Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Slug, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Team{}, errTeamNotFound
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("azuresql: scan team: %w", err)
	}
	t.CreatedAt = t.CreatedAt.UTC()
	return t, nil
}
