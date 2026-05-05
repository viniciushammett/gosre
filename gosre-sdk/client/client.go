// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package client provides a typed HTTP client for the gosre-api REST API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client is a typed HTTP client for gosre-api.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New constructs a Client. baseURL is the scheme+host with no trailing slash
// (e.g. "http://48.204.162.193").
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ── Request / response types ──────────────────────────────────────────────────

// HealthzResponse is the response body from GET /healthz.
type HealthzResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Target mirrors the gosre-api target object.
type Target struct {
	ID       string            `json:"id,omitempty"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Address  string            `json:"address"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CheckConfig mirrors the gosre-api check config object.
type CheckConfig struct {
	ID       string            `json:"id,omitempty"`
	Type     string            `json:"type"`
	TargetID string            `json:"target_id"`
	Interval int64             `json:"interval,omitempty"`
	Timeout  int64             `json:"timeout,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
}

// Result mirrors the gosre-api result object.
type Result struct {
	ID         string            `json:"id,omitempty"`
	CheckID    string            `json:"check_id,omitempty"`
	TargetID   string            `json:"target_id,omitempty"`
	AgentID    string            `json:"agent_id,omitempty"`
	Status     string            `json:"status,omitempty"`
	DurationMS int64             `json:"duration_ms,omitempty"`
	Error      string            `json:"error,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	TargetName string            `json:"target_name,omitempty"`
}

// Incident mirrors the gosre-api incident object.
type Incident struct {
	ID        string    `json:"id,omitempty"`
	TargetID  string    `json:"target_id,omitempty"`
	State     string    `json:"state,omitempty"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	ResultIDs []string  `json:"result_ids,omitempty"`
}

// Agent mirrors the gosre-api agent object.
type Agent struct {
	ID       string    `json:"id,omitempty"`
	Hostname string    `json:"hostname,omitempty"`
	Version  string    `json:"version,omitempty"`
	LastSeen time.Time `json:"last_seen"`
}

// Assignment mirrors the gosre-api assignment object.
type Assignment struct {
	CheckID  string `json:"check_id"`
	TargetID string `json:"target_id"`
	Type     string `json:"type"`
}

// CreateTargetRequest is the body for POST /api/v1/targets.
type CreateTargetRequest struct {
	ID       string            `json:"id,omitempty"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Address  string            `json:"address"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CreateCheckRequest is the body for POST /api/v1/checks.
type CreateCheckRequest struct {
	ID       string            `json:"id,omitempty"`
	Type     string            `json:"type"`
	TargetID string            `json:"target_id"`
	Interval int64             `json:"interval,omitempty"`
	Timeout  int64             `json:"timeout,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
}

// PatchIncidentRequest is the body for PATCH /api/v1/incidents/:id.
type PatchIncidentRequest struct {
	State string `json:"state"`
}

// RegisterAgentRequest is the body for POST /api/v1/agents/register.
type RegisterAgentRequest struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

// ListResultsParams holds optional query parameters for ListResults.
// From and To are encoded as RFC3339 strings; zero values are omitted.
type ListResultsParams struct {
	TargetID string
	Status   string
	From     time.Time
	To       time.Time
}

// ListIncidentsParams holds optional query parameters for ListIncidents.
type ListIncidentsParams struct {
	State string
}

// SLO mirrors the gosre-api SLO object.
// WindowSeconds is the observation window expressed in seconds (not time.Duration).
type SLO struct {
	ID            string  `json:"id,omitempty"`
	TargetID      string  `json:"target_id"`
	Name          string  `json:"name"`
	Metric        string  `json:"metric"`
	Threshold     float64 `json:"threshold"`
	WindowSeconds int64   `json:"window_seconds"`
}

// SLOBudget mirrors the error-budget calculation returned by GET /api/v1/slos/:id/budget.
type SLOBudget struct {
	SLOID            string  `json:"slo_id"`
	TargetID         string  `json:"target_id"`
	Compliance       float64 `json:"compliance"`
	BurnRate1h       float64 `json:"burn_rate_1h"`
	BurnRate6h       float64 `json:"burn_rate_6h"`
	BurnRate24h      float64 `json:"burn_rate_24h"`
	InsufficientData bool    `json:"insufficient_data"`
	TotalResults     int     `json:"total_results"`
}

// CreateSLORequest is the body for POST /api/v1/slos.
type CreateSLORequest struct {
	TargetID      string  `json:"target_id"`
	Name          string  `json:"name"`
	Metric        string  `json:"metric"`
	Threshold     float64 `json:"threshold"`
	WindowSeconds int64   `json:"window_seconds"`
}

// ListSLOsParams holds optional query parameters for ListSLOs.
type ListSLOsParams struct {
	TargetID string
}

// Organization mirrors the gosre-api organization object.
type Organization struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateOrgRequest is the body for POST /api/v1/organizations.
type CreateOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// Team mirrors the gosre-api team object.
type Team struct {
	ID             string    `json:"id,omitempty"`
	OrganizationID string    `json:"organization_id,omitempty"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateTeamRequest is the body for POST /api/v1/organizations/:org_id/teams.
type CreateTeamRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// Project mirrors the gosre-api project object.
type Project struct {
	ID             string    `json:"id,omitempty"`
	OrganizationID string    `json:"organization_id,omitempty"`
	TeamID         string    `json:"team_id,omitempty"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateProjectRequest is the body for POST /api/v1/organizations/:org_id/projects.
type CreateProjectRequest struct {
	Name   string `json:"name"`
	Slug   string `json:"slug,omitempty"`
	TeamID string `json:"team_id,omitempty"`
}

// SchedulerStatus mirrors the response from GET /api/v1/scheduler/status.
type SchedulerStatus struct {
	AgentsActive        int            `json:"agents_active"`
	AgentsDown          int            `json:"agents_down"`
	TotalChecks         int            `json:"total_checks"`
	AssignmentsPerAgent map[string]int `json:"assignments_per_agent"`
}

// ── Internal envelope and helpers ─────────────────────────────────────────────

type envelope[T any] struct {
	Data  T         `json:"data"`
	Error *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

const (
	maxAttempts = 3
	backoffBase = 100 * time.Millisecond
)

// isRetriable returns true when a response warrants an automatic retry.
// Only 5xx status codes and network-level timeouts are retried; client errors
// and context cancellation are not.
func isRetriable(status int, err error) bool {
	if status >= 500 {
		return true
	}
	var netErr net.Error
	return err != nil && errors.As(err, &netErr) && netErr.Timeout()
}

// do executes an HTTP request with exponential-backoff retry on 5xx and
// network timeouts. The caller is responsible for closing the response body.
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	var (
		resp *http.Response
		err  error
	)
	for attempt := range maxAttempts {
		if attempt > 0 {
			wait := backoffBase * (1 << (attempt - 1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
		if reqErr != nil {
			return nil, fmt.Errorf("build request: %w", reqErr)
		}
		if c.apiKey != "" {
			req.Header.Set("X-API-Key", c.apiKey)
		}
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err = c.httpClient.Do(req)

		if err != nil {
			if !isRetriable(0, err) {
				return nil, err
			}
			// retriable network error — try again unless this is the last attempt
			if attempt < maxAttempts-1 {
				continue
			}
			return nil, err
		}
		if !isRetriable(resp.StatusCode, nil) {
			return resp, nil
		}
		// 5xx — close body before retry unless this is the last attempt
		if attempt < maxAttempts-1 {
			resp.Body.Close() //nolint:errcheck
			resp = nil
		}
	}
	return resp, nil
}

// decode reads the standard API envelope and returns the data or an API error.
func decode[T any](resp *http.Response) (T, error) {
	defer resp.Body.Close() //nolint:errcheck
	var env envelope[T]
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		var zero T
		return zero, fmt.Errorf("decode response: %w", err)
	}
	if env.Error != nil {
		var zero T
		return zero, env.Error
	}
	return env.Data, nil
}

// expectNoContent verifies the response carries a 204 status. On any other
// status it attempts to decode an API error from the body.
func expectNoContent(resp *http.Response) error {
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	var env envelope[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&env); err == nil && env.Error != nil {
		return env.Error
	}
	return fmt.Errorf("unexpected status %d", resp.StatusCode)
}

// ── Endpoint methods ──────────────────────────────────────────────────────────

// Healthz calls GET /healthz.
// The healthz response is a plain JSON object, not the standard API envelope.
func (c *Client) Healthz(ctx context.Context) (*HealthzResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, "/healthz", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	var out HealthzResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode healthz response: %w", err)
	}
	return &out, nil
}

// ListTargets calls GET /api/v1/targets.
func (c *Client) ListTargets(ctx context.Context) ([]Target, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/targets", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Target](resp)
}

// CreateTarget calls POST /api/v1/targets.
func (c *Client) CreateTarget(ctx context.Context, req CreateTargetRequest) (*Target, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/targets", req)
	if err != nil {
		return nil, err
	}
	t, err := decode[Target](resp)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTarget calls GET /api/v1/targets/:id.
func (c *Client) GetTarget(ctx context.Context, id string) (*Target, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/targets/"+id, nil)
	if err != nil {
		return nil, err
	}
	t, err := decode[Target](resp)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// DeleteTarget calls DELETE /api/v1/targets/:id.
func (c *Client) DeleteTarget(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/api/v1/targets/"+id, nil)
	if err != nil {
		return err
	}
	return expectNoContent(resp)
}

// ListChecks calls GET /api/v1/checks.
func (c *Client) ListChecks(ctx context.Context) ([]CheckConfig, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/checks", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]CheckConfig](resp)
}

// CreateCheck calls POST /api/v1/checks.
func (c *Client) CreateCheck(ctx context.Context, req CreateCheckRequest) (*CheckConfig, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/checks", req)
	if err != nil {
		return nil, err
	}
	ch, err := decode[CheckConfig](resp)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// RunCheck calls POST /api/v1/checks/:id/run.
func (c *Client) RunCheck(ctx context.Context, id string) (*Result, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/checks/"+id+"/run", nil)
	if err != nil {
		return nil, err
	}
	r, err := decode[Result](resp)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ListResults calls GET /api/v1/results.
// Pass a zero-value ListResultsParams to return all results.
func (c *Client) ListResults(ctx context.Context, params ListResultsParams) ([]Result, error) {
	q := url.Values{}
	if params.TargetID != "" {
		q.Set("target_id", params.TargetID)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if !params.From.IsZero() {
		q.Set("from", params.From.Format(time.RFC3339))
	}
	if !params.To.IsZero() {
		q.Set("to", params.To.Format(time.RFC3339))
	}
	path := "/api/v1/results"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Result](resp)
}

// GetResult calls GET /api/v1/results/:id.
func (c *Client) GetResult(ctx context.Context, id string) (*Result, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/results/"+id, nil)
	if err != nil {
		return nil, err
	}
	r, err := decode[Result](resp)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ListIncidents calls GET /api/v1/incidents.
// Pass a zero-value ListIncidentsParams to return all incidents.
func (c *Client) ListIncidents(ctx context.Context, params ListIncidentsParams) ([]Incident, error) {
	path := "/api/v1/incidents"
	if params.State != "" {
		path += "?" + url.Values{"state": {params.State}}.Encode()
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Incident](resp)
}

// PatchIncident calls PATCH /api/v1/incidents/:id.
func (c *Client) PatchIncident(ctx context.Context, id string, req PatchIncidentRequest) (*Incident, error) {
	resp, err := c.do(ctx, http.MethodPatch, "/api/v1/incidents/"+id, req)
	if err != nil {
		return nil, err
	}
	inc, err := decode[Incident](resp)
	if err != nil {
		return nil, err
	}
	return &inc, nil
}

// ListAgents calls GET /api/v1/agents.
func (c *Client) ListAgents(ctx context.Context) ([]Agent, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/agents", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Agent](resp)
}

// RegisterAgent calls POST /api/v1/agents/register.
func (c *Client) RegisterAgent(ctx context.Context, req RegisterAgentRequest) (*Agent, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/agents/register", req)
	if err != nil {
		return nil, err
	}
	a, err := decode[Agent](resp)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// GetAgentAssignments calls GET /api/v1/agents/:id/assignments.
func (c *Client) GetAgentAssignments(ctx context.Context, id string) ([]Assignment, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/agents/"+id+"/assignments", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Assignment](resp)
}

// AgentHeartbeat calls POST /api/v1/agents/:id/heartbeat.
func (c *Client) AgentHeartbeat(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/agents/"+id+"/heartbeat", nil)
	if err != nil {
		return err
	}
	return expectNoContent(resp)
}

// ── SLOs ──────────────────────────────────────────────────────────────────────

// ListSLOs calls GET /api/v1/slos.
// Pass a zero-value ListSLOsParams to return all SLOs.
func (c *Client) ListSLOs(ctx context.Context, params ListSLOsParams) ([]SLO, error) {
	path := "/api/v1/slos"
	if params.TargetID != "" {
		path += "?" + url.Values{"target_id": {params.TargetID}}.Encode()
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]SLO](resp)
}

// CreateSLO calls POST /api/v1/slos.
func (c *Client) CreateSLO(ctx context.Context, req CreateSLORequest) (*SLO, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/slos", req)
	if err != nil {
		return nil, err
	}
	s, err := decode[SLO](resp)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSLO calls GET /api/v1/slos/:id.
func (c *Client) GetSLO(ctx context.Context, id string) (*SLO, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/slos/"+id, nil)
	if err != nil {
		return nil, err
	}
	s, err := decode[SLO](resp)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteSLO calls DELETE /api/v1/slos/:id.
func (c *Client) DeleteSLO(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/api/v1/slos/"+id, nil)
	if err != nil {
		return err
	}
	return expectNoContent(resp)
}

// GetSLOBudget calls GET /api/v1/slos/:id/budget.
func (c *Client) GetSLOBudget(ctx context.Context, id string) (*SLOBudget, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/slos/"+id+"/budget", nil)
	if err != nil {
		return nil, err
	}
	b, err := decode[SLOBudget](resp)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ── Organizations ─────────────────────────────────────────────────────────────

// ListOrganizations calls GET /api/v1/organizations.
func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/organizations", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Organization](resp)
}

// CreateOrganization calls POST /api/v1/organizations.
func (c *Client) CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/organizations", req)
	if err != nil {
		return nil, err
	}
	o, err := decode[Organization](resp)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// GetOrganization calls GET /api/v1/organizations/:org_id.
func (c *Client) GetOrganization(ctx context.Context, orgID string) (*Organization, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID, nil)
	if err != nil {
		return nil, err
	}
	o, err := decode[Organization](resp)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// DeleteOrganization calls DELETE /api/v1/organizations/:org_id.
func (c *Client) DeleteOrganization(ctx context.Context, orgID string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/api/v1/organizations/"+orgID, nil)
	if err != nil {
		return err
	}
	return expectNoContent(resp)
}

// ── Teams ─────────────────────────────────────────────────────────────────────

// ListTeams calls GET /api/v1/organizations/:org_id/teams.
func (c *Client) ListTeams(ctx context.Context, orgID string) ([]Team, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/teams", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Team](resp)
}

// CreateTeam calls POST /api/v1/organizations/:org_id/teams.
func (c *Client) CreateTeam(ctx context.Context, orgID string, req CreateTeamRequest) (*Team, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/organizations/"+orgID+"/teams", req)
	if err != nil {
		return nil, err
	}
	t, err := decode[Team](resp)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTeam calls GET /api/v1/organizations/:org_id/teams/:id.
func (c *Client) GetTeam(ctx context.Context, orgID, teamID string) (*Team, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/teams/"+teamID, nil)
	if err != nil {
		return nil, err
	}
	t, err := decode[Team](resp)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// DeleteTeam calls DELETE /api/v1/organizations/:org_id/teams/:id.
func (c *Client) DeleteTeam(ctx context.Context, orgID, teamID string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/api/v1/organizations/"+orgID+"/teams/"+teamID, nil)
	if err != nil {
		return err
	}
	return expectNoContent(resp)
}

// ── Projects ──────────────────────────────────────────────────────────────────

// ListProjects calls GET /api/v1/organizations/:org_id/projects.
func (c *Client) ListProjects(ctx context.Context, orgID string) ([]Project, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/projects", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Project](resp)
}

// CreateProject calls POST /api/v1/organizations/:org_id/projects.
func (c *Client) CreateProject(ctx context.Context, orgID string, req CreateProjectRequest) (*Project, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/organizations/"+orgID+"/projects", req)
	if err != nil {
		return nil, err
	}
	p, err := decode[Project](resp)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetProject calls GET /api/v1/organizations/:org_id/projects/:id.
func (c *Client) GetProject(ctx context.Context, orgID, projectID string) (*Project, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/projects/"+projectID, nil)
	if err != nil {
		return nil, err
	}
	p, err := decode[Project](resp)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProject calls DELETE /api/v1/organizations/:org_id/projects/:id.
func (c *Client) DeleteProject(ctx context.Context, orgID, projectID string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/api/v1/organizations/"+orgID+"/projects/"+projectID, nil)
	if err != nil {
		return err
	}
	return expectNoContent(resp)
}

// ── Scheduler ─────────────────────────────────────────────────────────────────

// GetSchedulerStatus calls GET /api/v1/scheduler/status.
func (c *Client) GetSchedulerStatus(ctx context.Context) (*SchedulerStatus, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/scheduler/status", nil)
	if err != nil {
		return nil, err
	}
	s, err := decode[SchedulerStatus](resp)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
