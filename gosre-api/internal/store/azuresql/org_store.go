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

var errOrgNotFound = errors.New("organization not found")

// OrgStore implements store.OrgStore for Azure SQL.
type OrgStore struct {
	db *sql.DB
}

// OrgStore returns an OrgStore backed by the same database connection.
func (s *Store) OrgStore() *OrgStore {
	return &OrgStore{db: s.db}
}

// Save inserts or updates an Organization using a MERGE statement.
func (s *OrgStore) Save(ctx context.Context, o domain.Organization) error {
	_, err := s.db.ExecContext(ctx, `
		MERGE organizations WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4))
			AS S(id, name, slug, created_at)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.name=S.name, T.slug=S.slug
		WHEN NOT MATCHED THEN
			INSERT (id, name, slug, created_at)
			VALUES (S.id, S.name, S.slug, S.created_at);`,
		o.ID, o.Name, o.Slug, o.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save organization %q: %w", o.ID, err)
	}
	return nil
}

// Get retrieves an Organization by ID. Returns errOrgNotFound if not present.
func (s *OrgStore) Get(ctx context.Context, id string) (domain.Organization, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, slug, created_at FROM organizations WHERE id = @p1`, id)
	return scanOrg(row)
}

// GetBySlug retrieves an Organization by slug. Returns errOrgNotFound if not present.
func (s *OrgStore) GetBySlug(ctx context.Context, slug string) (domain.Organization, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, slug, created_at FROM organizations WHERE slug = @p1`, slug)
	return scanOrg(row)
}

// List returns all organizations ordered by name.
func (s *OrgStore) List(ctx context.Context) ([]domain.Organization, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, slug, created_at FROM organizations ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list organizations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.Organization, 0)
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azuresql: iterate organizations: %w", err)
	}
	return out, nil
}

// Delete removes an Organization by ID. Returns errOrgNotFound if not present.
func (s *OrgStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM organizations WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete organization %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("azuresql: rows affected: %w", err)
	}
	if n == 0 {
		return errOrgNotFound
	}
	return nil
}

func scanOrg(s scanner) (domain.Organization, error) {
	var o domain.Organization
	err := s.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Organization{}, errOrgNotFound
	}
	if err != nil {
		return domain.Organization{}, fmt.Errorf("azuresql: scan organization: %w", err)
	}
	o.CreatedAt = o.CreatedAt.UTC()
	return o, nil
}
