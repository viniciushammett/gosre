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

var errOrgNotFound = errors.New("organization not found")

// OrgStore implements store.OrgStore for SQLite.
type OrgStore struct {
	db *sql.DB
}

// OrgStore returns an OrgStore backed by the same database connection.
func (s *Store) OrgStore() *OrgStore {
	return &OrgStore{db: s.db}
}

// Save inserts or replaces an Organization in the database.
func (s *OrgStore) Save(ctx context.Context, o domain.Organization) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO organizations (id, name, slug, created_at)
		 VALUES (?, ?, ?, ?)`,
		o.ID, o.Name, o.Slug, o.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save organization %q: %w", o.ID, err)
	}
	return nil
}

// Get retrieves an Organization by ID. Returns errOrgNotFound if not present.
func (s *OrgStore) Get(ctx context.Context, id string) (domain.Organization, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, slug, created_at FROM organizations WHERE id = ?`, id)
	return scanOrg(row)
}

// GetBySlug retrieves an Organization by slug. Returns errOrgNotFound if not present.
func (s *OrgStore) GetBySlug(ctx context.Context, slug string) (domain.Organization, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, slug, created_at FROM organizations WHERE slug = ?`, slug)
	return scanOrg(row)
}

// List returns all organizations ordered by name.
func (s *OrgStore) List(ctx context.Context) ([]domain.Organization, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, slug, created_at FROM organizations ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list organizations: %w", err)
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
		return nil, fmt.Errorf("sqlite: iterate organizations: %w", err)
	}
	return out, nil
}

// Delete removes an Organization by ID. Returns errOrgNotFound if not present.
func (s *OrgStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM organizations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete organization %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
	}
	if n == 0 {
		return errOrgNotFound
	}
	return nil
}

func scanOrg(s scanner) (domain.Organization, error) {
	var (
		o         domain.Organization
		createdAt string
	)
	err := s.Scan(&o.ID, &o.Name, &o.Slug, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Organization{}, errOrgNotFound
	}
	if err != nil {
		return domain.Organization{}, fmt.Errorf("sqlite: scan organization: %w", err)
	}
	o.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("sqlite: parse created_at %q: %w", createdAt, err)
	}
	return o, nil
}
