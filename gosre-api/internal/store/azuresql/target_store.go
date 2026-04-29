// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package azuresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gosre/gosre-sdk/domain"
)

// TargetStore implements store.TargetStore for Azure SQL.
type TargetStore struct {
	db *sql.DB
}

// TargetStore returns a TargetStore backed by the same database connection.
func (s *Store) TargetStore() *TargetStore {
	return &TargetStore{db: s.db}
}

// Save inserts or updates a Target using a MERGE statement.
func (s *TargetStore) Save(ctx context.Context, t domain.Target) error {
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("azuresql: marshal tags: %w", err)
	}
	meta, err := json.Marshal(t.Metadata)
	if err != nil {
		return fmt.Errorf("azuresql: marshal metadata: %w", err)
	}

	svcID := sql.NullString{}
	if t.ServiceID != nil {
		svcID = sql.NullString{String: *t.ServiceID, Valid: true}
	}

	_, err = s.db.ExecContext(ctx, `
		MERGE targets WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8))
			AS S(id, name, type, address, tags, metadata, project_id, service_id)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.name=S.name, T.type=S.type, T.address=S.address, T.tags=S.tags, T.metadata=S.metadata,
			           T.project_id=S.project_id, T.service_id=S.service_id
		WHEN NOT MATCHED THEN
			INSERT (id, name, type, address, tags, metadata, project_id, service_id)
			VALUES (S.id, S.name, S.type, S.address, S.tags, S.metadata, S.project_id, S.service_id);`,
		t.ID, t.Name, string(t.Type), t.Address, string(tags), string(meta), t.ProjectID, svcID,
	)
	if err != nil {
		return fmt.Errorf("azuresql: save target %q: %w", t.ID, err)
	}
	return nil
}

// Get retrieves a Target by ID. Returns domain.ErrTargetNotFound if not present.
func (s *TargetStore) Get(ctx context.Context, id string) (domain.Target, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, address, tags, metadata, project_id, service_id FROM targets WHERE id = @p1`, id)
	return scanTarget(row)
}

// List returns all targets. Returns an empty (non-nil) slice when none exist.
func (s *TargetStore) List(ctx context.Context) ([]domain.Target, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, address, tags, metadata, project_id, service_id FROM targets`)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list targets: %w", err)
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
		return nil, fmt.Errorf("azuresql: iterate targets: %w", err)
	}
	return targets, nil
}

// Delete removes a Target by ID. Returns domain.ErrTargetNotFound if not present.
func (s *TargetStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM targets WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete target %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("azuresql: rows affected: %w", err)
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
		svcID    sql.NullString
	)
	err := s.Scan(&t.ID, &t.Name, &typ, &t.Address, &tagsJSON, &metaJSON, &t.ProjectID, &svcID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Target{}, domain.ErrTargetNotFound
	}
	if err != nil {
		return domain.Target{}, fmt.Errorf("azuresql: scan target: %w", err)
	}
	t.Type = domain.TargetType(typ)
	if err := json.Unmarshal([]byte(tagsJSON), &t.Tags); err != nil {
		return domain.Target{}, fmt.Errorf("azuresql: unmarshal tags: %w", err)
	}
	if err := json.Unmarshal([]byte(metaJSON), &t.Metadata); err != nil {
		return domain.Target{}, fmt.Errorf("azuresql: unmarshal metadata: %w", err)
	}
	if svcID.Valid {
		t.ServiceID = &svcID.String
	}
	return t, nil
}
