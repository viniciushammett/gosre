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

	"github.com/gosre/gosre-sdk/domain"
)

// CheckStore implements store.CheckStore for SQLite.
type CheckStore struct {
	db *sql.DB
}

// CheckStore returns a CheckStore backed by the same database connection.
func (s *Store) CheckStore() *CheckStore {
	return &CheckStore{db: s.db}
}

// Save inserts or replaces a CheckConfig in the database.
func (s *CheckStore) Save(ctx context.Context, c domain.CheckConfig) error {
	params, err := json.Marshal(c.Params)
	if err != nil {
		return fmt.Errorf("sqlite: marshal params: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO checks (id, type, target_id, interval, timeout, params)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID, string(c.Type), c.TargetID,
		c.Interval.Nanoseconds(), c.Timeout.Nanoseconds(),
		string(params),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save check %q: %w", c.ID, err)
	}
	return nil
}

// Get retrieves a CheckConfig by ID. Returns sql.ErrNoRows if not present.
func (s *CheckStore) Get(ctx context.Context, id string) (domain.CheckConfig, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, type, target_id, interval, timeout, params
		 FROM checks WHERE id = ?`, id)
	return scanCheck(row)
}

// List returns all CheckConfigs. Returns an empty (non-nil) slice when none exist.
func (s *CheckStore) List(ctx context.Context) ([]domain.CheckConfig, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, type, target_id, interval, timeout, params FROM checks`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list checks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	checks := make([]domain.CheckConfig, 0)
	for rows.Next() {
		c, err := scanCheck(rows)
		if err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate checks: %w", err)
	}
	return checks, nil
}

// DeleteByTargetID removes all CheckConfigs associated with the given targetID.
func (s *CheckStore) DeleteByTargetID(ctx context.Context, targetID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM checks WHERE target_id = ?`, targetID)
	if err != nil {
		return fmt.Errorf("sqlite: delete checks for target %q: %w", targetID, err)
	}
	return nil
}

// Delete removes a CheckConfig by ID. Returns sql.ErrNoRows if not present.
func (s *CheckStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM checks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete check %q: %w", id, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanCheck(s scanner) (domain.CheckConfig, error) {
	var (
		c          domain.CheckConfig
		typ        string
		intervalNs int64
		timeoutNs  int64
		paramsJSON string
	)

	err := s.Scan(&c.ID, &typ, &c.TargetID, &intervalNs, &timeoutNs, &paramsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CheckConfig{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.CheckConfig{}, fmt.Errorf("sqlite: scan check: %w", err)
	}

	c.Type = domain.CheckType(typ)
	c.Interval = time.Duration(intervalNs)
	c.Timeout = time.Duration(timeoutNs)

	if err := json.Unmarshal([]byte(paramsJSON), &c.Params); err != nil {
		return domain.CheckConfig{}, fmt.Errorf("sqlite: unmarshal params: %w", err)
	}

	return c, nil
}
