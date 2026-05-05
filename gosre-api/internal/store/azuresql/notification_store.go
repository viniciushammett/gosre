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

// NotificationChannelStore implements store.NotificationChannelStore for Azure SQL.
type NotificationChannelStore struct {
	db *sql.DB
}

// NotificationChannelStore returns a NotificationChannelStore backed by the same connection.
func (s *Store) NotificationChannelStore() *NotificationChannelStore {
	return &NotificationChannelStore{db: s.db}
}

// Save inserts or updates a NotificationChannel using a MERGE statement.
func (s *NotificationChannelStore) Save(ctx context.Context, c domain.NotificationChannel) error {
	cfg, err := json.Marshal(c.Config)
	if err != nil {
		return fmt.Errorf("azuresql: marshal channel config: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		MERGE notification_channels WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5))
			AS S(id, project_id, name, kind, config)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.project_id=S.project_id, T.name=S.name, T.kind=S.kind, T.config=S.config
		WHEN NOT MATCHED THEN
			INSERT (id, project_id, name, kind, config)
			VALUES (S.id, S.project_id, S.name, S.kind, S.config);`,
		c.ID, c.ProjectID, c.Name, string(c.Kind), string(cfg),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save notification_channel %q: %w", c.ID, err)
	}
	return nil
}

// Get retrieves a NotificationChannel by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationChannelStore) Get(ctx context.Context, id string) (domain.NotificationChannel, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, name, kind, config FROM notification_channels WHERE id = @p1`, id)
	return scanNotifChannel(row)
}

// ListByProject returns all NotificationChannels for the given project.
func (s *NotificationChannelStore) ListByProject(ctx context.Context, projectID string) ([]domain.NotificationChannel, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, name, kind, config FROM notification_channels WHERE project_id = @p1`, projectID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list notification_channels for project %q: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()

	channels := make([]domain.NotificationChannel, 0)
	for rows.Next() {
		ch, err := scanNotifChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azuresql: iterate notification_channels: %w", err)
	}
	return channels, nil
}

// Delete removes a NotificationChannel by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationChannelStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM notification_channels WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete notification_channel %q: %w", id, err)
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

func scanNotifChannel(s scanner) (domain.NotificationChannel, error) {
	var c domain.NotificationChannel
	var kindStr, cfgStr string
	err := s.Scan(&c.ID, &c.ProjectID, &c.Name, &kindStr, &cfgStr)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NotificationChannel{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.NotificationChannel{}, fmt.Errorf("azuresql: scan notification_channel: %w", err)
	}
	c.Kind = domain.NotificationChannelKind(kindStr)
	if err := json.Unmarshal([]byte(cfgStr), &c.Config); err != nil {
		return domain.NotificationChannel{}, fmt.Errorf("azuresql: unmarshal channel config: %w", err)
	}
	return c, nil
}

// NotificationRuleStore implements store.NotificationRuleStore for Azure SQL.
type NotificationRuleStore struct {
	db *sql.DB
}

// NotificationRuleStore returns a NotificationRuleStore backed by the same connection.
func (s *Store) NotificationRuleStore() *NotificationRuleStore {
	return &NotificationRuleStore{db: s.db}
}

// Save inserts or updates a NotificationRule using a MERGE statement.
func (s *NotificationRuleStore) Save(ctx context.Context, r domain.NotificationRule) error {
	tf, err := json.Marshal(r.TagFilter)
	if err != nil {
		return fmt.Errorf("azuresql: marshal rule tag_filter: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		MERGE notification_rules WITH (HOLDLOCK) AS T
		USING (VALUES (@p1, @p2, @p3, @p4, @p5))
			AS S(id, project_id, channel_id, event_kind, tag_filter)
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET T.project_id=S.project_id, T.channel_id=S.channel_id,
			           T.event_kind=S.event_kind, T.tag_filter=S.tag_filter
		WHEN NOT MATCHED THEN
			INSERT (id, project_id, channel_id, event_kind, tag_filter)
			VALUES (S.id, S.project_id, S.channel_id, S.event_kind, S.tag_filter);`,
		r.ID, r.ProjectID, r.ChannelID, r.EventKind, string(tf),
	)
	if err != nil {
		return fmt.Errorf("azuresql: save notification_rule %q: %w", r.ID, err)
	}
	return nil
}

// Get retrieves a NotificationRule by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationRuleStore) Get(ctx context.Context, id string) (domain.NotificationRule, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, channel_id, event_kind, tag_filter FROM notification_rules WHERE id = @p1`, id)
	return scanNotifRule(row)
}

// ListByProject returns all NotificationRules for the given project.
func (s *NotificationRuleStore) ListByProject(ctx context.Context, projectID string) ([]domain.NotificationRule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, channel_id, event_kind, tag_filter FROM notification_rules WHERE project_id = @p1`, projectID)
	if err != nil {
		return nil, fmt.Errorf("azuresql: list notification_rules for project %q: %w", projectID, err)
	}
	defer func() { _ = rows.Close() }()

	rules := make([]domain.NotificationRule, 0)
	for rows.Next() {
		r, err := scanNotifRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azuresql: iterate notification_rules: %w", err)
	}
	return rules, nil
}

// Delete removes a NotificationRule by ID. Returns sql.ErrNoRows if absent.
func (s *NotificationRuleStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM notification_rules WHERE id = @p1`, id)
	if err != nil {
		return fmt.Errorf("azuresql: delete notification_rule %q: %w", id, err)
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

func scanNotifRule(s scanner) (domain.NotificationRule, error) {
	var r domain.NotificationRule
	var tfStr string
	err := s.Scan(&r.ID, &r.ProjectID, &r.ChannelID, &r.EventKind, &tfStr)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NotificationRule{}, sql.ErrNoRows
	}
	if err != nil {
		return domain.NotificationRule{}, fmt.Errorf("azuresql: scan notification_rule: %w", err)
	}
	if err := json.Unmarshal([]byte(tfStr), &r.TagFilter); err != nil {
		return domain.NotificationRule{}, fmt.Errorf("azuresql: unmarshal rule tag_filter: %w", err)
	}
	return r, nil
}
