// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/gosre/gosre-sdk/store"
)

// NotificationService handles business logic for NotificationChannel and NotificationRule entities.
type NotificationService struct {
	channels store.NotificationChannelStore
	rules    store.NotificationRuleStore
}

// NewNotificationService constructs a NotificationService backed by the given stores.
func NewNotificationService(channels store.NotificationChannelStore, rules store.NotificationRuleStore) *NotificationService {
	return &NotificationService{channels: channels, rules: rules}
}

// CreateChannel validates and persists a new NotificationChannel.
func (svc *NotificationService) CreateChannel(ctx context.Context, c domain.NotificationChannel) (domain.NotificationChannel, error) {
	if c.Name == "" {
		return domain.NotificationChannel{}, fmt.Errorf("name is required")
	}
	if c.ProjectID == "" {
		return domain.NotificationChannel{}, fmt.Errorf("project_id is required")
	}
	switch c.Kind {
	case domain.ChannelKindSlack, domain.ChannelKindEmail, domain.ChannelKindWebhook:
	default:
		return domain.NotificationChannel{}, fmt.Errorf("invalid kind %q: must be slack, email, or webhook", c.Kind)
	}
	if c.ID == "" {
		c.ID = newNotifID()
	}
	if c.Config == nil {
		c.Config = map[string]string{}
	}
	if err := svc.channels.Save(ctx, c); err != nil {
		return domain.NotificationChannel{}, fmt.Errorf("create channel: %w", err)
	}
	return c, nil
}

// GetChannel retrieves a NotificationChannel by ID.
func (svc *NotificationService) GetChannel(ctx context.Context, id string) (domain.NotificationChannel, error) {
	return svc.channels.Get(ctx, id)
}

// ListChannels returns all NotificationChannels for the given project.
func (svc *NotificationService) ListChannels(ctx context.Context, projectID string) ([]domain.NotificationChannel, error) {
	return svc.channels.ListByProject(ctx, projectID)
}

// DeleteChannel removes a NotificationChannel by ID.
func (svc *NotificationService) DeleteChannel(ctx context.Context, id string) error {
	return svc.channels.Delete(ctx, id)
}

// CreateRule validates and persists a new NotificationRule.
func (svc *NotificationService) CreateRule(ctx context.Context, r domain.NotificationRule) (domain.NotificationRule, error) {
	if r.ProjectID == "" {
		return domain.NotificationRule{}, fmt.Errorf("project_id is required")
	}
	if r.ChannelID == "" {
		return domain.NotificationRule{}, fmt.Errorf("channel_id is required")
	}
	if r.EventKind == "" {
		return domain.NotificationRule{}, fmt.Errorf("event_kind is required")
	}
	if r.ID == "" {
		r.ID = newNotifID()
	}
	if r.TagFilter == nil {
		r.TagFilter = []string{}
	}
	if err := svc.rules.Save(ctx, r); err != nil {
		return domain.NotificationRule{}, fmt.Errorf("create rule: %w", err)
	}
	return r, nil
}

// GetRule retrieves a NotificationRule by ID.
func (svc *NotificationService) GetRule(ctx context.Context, id string) (domain.NotificationRule, error) {
	return svc.rules.Get(ctx, id)
}

// ListRules returns all NotificationRules for the given project.
func (svc *NotificationService) ListRules(ctx context.Context, projectID string) ([]domain.NotificationRule, error) {
	return svc.rules.ListByProject(ctx, projectID)
}

// DeleteRule removes a NotificationRule by ID.
func (svc *NotificationService) DeleteRule(ctx context.Context, id string) error {
	return svc.rules.Delete(ctx, id)
}

// newNotifID returns a random UUID v4 using crypto/rand.
func newNotifID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("gosre-api: crypto/rand unavailable: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
