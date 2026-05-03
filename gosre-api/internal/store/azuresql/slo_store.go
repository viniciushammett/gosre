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

// SLOStore implements store.SLOStore for Azure SQL.
type SLOStore struct {
	db *sql.DB
}

// SLOStore returns a SLOStore backed by the same database connection.
func (s *Store) SLOStore() *SLOStore {
	return &SLOStore{db: s.db}
}

// Save inserts or updates a SLO using a MERGE statement.
func (s *SLOStore) Save(ctx context.Context, slo domain.SLO) error {
	_, err := s.db.ExecContext(ctx, `
		MERGE slos WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5, @p6))
			AS S(id, target_id, name, metric, threshold, window_seconds)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.target_id=S.target_id, T.name=S.name, T.metric=S.metric,
			           T.threshold=S.threshold, T.window_seconds=S.window_seconds
		WHEN NOT MATCHED THEN
			INSERT (id, target_id, name, metric, threshold, window_seconds)
			VALUES (S.id, S.target_id, S.name, S.metric, S.threshold, S.window_seconds);`,
		slo.ID, slo.TargetID, slo.Name, slo.Metric, slo.Threshold, slo.WindowSeconds,
	)
	if err != nil {
		return fmt.Errorf("azuresql: save slo %q: %w", slo.ID, err)
	}
	return nil
}

// Get retrieves a SLO by ID. Returns sql.ErrNoRows if not present.
func (s *SLOStore) Get(ctx context.Context, id string) (domain.SLO, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, target_id, name, metric, threshold, window_seconds FROM slos WHERE id = @p1`, id)
	return scanSLO(row)
}

// ListByTarget returns all SLOs for the given targetID.
func (s *SLOStore) ListByTarget(ctx context.Context, targetID string) ([]domain.SLO, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, target_id, name, metric, threshold, window_seconds FROM slos WHERE target_id = @p1`, targetID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list slos for target %q: %w", targetID, err)
	}
	defer func() { _ = rows.Close() }()

	slos := make([]domain.SLO, 0)
	for rows.Next() {
		slo, err := scanSLO(rows)
		if err != nil {
			return nil, err
		}
		slos = append(slos, slo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azuresql: iterate slos: %w", err)
	}
	return slos, nil
}

// Delete removes a SLO by ID. Returns sql.ErrNoRows if not present.
func (s *SLOStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM slos WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete slo %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("azuresql: rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanSLO(s scanner) (domain.SLO, error) {
	var slo domain.SLO
	err := s.Scan(&slo.ID, &slo.TargetID, &slo.Name, &slo.Metric, &slo.Threshold, &slo.WindowSeconds)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SLO{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.SLO{}, fmt.Errorf("azuresql: scan slo: %w", err)
	}
	return slo, nil
}
