// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package apiclient fetches notification configuration from gosre-api over HTTP.
package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gosre/gosre-sdk/domain"
)

// Client fetches notification rules and channels from gosre-api.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New constructs a Client. token is a Bearer JWT issued for the notifier service.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type apiEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *apiErr         `json:"error"`
}

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ListRulesByProject returns all notification rules for the given project.
func (c *Client) ListRulesByProject(ctx context.Context, projectID string) ([]domain.NotificationRule, error) {
	raw, err := c.get(ctx, "/api/v1/notification/rules?project_id="+projectID)
	if err != nil {
		return nil, err
	}
	var rules []domain.NotificationRule
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, fmt.Errorf("apiclient: decode rules: %w", err)
	}
	return rules, nil
}

// GetChannel returns a notification channel by ID.
func (c *Client) GetChannel(ctx context.Context, id string) (domain.NotificationChannel, error) {
	raw, err := c.get(ctx, "/api/v1/notification/channels/"+id)
	if err != nil {
		return domain.NotificationChannel{}, err
	}
	var ch domain.NotificationChannel
	if err := json.Unmarshal(raw, &ch); err != nil {
		return domain.NotificationChannel{}, fmt.Errorf("apiclient: decode channel: %w", err)
	}
	return ch, nil
}

func (c *Client) get(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("apiclient: build request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apiclient: GET %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var env apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, fmt.Errorf("apiclient: decode response from %s: %w", path, err)
	}
	if env.Error != nil {
		return nil, fmt.Errorf("apiclient: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}
