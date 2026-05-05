// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// Client fetches data from gosre-api over HTTP.
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
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// apiEnvelope is the standard gosre-api response wrapper.
type apiEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *apiErr         `json:"error"`
}

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ListTargets returns all targets from the API.
func (c *Client) ListTargets(ctx context.Context) ([]domain.Target, error) {
	raw, err := c.get(ctx, "/api/v1/targets")
	if err != nil {
		return nil, err
	}
	var targets []domain.Target
	if err := json.Unmarshal(raw, &targets); err != nil {
		return nil, fmt.Errorf("apiclient: decode targets: %w", err)
	}
	return targets, nil
}

// ListResults returns all results from the API.
func (c *Client) ListResults(ctx context.Context) ([]domain.Result, error) {
	raw, err := c.get(ctx, "/api/v1/results")
	if err != nil {
		return nil, err
	}
	var results []domain.Result
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, fmt.Errorf("apiclient: decode results: %w", err)
	}
	return results, nil
}

// ListChecks returns all check configs from the API.
func (c *Client) ListChecks(ctx context.Context) ([]domain.CheckConfig, error) {
	raw, err := c.get(ctx, "/api/v1/checks")
	if err != nil {
		return nil, err
	}
	var checks []domain.CheckConfig
	if err := json.Unmarshal(raw, &checks); err != nil {
		return nil, fmt.Errorf("apiclient: decode checks: %w", err)
	}
	return checks, nil
}

// ListIncidents returns all incidents from the API.
func (c *Client) ListIncidents(ctx context.Context) ([]domain.Incident, error) {
	raw, err := c.get(ctx, "/api/v1/incidents")
	if err != nil {
		return nil, err
	}
	var incidents []domain.Incident
	if err := json.Unmarshal(raw, &incidents); err != nil {
		return nil, fmt.Errorf("apiclient: decode incidents: %w", err)
	}
	return incidents, nil
}

// AgentRecord holds the registration data returned by GET /api/v1/agents.
type AgentRecord struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	Version  string    `json:"version"`
	LastSeen time.Time `json:"last_seen"`
}

// ListAgents returns all registered agents from the API.
func (c *Client) ListAgents(ctx context.Context) ([]AgentRecord, error) {
	raw, err := c.get(ctx, "/api/v1/agents")
	if err != nil {
		return nil, err
	}
	var agents []AgentRecord
	if err := json.Unmarshal(raw, &agents); err != nil {
		return nil, fmt.Errorf("apiclient: decode agents: %w", err)
	}
	return agents, nil
}

// SLORecord holds SLO data returned by GET /api/v1/slos.
type SLORecord struct {
	ID            string  `json:"id"`
	TargetID      string  `json:"target_id"`
	Name          string  `json:"name"`
	Metric        string  `json:"metric"`
	Threshold     float64 `json:"threshold"`
	WindowSeconds int64   `json:"window_seconds"`
}

// BudgetResult holds error budget metrics returned by GET /api/v1/slos/:id/budget.
type BudgetResult struct {
	SLOID            string  `json:"slo_id"`
	TargetID         string  `json:"target_id"`
	Compliance       float64 `json:"compliance"`
	BurnRate1h       float64 `json:"burn_rate_1h"`
	BurnRate6h       float64 `json:"burn_rate_6h"`
	InsufficientData bool    `json:"insufficient_data"`
}

// ListSLOsByTarget returns all SLOs for the given targetID.
func (c *Client) ListSLOsByTarget(ctx context.Context, targetID string) ([]SLORecord, error) {
	raw, err := c.get(ctx, "/api/v1/slos?target_id="+targetID)
	if err != nil {
		return nil, err
	}
	var slos []SLORecord
	if err := json.Unmarshal(raw, &slos); err != nil {
		return nil, fmt.Errorf("apiclient: decode slos: %w", err)
	}
	return slos, nil
}

// GetSLOBudget returns the error budget result for the given sloID.
func (c *Client) GetSLOBudget(ctx context.Context, sloID string) (BudgetResult, error) {
	raw, err := c.get(ctx, "/api/v1/slos/"+sloID+"/budget")
	if err != nil {
		return BudgetResult{}, err
	}
	var result BudgetResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return BudgetResult{}, fmt.Errorf("apiclient: decode budget: %w", err)
	}
	return result, nil
}

func (c *Client) get(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("apiclient: build request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
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
