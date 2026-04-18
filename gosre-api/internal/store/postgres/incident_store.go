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

// IncidentStore implements store.IncidentStore for PostgreSQL.
type IncidentStore struct {
	db *sql.DB
}

// IncidentStore returns an IncidentStore backed by the same database connection.
func (s *Store) IncidentStore() *IncidentStore {
	return &IncidentStore{db: s.db}
}

// Save inserts or updates an Incident in the database.
func (s *IncidentStore) Save(ctx context.Context, i domain.Incident) error {
	ids, err := json.Marshal(i.ResultIDs)
	if err != nil {
		return fmt.Errorf("postgres: marshal result_ids: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO incidents (id, target_id, state, first_seen, last_seen, result_ids)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE
		 SET target_id=$2, state=$3, first_seen=$4, last_seen=$5, result_ids=$6`,
		i.ID, i.TargetID, string(i.State),
		i.FirstSeen.UTC(), i.LastSeen.UTC(), string(ids),
	)
	if err != nil {
		return fmt.Errorf("postgres: save incident %q: %w", i.ID, err)
	}
	return nil
}

// Get retrieves an Incident by ID. Returns sql.ErrNoRows if not present.
func (s *IncidentStore) Get(ctx context.Context, id string) (domain.Incident, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, target_id, state, first_seen, last_seen, result_ids
		 FROM incidents WHERE id = $1`, id)
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
			 FROM incidents WHERE state = $1 ORDER BY last_seen DESC`, string(state))
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: list incidents: %w", err)
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
		return nil, fmt.Errorf("postgres: iterate incidents: %w", err)
	}
	return incidents, nil
}

// Update replaces an existing Incident record in full.
func (s *IncidentStore) Update(ctx context.Context, i domain.Incident) error {
	ids, err := json.Marshal(i.ResultIDs)
	if err != nil {
		return fmt.Errorf("postgres: marshal result_ids: %w", err)
	}

	res, err := s.db.ExecContext(ctx,
		`UPDATE incidents
		 SET target_id=$1, state=$2, first_seen=$3, last_seen=$4, result_ids=$5
		 WHERE id=$6`,
		i.TargetID, string(i.State),
		i.FirstSeen.UTC(), i.LastSeen.UTC(),
		string(ids), i.ID,
	)
	if err != nil {
		return fmt.Errorf("postgres: update incident %q: %w", i.ID, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanIncident(s scanner) (domain.Incident, error) {
	var (
		i       domain.Incident
		state   string
		idsJSON string
	)

	err := s.Scan(&i.ID, &i.TargetID, &state, &i.FirstSeen, &i.LastSeen, &idsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Incident{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.Incident{}, fmt.Errorf("postgres: scan incident: %w", err)
	}

	i.State = domain.IncidentState(state)
	i.FirstSeen = i.FirstSeen.UTC()
	i.LastSeen = i.LastSeen.UTC()

	if err := json.Unmarshal([]byte(idsJSON), &i.ResultIDs); err != nil {
		return domain.Incident{}, fmt.Errorf("postgres: unmarshal result_ids: %w", err)
	}

	return i, nil
}
