// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	v1 "github.com/viniciushammett/gosre/gosre-api/internal/api/v1"
)

// AgentRecord is a type alias for the shared agent record type defined in the API handler.
type AgentRecord = v1.AgentRecord

// AgentStore provides agent persistence backed by SQLite.
type AgentStore struct {
	db *sql.DB
}

// AgentStore returns an AgentStore backed by the same database connection.
func (s *Store) AgentStore() *AgentStore {
	return &AgentStore{db: s.db}
}

// Register inserts or replaces an agent record.
func (s *AgentStore) Register(ctx context.Context, a AgentRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO agents (id, hostname, version, last_seen)
		 VALUES (?, ?, ?, ?)`,
		a.ID, a.Hostname, a.Version, a.LastSeen,
	)
	if err != nil {
		return fmt.Errorf("sqlite: register agent %q: %w", a.ID, err)
	}
	return nil
}

// List returns all registered agents ordered by last_seen descending.
func (s *AgentStore) List(ctx context.Context) ([]AgentRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, hostname, version, last_seen FROM agents ORDER BY last_seen DESC`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list agents: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []AgentRecord
	for rows.Next() {
		var r AgentRecord
		if err := rows.Scan(&r.ID, &r.Hostname, &r.Version, &r.LastSeen); err != nil {
			return nil, fmt.Errorf("sqlite: scan agent: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Heartbeat updates the last_seen timestamp for an agent.
func (s *AgentStore) Heartbeat(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE agents SET last_seen = ? WHERE id = ?`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("sqlite: heartbeat agent %q: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
