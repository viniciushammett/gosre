// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gosre/gosre-sdk/domain"
)

// Save inserts or replaces a Target in the database.
func (s *Store) Save(ctx context.Context, t domain.Target) error {
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("sqlite: marshal tags: %w", err)
	}

	meta, err := json.Marshal(t.Metadata)
	if err != nil {
		return fmt.Errorf("sqlite: marshal metadata: %w", err)
	}

	svcID := sql.NullString{}
	if t.ServiceID != nil {
		svcID = sql.NullString{String: *t.ServiceID, Valid: true}
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO targets (id, name, type, address, tags, metadata, project_id, service_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, string(t.Type), t.Address, string(tags), string(meta), t.ProjectID, svcID,
	)
	if err != nil {
		return fmt.Errorf("sqlite: save target %q: %w", t.ID, err)
	}
	return nil
}

// Get retrieves a Target by ID. Returns domain.ErrTargetNotFound if not present.
func (s *Store) Get(ctx context.Context, id string) (domain.Target, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, address, tags, metadata, project_id, service_id FROM targets WHERE id = ?`, id)

	return scanTarget(row)
}

// List returns all targets. Returns an empty (non-nil) slice when none exist.
func (s *Store) List(ctx context.Context) ([]domain.Target, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, address, tags, metadata, project_id, service_id FROM targets`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list targets: %w", err)
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
		return nil, fmt.Errorf("sqlite: iterate targets: %w", err)
	}
	return targets, nil
}

// Delete removes a Target by ID. Returns domain.ErrTargetNotFound if not present.
func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM targets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete target %q: %w", id, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrTargetNotFound
	}
	return nil
}

// scanner abstracts *sql.Row and *sql.Rows to share scanTarget.
type scanner interface {
	Scan(dest ...any) error
}

func scanTarget(s scanner) (domain.Target, error) {
	var (
		t        domain.Target
		typ      string
		tagsJSON string
		metaJSON string
		svcID    sql.NullString
	)

	err := s.Scan(&t.ID, &t.Name, &typ, &t.Address, &tagsJSON, &metaJSON, &t.ProjectID, &svcID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Target{}, domain.ErrTargetNotFound
	}
	if err != nil {
		return domain.Target{}, fmt.Errorf("sqlite: scan target: %w", err)
	}

	t.Type = domain.TargetType(typ)

	if err := json.Unmarshal([]byte(tagsJSON), &t.Tags); err != nil {
		return domain.Target{}, fmt.Errorf("sqlite: unmarshal tags: %w", err)
	}
	if err := json.Unmarshal([]byte(metaJSON), &t.Metadata); err != nil {
		return domain.Target{}, fmt.Errorf("sqlite: unmarshal metadata: %w", err)
	}

	if svcID.Valid {
		t.ServiceID = &svcID.String
	}

	return t, nil
}
