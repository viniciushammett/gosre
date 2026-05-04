// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	natsjets "github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/gosre/gosre-agent/internal/check"
	"github.com/gosre/gosre-sdk/domain"
	"github.com/viniciushammett/gosre/gosre-events/events"
	gosrejs "github.com/viniciushammett/gosre/gosre-events/jetstream"
)

type Agent struct {
	id         string
	hostname   string
	apiURL     string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
	checkers   map[domain.CheckType]domain.Checker
	pub        *gosrejs.Publisher
}

func New(apiURL, apiKey string, logger *zap.Logger) *Agent {
	return &Agent{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
		checkers: map[domain.CheckType]domain.Checker{
			domain.CheckTypeHTTP: check.NewHTTPChecker(),
			domain.CheckTypeTCP:  check.NewTCPChecker(),
			domain.CheckTypeDNS:  check.NewDNSChecker(),
			domain.CheckTypeTLS:  check.NewTLSChecker(),
		},
	}
}

func (a *Agent) ID() string { return a.id }

type registerRequest struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

type registerResponse struct {
	Data  *agentData `json:"data"`
	Error *apiError  `json:"error"`
}

type agentData struct {
	ID string `json:"id"`
}

type apiError struct {
	Message string `json:"message"`
}

func (a *Agent) Register(ctx context.Context) error {
	hostname, _ := os.Hostname()
	body, err := json.Marshal(registerRequest{Hostname: hostname, Version: "0.1.0"})
	if err != nil {
		return fmt.Errorf("marshal register request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.apiURL+"/api/v1/agents/register", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("X-API-Key", a.apiKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("register request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var out registerResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("decode register response: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		msg := "unknown error"
		if out.Error != nil {
			msg = out.Error.Message
		}
		return fmt.Errorf("register failed (%d): %s", resp.StatusCode, msg)
	}
	if out.Data == nil || out.Data.ID == "" {
		return fmt.Errorf("register response missing agent id")
	}

	a.id = out.Data.ID
	a.hostname = hostname
	a.logger.Info("agent registered", zap.String("agent_id", a.id), zap.String("hostname", hostname))
	return nil
}

// SetPublisher attaches a NATS publisher for parallel heartbeat events.
// If not called the agent operates in HTTP-only mode.
func (a *Agent) SetPublisher(pub *gosrejs.Publisher) {
	a.pub = pub
}

func (a *Agent) Heartbeat(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.sendHTTPHeartbeat(ctx)
			a.sendNATSHeartbeat(ctx)
		}
	}
}

func (a *Agent) sendHTTPHeartbeat(ctx context.Context) {
	url := fmt.Sprintf("%s/api/v1/agents/%s/heartbeat", a.apiURL, a.id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		a.logger.Warn("heartbeat request build failed", zap.Error(err))
		return
	}
	if a.apiKey != "" {
		req.Header.Set("X-API-Key", a.apiKey)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.logger.Warn("heartbeat failed", zap.Error(err))
		return
	}
	_ = resp.Body.Close()
	a.logger.Debug("heartbeat sent", zap.String("agent_id", a.id))
}

// sendNATSHeartbeat publishes gosre.agents.heartbeat when a NATS publisher is set.
// Failure is non-fatal — HTTP heartbeat is the authoritative liveness signal.
func (a *Agent) sendNATSHeartbeat(ctx context.Context) {
	if a.pub == nil {
		return
	}
	payload := events.AgentHeartbeatPayload{
		EventEnvelope: events.NewEnvelope(),
		AgentID:       a.id,
		Hostname:      a.hostname,
		Version:       "0.1.0",
	}
	if err := a.pub.Publish(ctx, events.SubjectAgentsHeartbeat, payload); err != nil {
		a.logger.Warn("heartbeat nats publish failed", zap.Error(err))
	}
}

type assignmentsResponse struct {
	Data  []assignment `json:"data"`
	Error *apiError    `json:"error"`
}

type assignment struct {
	CheckID  string           `json:"check_id"`
	TargetID string           `json:"target_id"`
	Type     domain.CheckType `json:"type"`
}

type targetEnvelope struct {
	Data  *domain.Target `json:"data"`
	Error *apiError      `json:"error"`
}

func (a *Agent) fetchTarget(ctx context.Context, targetID string) (domain.Target, error) {
	url := fmt.Sprintf("%s/api/v1/targets/%s", a.apiURL, targetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Target{}, fmt.Errorf("build target request: %w", err)
	}
	if a.apiKey != "" {
		req.Header.Set("X-API-Key", a.apiKey)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return domain.Target{}, fmt.Errorf("fetch target: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var env targetEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return domain.Target{}, fmt.Errorf("decode target: %w", err)
	}
	if env.Data == nil {
		return domain.Target{}, fmt.Errorf("target %s not found", targetID)
	}
	return *env.Data, nil
}

func (a *Agent) postResult(ctx context.Context, r domain.Result) error {
	r.AgentID = a.id

	// API expects duration_ms as int64 milliseconds, not nanoseconds
	type resultPayload struct {
		ID         string             `json:"id"`
		CheckID    string             `json:"check_id"`
		TargetID   string             `json:"target_id"`
		AgentID    string             `json:"agent_id"`
		Status     domain.CheckStatus `json:"status"`
		DurationMs int64              `json:"duration_ms"`
		Error      string             `json:"error,omitempty"`
		Timestamp  time.Time          `json:"timestamp"`
	}

	payload := resultPayload{
		ID:         r.ID,
		CheckID:    r.CheckID,
		TargetID:   r.TargetID,
		AgentID:    r.AgentID,
		Status:     r.Status,
		DurationMs: r.Duration.Milliseconds(),
		Error:      r.Error,
		Timestamp:  r.Timestamp,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.apiURL+"/api/v1/results", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build result request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("X-API-Key", a.apiKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post result: %w", err)
	}
	_ = resp.Body.Close()
	return nil
}

func (a *Agent) executeAssignment(ctx context.Context, asgn assignment) {
	target, err := a.fetchTarget(ctx, asgn.TargetID)
	if err != nil {
		a.logger.Warn("fetch target failed", zap.String("target_id", asgn.TargetID), zap.Error(err))
		return
	}

	checker, ok := a.checkers[asgn.Type]
	if !ok {
		a.logger.Warn("no checker for type", zap.String("type", string(asgn.Type)))
		return
	}

	cfg := domain.CheckConfig{
		ID:       asgn.CheckID,
		Type:     asgn.Type,
		TargetID: asgn.TargetID,
		Timeout:  10 * time.Second,
	}

	result := checker.Execute(ctx, target, cfg)

	if err := a.postResult(ctx, result); err != nil {
		a.logger.Warn("post result failed", zap.String("check_id", asgn.CheckID), zap.Error(err))
		return
	}

	a.logger.Info("check executed",
		zap.String("check_id", asgn.CheckID),
		zap.String("target", target.Address),
		zap.String("status", string(result.Status)),
		zap.Int64("duration_ms", result.Duration.Milliseconds()),
	)
}

// SubscribeAssignments subscribes to gosre.checks.assigned on NATS JetStream.
// Messages destined for other agents are acknowledged and skipped.
// This is a push complement to the HTTP poll loop — both run concurrently.
// Must be called after Register so that a.id is known.
func (a *Agent) SubscribeAssignments(ctx context.Context, js natsjets.JetStream) error {
	durableName := "agent-" + a.id
	return gosrejs.Subscribe(ctx, js, events.SubjectChecksAssigned, durableName,
		func(ctx context.Context, data []byte) error {
			var payload events.CheckAssignedPayload
			if err := json.Unmarshal(data, &payload); err != nil {
				return fmt.Errorf("decode check assigned: %w", err)
			}
			if payload.AgentID != a.id {
				return nil // not for this agent — ack and discard
			}
			go a.executeAssignment(ctx, assignment{
				CheckID:  payload.CheckID,
				TargetID: payload.TargetID,
				Type:     domain.CheckType(payload.CheckType),
			})
			return nil
		},
	)
}

func (a *Agent) Run(ctx context.Context, pollInterval time.Duration) {
	go a.Heartbeat(ctx)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			a.logger.Info("agent stopping", zap.String("agent_id", a.id))
			return
		case <-ticker.C:
			url := fmt.Sprintf("%s/api/v1/agents/%s/assignments", a.apiURL, a.id)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				a.logger.Warn("assignments request build failed", zap.Error(err))
				continue
			}
			if a.apiKey != "" {
				req.Header.Set("X-API-Key", a.apiKey)
			}
			resp, err := a.httpClient.Do(req)
			if err != nil {
				a.logger.Warn("poll assignments failed", zap.Error(err))
				continue
			}
			var out assignmentsResponse
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				_ = resp.Body.Close()
				a.logger.Warn("decode assignments failed", zap.Error(err))
				continue
			}
			_ = resp.Body.Close()

			for _, asgn := range out.Data {
				go a.executeAssignment(ctx, asgn)
			}
		}
	}
}
