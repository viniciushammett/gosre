// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package scheduler implements the distributed check scheduler.
// It uses consistent hashing to assign checks to active agents and publishes
// gosre.checks.assigned to NATS JetStream on every assignment change.
package scheduler

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	sdkclient "github.com/viniciushammett/gosre/gosre-sdk/client"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"
	"github.com/viniciushammett/gosre/gosre-sdk/store"
	"github.com/viniciushammett/gosre/gosre-events/events"
	"github.com/viniciushammett/gosre/gosre-events/jetstream"
	"github.com/viniciushammett/gosre/gosre-scheduler/internal/hash"
)

// agentDownThreshold is the maximum age of a heartbeat before an agent is
// considered down. Reassignment happens only after this window expires to
// avoid flapping on transient network instability (L-031).
const agentDownThreshold = 60 * time.Second

// apiClient is the subset of the gosre-sdk HTTP client used by the scheduler.
// *sdkclient.Client satisfies this interface automatically.
type apiClient interface {
	ListAgents(ctx context.Context) ([]sdkclient.Agent, error)
	ListChecks(ctx context.Context) ([]sdkclient.CheckConfig, error)
}

// Scheduler distributes checks across active agents using consistent hashing
// and republishes gosre.checks.assigned whenever an assignment changes.
type Scheduler struct {
	client          apiClient
	store           store.AssignmentStore
	publisher       *jetstream.Publisher
	interval        time.Duration
	log             *zap.Logger
	prevAssignments map[string]string // checkID → agentID from last reconcile
}

// New constructs a Scheduler.
func New(
	client apiClient,
	store store.AssignmentStore,
	publisher *jetstream.Publisher,
	interval time.Duration,
	log *zap.Logger,
) *Scheduler {
	return &Scheduler{
		client:          client,
		store:           store,
		publisher:       publisher,
		interval:        interval,
		log:             log,
		prevAssignments: make(map[string]string),
	}
}

// Run starts the scheduling loop and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	s.log.Info("scheduler started", zap.Duration("interval", s.interval))
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	if err := s.reconcile(ctx); err != nil {
		s.log.Warn("initial reconcile failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return nil
		case <-ticker.C:
			if err := s.reconcile(ctx); err != nil {
				s.log.Warn("reconcile failed", zap.Error(err))
			}
		}
	}
}

// reconcile fetches all agents and checks, identifies active agents (heartbeat
// within 60s), rebuilds the consistent hash ring, and persists + publishes only
// the assignments that changed since the last run.
func (s *Scheduler) reconcile(ctx context.Context) error {
	agents, err := s.client.ListAgents(ctx)
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}
	checks, err := s.client.ListChecks(ctx)
	if err != nil {
		return fmt.Errorf("list checks: %w", err)
	}

	now := time.Now().UTC()
	active := make([]sdkclient.Agent, 0, len(agents))
	for _, a := range agents {
		if now.Sub(a.LastSeen) <= agentDownThreshold {
			active = append(active, a)
		}
	}

	s.log.Info("reconcile tick",
		zap.Int("agents_total", len(agents)),
		zap.Int("agents_active", len(active)),
		zap.Int("checks", len(checks)),
	)

	if len(active) == 0 || len(checks) == 0 {
		return nil
	}

	// Rebuild ring with only active agents. agent_id is the key — never IP (L-031).
	ring := hash.New()
	for _, a := range active {
		ring.Add(a.ID)
	}

	for _, ch := range checks {
		agentID := ring.Get(ch.ID)
		if agentID == "" {
			continue
		}

		// Skip unchanged assignments to avoid redundant writes and NATS publishes.
		if prev, ok := s.prevAssignments[ch.ID]; ok && prev == agentID {
			continue
		}

		a := domain.Assignment{
			ID:         "asgn-" + ch.ID,
			CheckID:    ch.ID,
			AgentID:    agentID,
			AssignedAt: now,
		}
		if err := s.store.Save(ctx, a); err != nil {
			s.log.Warn("save assignment failed",
				zap.String("check_id", ch.ID),
				zap.Error(err),
			)
			continue
		}
		if err := s.publisher.Publish(ctx, events.SubjectChecksAssigned, a); err != nil {
			s.log.Warn("publish assignment failed",
				zap.String("check_id", ch.ID),
				zap.Error(err),
			)
		}
		s.prevAssignments[ch.ID] = agentID
	}

	return nil
}
