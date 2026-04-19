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

// IncidentStore implements store.IncidentStore for SQLite.
type IncidentStore struct {
	db *sql.DB
}

// IncidentStore returns an IncidentStore backed by the same database connection.
func (s *Store) IncidentStore() *IncidentStore {
	return &IncidentStore{db: s.db}
}

// Save inserts or replaces an Incident in the database.
func (s *IncidentStore) Save(ctx context.Context, i domain.Incident) error {
	ids, err := json.Marshal(i.ResultIDs)
	if err != nil {
		return fmt.Errorf("sqlite: marshal result_ids: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO incidents
		 (id, target_id, state, first_seen, last_seen, result_ids)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		i.ID, i.TargetID, string(i.State),
		i.FirstSeen.UTC().Format(time.RFC3339Nano),
		i.LastSeen.UTC().Format(time.RFC3339Nano),
		string(ids),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save incident %q: %w", i.ID, err)
	}
	return nil
}

// Get retrieves an Incident by ID. Returns sql.ErrNoRows if not present.
func (s *IncidentStore) Get(ctx context.Context, id string) (domain.Incident, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, target_id, state, first_seen, last_seen, result_ids
		 FROM incidents WHERE id = ?`, id)
	return scanIncident(row)
}

// ListByState returns all Incidents with the given state ordered by last_seen DESC.
// An empty state returns all incidents.
func (s *IncidentStore) ListByState(ctx context.Context, state domain.IncidentState) ([]domain.Incident, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if state == "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, target_id, state, first_seen, last_seen, result_ids
			 FROM incidents ORDER BY last_seen DESC`)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, target_id, state, first_seen, last_seen, result_ids
			 FROM incidents WHERE state = ? ORDER BY last_seen DESC`, string(state))
	}
	if err != nil {
		return nil, fmt.Errorf("sqlite: list incidents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	incidents := make([]domain.Incident, 0)
	for rows.Next() {
		i, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate incidents: %w", err)
	}
	return incidents, nil
}

// Update replaces an existing Incident record in full.
func (s *IncidentStore) Update(ctx context.Context, i domain.Incident) error {
	ids, err := json.Marshal(i.ResultIDs)
	if err != nil {
		return fmt.Errorf("sqlite: marshal result_ids: %w", err)
	}

	res, err := s.db.ExecContext(ctx,
		`UPDATE incidents
		 SET target_id = ?, state = ?, first_seen = ?, last_seen = ?, result_ids = ?
		 WHERE id = ?`,
		i.TargetID, string(i.State),
		i.FirstSeen.UTC().Format(time.RFC3339Nano),
		i.LastSeen.UTC().Format(time.RFC3339Nano),
		string(ids), i.ID,
	)
	if err != nil {
		return fmt.Errorf("sqlite: update incident %q: %w", i.ID, err)
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

// DeleteByTargetID removes all Incidents associated with the given targetID.
func (s *IncidentStore) DeleteByTargetID(ctx context.Context, targetID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM incidents WHERE target_id = ?`, targetID)
	if err != nil {
		return fmt.Errorf("sqlite: delete incidents for target %q: %w", targetID, err)
	}
	return nil
}

func scanIncident(s scanner) (domain.Incident, error) {
	var (
		i         domain.Incident
		state     string
		firstSeen string
		lastSeen  string
		idsJSON   string
	)

	err := s.Scan(&i.ID, &i.TargetID, &state, &firstSeen, &lastSeen, &idsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Incident{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.Incident{}, fmt.Errorf("sqlite: scan incident: %w", err)
	}

	i.State = domain.IncidentState(state)

	i.FirstSeen, err = time.Parse(time.RFC3339Nano, firstSeen)
	if err != nil {
		return domain.Incident{}, fmt.Errorf("sqlite: parse first_seen %q: %w", firstSeen, err)
	}

	i.LastSeen, err = time.Parse(time.RFC3339Nano, lastSeen)
	if err != nil {
		return domain.Incident{}, fmt.Errorf("sqlite: parse last_seen %q: %w", lastSeen, err)
	}

	if err := json.Unmarshal([]byte(idsJSON), &i.ResultIDs); err != nil {
		return domain.Incident{}, fmt.Errorf("sqlite: unmarshal result_ids: %w", err)
	}

	return i, nil
}
