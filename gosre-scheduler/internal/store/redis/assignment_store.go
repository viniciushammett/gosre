// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package redis implements store.AssignmentStore backed by Redis.
// Each assignment is stored as a JSON string keyed by ID.
// Per-agent membership is tracked via a Redis SET for O(1) listing and bulk deletion.
package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/redis/go-redis/v9"
)

const (
	assignmentPrefix    = "gosre:assignment:"
	agentAssignmentsFmt = "gosre:agent:%s:assignments"
)

// AssignmentStore implements store.AssignmentStore backed by Redis.
type AssignmentStore struct {
	rdb *redis.Client
}

// NewAssignmentStore constructs an AssignmentStore.
func NewAssignmentStore(rdb *redis.Client) *AssignmentStore {
	return &AssignmentStore{rdb: rdb}
}

// Save persists an Assignment and registers its ID in the agent's assignment set.
func (s *AssignmentStore) Save(ctx context.Context, a domain.Assignment) error {
	data, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("redis: marshal assignment %s: %w", a.ID, err)
	}
	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, assignmentPrefix+a.ID, data, 0)
	pipe.SAdd(ctx, fmt.Sprintf(agentAssignmentsFmt, a.AgentID), a.ID)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis: save assignment %s: %w", a.ID, err)
	}
	return nil
}

// Get retrieves an Assignment by ID.
func (s *AssignmentStore) Get(ctx context.Context, id string) (domain.Assignment, error) {
	data, err := s.rdb.Get(ctx, assignmentPrefix+id).Bytes()
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("redis: get assignment %s: %w", id, err)
	}
	var a domain.Assignment
	if err := json.Unmarshal(data, &a); err != nil {
		return domain.Assignment{}, fmt.Errorf("redis: unmarshal assignment %s: %w", id, err)
	}
	return a, nil
}

// ListByAgent returns all assignments currently held by an agent.
// Stale IDs in the membership set (whose keys were deleted externally) are silently skipped.
func (s *AssignmentStore) ListByAgent(ctx context.Context, agentID string) ([]domain.Assignment, error) {
	ids, err := s.rdb.SMembers(ctx, fmt.Sprintf(agentAssignmentsFmt, agentID)).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: list assignments for agent %s: %w", agentID, err)
	}
	out := make([]domain.Assignment, 0, len(ids))
	for _, id := range ids {
		a, err := s.Get(ctx, id)
		if err != nil {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

// DeleteByAgent removes all assignments belonging to an agent in a single pipeline.
func (s *AssignmentStore) DeleteByAgent(ctx context.Context, agentID string) error {
	ids, err := s.rdb.SMembers(ctx, fmt.Sprintf(agentAssignmentsFmt, agentID)).Result()
	if err != nil {
		return fmt.Errorf("redis: list assignments for agent %s: %w", agentID, err)
	}
	if len(ids) == 0 {
		return nil
	}
	pipe := s.rdb.Pipeline()
	for _, id := range ids {
		pipe.Del(ctx, assignmentPrefix+id)
	}
	pipe.Del(ctx, fmt.Sprintf(agentAssignmentsFmt, agentID))
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis: delete assignments for agent %s: %w", agentID, err)
	}
	return nil
}
