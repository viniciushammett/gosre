// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package apiclient wraps the gosre-sdk HTTP client with the upsert+run
// semantics needed by the CLI check commands.
package apiclient

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	sdkclient "github.com/viniciushammett/gosre/gosre-sdk/client"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// Client delegates all HTTP communication to the gosre-sdk typed client,
// which provides retry, exponential backoff, and context propagation.
type Client struct {
	sdk *sdkclient.Client
}

// New constructs a Client for the given API base URL and optional API key.
func New(baseURL, apiKey string) *Client {
	return &Client{
		sdk: sdkclient.New(strings.TrimRight(baseURL, "/"), apiKey),
	}
}

// RunCheck upserts the target and check config in the API, executes the check,
// and returns the persisted Result. IDs are derived from the target address and
// check type so that repeated runs update the same records instead of creating
// new ones.
func (c *Client) RunCheck(ctx context.Context, t domain.Target, cfg domain.CheckConfig) (domain.Result, error) {
	t.ID = stableID(t.Address)
	cfg.ID = stableID(t.Address, string(cfg.Type))
	cfg.TargetID = t.ID

	if err := c.upsertTarget(ctx, t); err != nil {
		return domain.Result{}, fmt.Errorf("apiclient: upsert target: %w", err)
	}
	if err := c.upsertCheck(ctx, cfg); err != nil {
		return domain.Result{}, fmt.Errorf("apiclient: upsert check: %w", err)
	}
	return c.runCheck(ctx, cfg.ID)
}

func (c *Client) upsertTarget(ctx context.Context, t domain.Target) error {
	name := t.Name
	if name == "" {
		name = t.Address // API requires name; fall back to address when CLI omits it
	}
	_, err := c.sdk.CreateTarget(ctx, sdkclient.CreateTargetRequest{
		ID:       t.ID,
		Name:     name,
		Type:     string(t.Type),
		Address:  t.Address,
		Tags:     t.Tags,
		Metadata: t.Metadata,
	})
	return err
}

func (c *Client) upsertCheck(ctx context.Context, cfg domain.CheckConfig) error {
	_, err := c.sdk.CreateCheck(ctx, sdkclient.CreateCheckRequest{
		ID:       cfg.ID,
		Type:     string(cfg.Type),
		TargetID: cfg.TargetID,
		Interval: cfg.Interval.Nanoseconds(),
		Timeout:  cfg.Timeout.Nanoseconds(),
		Params:   cfg.Params,
	})
	return err
}

func (c *Client) runCheck(ctx context.Context, checkID string) (domain.Result, error) {
	r, err := c.sdk.RunCheck(ctx, checkID)
	if err != nil {
		return domain.Result{}, fmt.Errorf("apiclient: run check: %w", err)
	}
	return domain.Result{
		ID:         r.ID,
		CheckID:    r.CheckID,
		TargetID:   r.TargetID,
		AgentID:    r.AgentID,
		Status:     domain.CheckStatus(r.Status),
		Duration:   time.Duration(r.DurationMS),
		Error:      r.Error,
		Timestamp:  r.Timestamp,
		Metadata:   r.Metadata,
		TargetName: r.TargetName,
	}, nil
}

// stableID returns a stable decimal string derived from the given parts.
// The same inputs always produce the same ID, enabling idempotent upserts.
func stableID(parts ...string) string {
	h := fnv.New64a()
	for i, p := range parts {
		if i > 0 {
			_, _ = h.Write([]byte("-"))
		}
		_, _ = h.Write([]byte(p))
	}
	return fmt.Sprintf("%d", h.Sum64())
}
