// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package azuresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// IncidentStore implements store.IncidentStore for Azure SQL.
type IncidentStore struct {
	db *sql.DB
}

// IncidentStore returns an IncidentStore backed by the same database connection.
func (s *Store) IncidentStore() *IncidentStore {
	return &IncidentStore{db: s.db}
}

// Save inserts or updates an Incident using a MERGE statement.
func (s *IncidentStore) Save(ctx context.Context, i domain.Incident) error {
	ids, err := json.Marshal(i.ResultIDs)
	if err != nil {
		return fmt.Errorf("azuresql: marshal result_ids: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		MERGE incidents WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7))
			AS S(id, target_id, state, first_seen, last_seen, result_ids, project_id)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.target_id=S.target_id, T.state=S.state,
			           T.first_seen=S.first_seen, T.last_seen=S.last_seen, T.result_ids=S.result_ids,
			           T.project_id=S.project_id
		WHEN NOT MATCHED THEN
			INSERT (id, target_id, state, first_seen, last_seen, result_ids, project_id)
			VALUES (S.id, S.target_id, S.state, S.first_seen, S.last_seen, S.result_ids, S.project_id);`,
		i.ID, i.TargetID, string(i.State),
		i.FirstSeen.UTC(), i.LastSeen.UTC(), string(ids), i.ProjectID,
	)
	if err != nil {
		return fmt.Errorf("azuresql: save incident %q: %w", i.ID, err)
	}
	return nil
}

// Get retrieves an Incident by ID. Returns sql.ErrNoRows if not present.
func (s *IncidentStore) Get(ctx context.Context, id string) (domain.Incident, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, target_id, state, first_seen, last_seen, result_ids, project_id
		 FROM incidents WHERE id = @p1`, id)
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
			`SELECT id, target_id, state, first_seen, last_seen, result_ids, project_id
			 FROM incidents ORDER BY last_seen DESC`)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, target_id, state, first_seen, last_seen, result_ids, project_id
			 FROM incidents WHERE state = @p1 ORDER BY last_seen DESC`, string(state))
	}
	if err != nil {
		return nil, fmt.Errorf("azuresql: list incidents: %w", err)
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
		return nil, fmt.Errorf("azuresql: iterate incidents: %w", err)
	}
	return incidents, nil
}

// Update replaces an existing Incident record in full.
func (s *IncidentStore) Update(ctx context.Context, i domain.Incident) error {
	ids, err := json.Marshal(i.ResultIDs)
	if err != nil {
		return fmt.Errorf("azuresql: marshal result_ids: %w", err)
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE incidents
		SET target_id=@p1, state=@p2, first_seen=@p3, last_seen=@p4, result_ids=@p5
		WHERE id=@p6`,
		i.TargetID, string(i.State),
		i.FirstSeen.UTC(), i.LastSeen.UTC(),
		string(ids), i.ID,
	)
	if err != nil {
		return fmt.Errorf("azuresql: update incident %q: %w", i.ID, err)
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

// DeleteByTargetID removes all Incidents associated with the given targetID.
func (s *IncidentStore) DeleteByTargetID(ctx context.Context, targetID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM incidents WHERE target_id = @p1`, targetID)
	if err != nil {
		return fmt.Errorf("azuresql: delete incidents for target %q: %w", targetID, err)
	}
	return nil
}

func scanIncident(s scanner) (domain.Incident, error) {
	var (
		i       domain.Incident
		state   string
		idsJSON string
	)
	err := s.Scan(&i.ID, &i.TargetID, &state, &i.FirstSeen, &i.LastSeen, &idsJSON, &i.ProjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Incident{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.Incident{}, fmt.Errorf("azuresql: scan incident: %w", err)
	}
	i.State = domain.IncidentState(state)
	i.FirstSeen = i.FirstSeen.UTC()
	i.LastSeen = i.LastSeen.UTC()
	if err := json.Unmarshal([]byte(idsJSON), &i.ResultIDs); err != nil {
		return domain.Incident{}, fmt.Errorf("azuresql: unmarshal result_ids: %w", err)
	}
	return i, nil
}
