// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// NotificationChannelStore implements store.NotificationChannelStore for SQLite.
type NotificationChannelStore struct {
	db *sql.DB
}

// NotificationChannelStore returns a NotificationChannelStore backed by the same connection.
func (s *Store) NotificationChannelStore() *NotificationChannelStore {
	return &NotificationChannelStore{db: s.db}
}

// Save inserts or replaces a NotificationChannel.
func (s *NotificationChannelStore) Save(ctx context.Context, c domain.NotificationChannel) error {
	cfg, err := json.Marshal(c.Config)
	if err != nil {
		return fmt.Errorf("sqlite: marshal channel config: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO notification_channels (id, project_id, name, kind, config)
		 VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.ProjectID, c.Name, string(c.Kind), string(cfg),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save notification_channel %q: %w", c.ID, err)
	}
	return nil
}

// Get retrieves a NotificationChannel by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationChannelStore) Get(ctx context.Context, id string) (domain.NotificationChannel, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, name, kind, config FROM notification_channels WHERE id = ?`, id)
	return scanChannel(row)
}

// ListByProject returns all NotificationChannels for the given project.
func (s *NotificationChannelStore) ListByProject(ctx context.Context, projectID string) ([]domain.NotificationChannel, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, name, kind, config FROM notification_channels WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list notification_channels for project %q: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()

	channels := make([]domain.NotificationChannel, 0)
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate notification_channels: %w", err)
	}
	return channels, nil
}

// Delete removes a NotificationChannel by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationChannelStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM notification_channels WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete notification_channel %q: %w", id, err)
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

func scanChannel(s scanner) (domain.NotificationChannel, error) {
	var c domain.NotificationChannel
	var kindStr, cfgStr string
	err := s.Scan(&c.ID, &c.ProjectID, &c.Name, &kindStr, &cfgStr)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NotificationChannel{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.NotificationChannel{}, fmt.Errorf("sqlite: scan notification_channel: %w", err)
	}
	c.Kind = domain.NotificationChannelKind(kindStr)
	if err := json.Unmarshal([]byte(cfgStr), &c.Config); err != nil {
		return domain.NotificationChannel{}, fmt.Errorf("sqlite: unmarshal channel config: %w", err)
	}
	return c, nil
}

// NotificationRuleStore implements store.NotificationRuleStore for SQLite.
type NotificationRuleStore struct {
	db *sql.DB
}

// NotificationRuleStore returns a NotificationRuleStore backed by the same connection.
func (s *Store) NotificationRuleStore() *NotificationRuleStore {
	return &NotificationRuleStore{db: s.db}
}

// Save inserts or replaces a NotificationRule.
func (s *NotificationRuleStore) Save(ctx context.Context, r domain.NotificationRule) error {
	tf, err := json.Marshal(r.TagFilter)
	if err != nil {
		return fmt.Errorf("sqlite: marshal rule tag_filter: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO notification_rules (id, project_id, channel_id, event_kind, tag_filter)
		 VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.ProjectID, r.ChannelID, r.EventKind, string(tf),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save notification_rule %q: %w", r.ID, err)
	}
	return nil
}

// Get retrieves a NotificationRule by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationRuleStore) Get(ctx context.Context, id string) (domain.NotificationRule, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, channel_id, event_kind, tag_filter FROM notification_rules WHERE id = ?`, id)
	return scanRule(row)
}

// ListByProject returns all NotificationRules for the given project.
func (s *NotificationRuleStore) ListByProject(ctx context.Context, projectID string) ([]domain.NotificationRule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, channel_id, event_kind, tag_filter FROM notification_rules WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list notification_rules for project %q: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()

	rules := make([]domain.NotificationRule, 0)
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate notification_rules: %w", err)
	}
	return rules, nil
}

// Delete removes a NotificationRule by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationRuleStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM notification_rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete notification_rule %q: %w", id, err)
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

func scanRule(s scanner) (domain.NotificationRule, error) {
	var r domain.NotificationRule
	var tfStr string
	err := s.Scan(&r.ID, &r.ProjectID, &r.ChannelID, &r.EventKind, &tfStr)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NotificationRule{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.NotificationRule{}, fmt.Errorf("sqlite: scan notification_rule: %w", err)
	}
	if err := json.Unmarshal([]byte(tfStr), &r.TagFilter); err != nil {
		return domain.NotificationRule{}, fmt.Errorf("sqlite: unmarshal rule tag_filter: %w", err)
	}
	return r, nil
}
