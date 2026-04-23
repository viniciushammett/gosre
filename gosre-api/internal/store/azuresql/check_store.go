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

// CheckStore implements store.CheckStore for Azure SQL.
type CheckStore struct {
	db *sql.DB
}

// CheckStore returns a CheckStore backed by the same database connection.
func (s *Store) CheckStore() *CheckStore {
	return &CheckStore{db: s.db}
}

// Save inserts or updates a CheckConfig using a MERGE statement.
func (s *CheckStore) Save(ctx context.Context, c domain.CheckConfig) error {
	params, err := json.Marshal(c.Params)
	if err != nil {
		return fmt.Errorf("azuresql: marshal params: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		MERGE checks WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7))
			AS S(id, type, target_id, interval_ns, timeout_ns, params, project_id)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.type=S.type, T.target_id=S.target_id,
			           T.interval_ns=S.interval_ns, T.timeout_ns=S.timeout_ns, T.params=S.params,
			           T.project_id=S.project_id
		WHEN NOT MATCHED THEN
			INSERT (id, type, target_id, interval_ns, timeout_ns, params, project_id)
			VALUES (S.id, S.type, S.target_id, S.interval_ns, S.timeout_ns, S.params, S.project_id);`,
		c.ID, string(c.Type), c.TargetID,
		c.Interval.Nanoseconds(), c.Timeout.Nanoseconds(), string(params), c.ProjectID,
	)
	if err != nil {
		return fmt.Errorf("azuresql: save check %q: %w", c.ID, err)
	}
	return nil
}

// Get retrieves a CheckConfig by ID. Returns sql.ErrNoRows if not present.
func (s *CheckStore) Get(ctx context.Context, id string) (domain.CheckConfig, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, type, target_id, interval_ns, timeout_ns, params, project_id FROM checks WHERE id = @p1`, id)
	return scanCheck(row)
}

// List returns all CheckConfigs. Returns an empty (non-nil) slice when none exist.
func (s *CheckStore) List(ctx context.Context) ([]domain.CheckConfig, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, type, target_id, interval_ns, timeout_ns, params, project_id FROM checks`)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list checks: %w", err)
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
		return nil, fmt.Errorf("azuresql: iterate checks: %w", err)
	}
	return checks, nil
}

// Delete removes a CheckConfig by ID. Returns sql.ErrNoRows if not present.
func (s *CheckStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM checks WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete check %q: %w", id, err)
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

// DeleteByTargetID removes all CheckConfigs associated with the given targetID.
func (s *CheckStore) DeleteByTargetID(ctx context.Context, targetID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM checks WHERE target_id = @p1`, targetID)
	if err != nil {
		return fmt.Errorf("azuresql: delete checks for target %q: %w", targetID, err)
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
	err := s.Scan(&c.ID, &typ, &c.TargetID, &intervalNs, &timeoutNs, &paramsJSON, &c.ProjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CheckConfig{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.CheckConfig{}, fmt.Errorf("azuresql: scan check: %w", err)
	}
	c.Type = domain.CheckType(typ)
	c.Interval = time.Duration(intervalNs)
	c.Timeout = time.Duration(timeoutNs)
	if err := json.Unmarshal([]byte(paramsJSON), &c.Params); err != nil {
		return domain.CheckConfig{}, fmt.Errorf("azuresql: unmarshal params: %w", err)
	}
	return c, nil
}
