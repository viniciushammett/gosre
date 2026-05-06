// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// ResultStore implements store.ResultStore for SQLite.
type ResultStore struct {
	db *sql.DB
}

// ResultStore returns a ResultStore backed by the same database connection.
func (s *Store) ResultStore() *ResultStore {
	return &ResultStore{db: s.db}
}

// Save inserts or replaces a Result in the database.
func (s *ResultStore) Save(ctx context.Context, r domain.Result) error {
	meta, err := json.Marshal(r.Metadata)
	if err != nil {
		return fmt.Errorf("sqlite: marshal metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO results
		 (id, check_id, target_id, agent_id, status, duration, error, timestamp, metadata, project_id, target_name)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.CheckID, r.TargetID, r.AgentID,
		string(r.Status), r.Duration.Nanoseconds(), r.Error,
		r.Timestamp.UTC().Format(time.RFC3339Nano),
		string(meta), r.ProjectID, r.TargetName,
	)
	if err != nil {
		return fmt.Errorf("sqlite: save result %q: %w", r.ID, err)
	}
	return nil
}

// Get retrieves a Result by ID. Returns sql.ErrNoRows if not present.
func (s *ResultStore) Get(ctx context.Context, id string) (domain.Result, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, check_id, target_id, agent_id, status, duration, error, timestamp, metadata, project_id, target_name
		 FROM results WHERE id = ?`, id)
	return scanResult(row)
}

// List returns all Results ordered by timestamp DESC.
func (s *ResultStore) List(ctx context.Context) ([]domain.Result, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, check_id, target_id, agent_id, status, duration, error, timestamp, metadata, project_id, target_name
		 FROM results ORDER BY timestamp DESC`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list results: %w", err)
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
		return nil, fmt.Errorf("sqlite: iterate results: %w", err)
	}
	return results, nil
}

// ListFiltered returns Results matching the given filter, ordered by timestamp DESC.
// All filter fields are optional; zero values impose no constraint.
func (s *ResultStore) ListFiltered(ctx context.Context, f domain.ResultFilter) ([]domain.Result, error) {
	q := `SELECT id, check_id, target_id, agent_id, status, duration, error, timestamp, metadata, project_id, target_name
	      FROM results WHERE 1=1`
	args := make([]any, 0, 4)

	if f.TargetID != "" {
		q += " AND target_id = ?"
		args = append(args, f.TargetID)
	}
	if f.Status != "" {
		q += " AND status = ?"
		args = append(args, string(f.Status))
	}
	if !f.From.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, f.From.UTC().Format(time.RFC3339Nano))
	}
	if !f.To.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, f.To.UTC().Format(time.RFC3339Nano))
	}
	q += " ORDER BY timestamp DESC"

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list filtered results: %w", err)
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
		return nil, fmt.Errorf("sqlite: iterate filtered results: %w", err)
	}
	return results, nil
}

// DeleteByTargetID removes all Results associated with the given targetID.
func (s *ResultStore) DeleteByTargetID(ctx context.Context, targetID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM results WHERE target_id = ?`, targetID)
	if err != nil {
		return fmt.Errorf("sqlite: delete results for target %q: %w", targetID, err)
	}
	return nil
}

// ListByTarget returns all Results for the given targetID ordered by timestamp DESC.
func (s *ResultStore) ListByTarget(ctx context.Context, targetID string) ([]domain.Result, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, check_id, target_id, agent_id, status, duration, error, timestamp, metadata, project_id, target_name
		 FROM results WHERE target_id = ? ORDER BY timestamp DESC`, targetID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list results for target %q: %w", targetID, err)
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
		return nil, fmt.Errorf("sqlite: iterate results: %w", err)
	}
	return results, nil
}

func scanResult(s scanner) (domain.Result, error) {
	var (
		r          domain.Result
		status     string
		durationNs int64
		ts         string
		metaJSON   string
	)

	err := s.Scan(&r.ID, &r.CheckID, &r.TargetID, &r.AgentID,
		&status, &durationNs, &r.Error, &ts, &metaJSON, &r.ProjectID, &r.TargetName)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Result{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.Result{}, fmt.Errorf("sqlite: scan result: %w", err)
	}

	r.Status = domain.CheckStatus(status)
	r.Duration = time.Duration(durationNs)

	r.Timestamp, err = time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return domain.Result{}, fmt.Errorf("sqlite: parse timestamp %q: %w", ts, err)
	}

	if err := json.Unmarshal([]byte(metaJSON), &r.Metadata); err != nil {
		return domain.Result{}, fmt.Errorf("sqlite: unmarshal metadata: %w", err)
	}

	return r, nil
}
