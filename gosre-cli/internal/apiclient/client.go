// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"time"

	"github.com/gosre/gosre-sdk/domain"
)

// Client is a minimal HTTP client for gosre-api.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// New constructs a Client for the given API base URL and optional API key.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// apiEnvelope is the standard API response wrapper.
type apiEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *apiErr         `json:"error"`
}

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RunCheck upserts the target and check config in the API, executes the check,
// and returns the persisted Result. IDs are derived from target address and check
// type so that repeated runs update the same records instead of creating new ones.
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
	_, err := c.post(ctx, "/api/v1/targets", t)
	return err
}

func (c *Client) upsertCheck(ctx context.Context, cfg domain.CheckConfig) error {
	_, err := c.post(ctx, "/api/v1/checks", cfg)
	return err
}

func (c *Client) runCheck(ctx context.Context, checkID string) (domain.Result, error) {
	raw, err := c.post(ctx, "/api/v1/checks/"+checkID+"/run", nil)
	if err != nil {
		return domain.Result{}, err
	}
	var r domain.Result
	if err := json.Unmarshal(raw, &r); err != nil {
		return domain.Result{}, fmt.Errorf("apiclient: decode result: %w", err)
	}
	return r, nil
}

func (c *Client) post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("apiclient: marshal body: %w", err)
		}
		bodyBytes = b
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("apiclient: build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apiclient: POST %s: %w", path, err)
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

// stableID returns a URL-safe decimal string derived from the given parts.
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
