// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package azuresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gosre/gosre-sdk/domain"
)

// ResultStore implements store.ResultStore for Azure SQL.
type ResultStore struct {
	db *sql.DB
}

// ResultStore returns a ResultStore backed by the same database connection.
func (s *Store) ResultStore() *ResultStore {
	return &ResultStore{db: s.db}
}

// Save inserts or updates a Result using a MERGE statement.
func (s *ResultStore) Save(ctx context.Context, r domain.Result) error {
	meta, err := json.Marshal(r.Metadata)
	if err != nil {
		return fmt.Errorf("azuresql: marshal metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		MERGE results WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9))
			AS S(id, check_id, target_id, agent_id, status, duration_ns, error, timestamp, metadata)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.check_id=S.check_id, T.target_id=S.target_id, T.agent_id=S.agent_id,
			           T.status=S.status, T.duration_ns=S.duration_ns, T.error=S.error,
			           T.timestamp=S.timestamp, T.metadata=S.metadata
		WHEN NOT MATCHED THEN
			INSERT (id, check_id, target_id, agent_id, status, duration_ns, error, timestamp, metadata)
			VALUES (S.id, S.check_id, S.target_id, S.agent_id, S.status, S.duration_ns, S.error, S.timestamp, S.metadata);`,
		r.ID, r.CheckID, r.TargetID, r.AgentID,
		string(r.Status), r.Duration.Nanoseconds(), r.Error,
		r.Timestamp.UTC(), string(meta),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save result %q: %w", r.ID, err)
	}
	return nil
}

// Get retrieves a Result by ID. Returns sql.ErrNoRows if not present.
func (s *ResultStore) Get(ctx context.Context, id string) (domain.Result, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, check_id, target_id, agent_id, status, duration_ns, error, timestamp, metadata
		 FROM results WHERE id = @p1`, id)
	return scanResult(row)
}

// List returns all Results ordered by timestamp DESC.
func (s *ResultStore) List(ctx context.Context) ([]domain.Result, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, check_id, target_id, agent_id, status, duration_ns, error, timestamp, metadata
		 FROM results ORDER BY timestamp DESC`)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list results: %w", err)
	}
	defer func() { _ = rows.Close() }()

	results := make([]domain.Result, 0)
	for rows.Next() {
		r, err := scanResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azuresql: iterate results: %w", err)
	}
	return results, nil
}

// ListByTarget returns all Results for the given targetID ordered by timestamp DESC.
func (s *ResultStore) ListByTarget(ctx context.Context, targetID string) ([]domain.Result, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, check_id, target_id, agent_id, status, duration_ns, error, timestamp, metadata
		 FROM results WHERE target_id = @p1 ORDER BY timestamp DESC`, targetID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list results for target %q: %w", targetID, err)
	}
	defer func() { _ = rows.Close() }()

	results := make([]domain.Result, 0)
	for rows.Next() {
		r, err := scanResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azuresql: iterate results: %w", err)
	}
	return results, nil
}

// DeleteByTargetID removes all Results associated with the given targetID.
func (s *ResultStore) DeleteByTargetID(ctx context.Context, targetID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM results WHERE target_id = @p1`, targetID)
	if err != nil {
		return fmt.Errorf("azuresql: delete results for target %q: %w", targetID, err)
	}
	return nil
}

func scanResult(s scanner) (domain.Result, error) {
	var (
		r          domain.Result
		status     string
		durationNs int64
		metaJSON   string
	)
	err := s.Scan(&r.ID, &r.CheckID, &r.TargetID, &r.AgentID,
		&status, &durationNs, &r.Error, &r.Timestamp, &metaJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Result{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.Result{}, fmt.Errorf("azuresql: scan result: %w", err)
	}
	r.Status = domain.CheckStatus(status)
	r.Duration = time.Duration(durationNs)
	r.Timestamp = r.Timestamp.UTC()
	if err := json.Unmarshal([]byte(metaJSON), &r.Metadata); err != nil {
		return domain.Result{}, fmt.Errorf("azuresql: unmarshal metadata: %w", err)
	}
	return r, nil
}
