// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gosre/gosre-sdk/domain"
)

// scanner abstracts *sql.Row and *sql.Rows to share scan helpers.
type scanner interface {
	Scan(dest ...any) error
}

// Save inserts or updates a Target in the database.
func (s *Store) Save(ctx context.Context, t domain.Target) error {
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("postgres: marshal tags: %w", err)
	}

	meta, err := json.Marshal(t.Metadata)
	if err != nil {
		return fmt.Errorf("postgres: marshal metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO targets (id, name, type, address, tags, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE
		 SET name=$2, type=$3, address=$4, tags=$5, metadata=$6`,
		t.ID, t.Name, string(t.Type), t.Address, string(tags), string(meta),
	)
	if err != nil {
		return fmt.Errorf("postgres: save target %q: %w", t.ID, err)
	}
	return nil
}

// Get retrieves a Target by ID. Returns domain.ErrTargetNotFound if not present.
func (s *Store) Get(ctx context.Context, id string) (domain.Target, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, address, tags, metadata FROM targets WHERE id = $1`, id)
	return scanTarget(row)
}

// List returns all targets. Returns an empty (non-nil) slice when none exist.
func (s *Store) List(ctx context.Context) ([]domain.Target, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, address, tags, metadata FROM targets`)
	if err != nil {
		return nil, fmt.Errorf("postgres: list targets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	targets := make([]domain.Target, 0)
	for rows.Next() {
		t, err := scanTarget(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate targets: %w", err)
	}
	return targets, nil
}

// Delete removes a Target by ID. Returns domain.ErrTargetNotFound if not present.
func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM targets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete target %q: %w", id, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrTargetNotFound
	}
	return nil
}

func scanTarget(s scanner) (domain.Target, error) {
	var (
		t        domain.Target
		typ      string
		tagsJSON string
		metaJSON string
	)

	err := s.Scan(&t.ID, &t.Name, &typ, &t.Address, &tagsJSON, &metaJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Target{}, domain.ErrTargetNotFound
	}
	if err != nil {
		return domain.Target{}, fmt.Errorf("postgres: scan target: %w", err)
	}

	t.Type = domain.TargetType(typ)

	if err := json.Unmarshal([]byte(tagsJSON), &t.Tags); err != nil {
		return domain.Target{}, fmt.Errorf("postgres: unmarshal tags: %w", err)
	}
	if err := json.Unmarshal([]byte(metaJSON), &t.Metadata); err != nil {
		return domain.Target{}, fmt.Errorf("postgres: unmarshal metadata: %w", err)
	}

	return t, nil
}
