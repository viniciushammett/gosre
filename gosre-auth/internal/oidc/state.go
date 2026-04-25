// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package oidc provides OAuth2 and OIDC authentication providers.
// GitHub uses OAuth2 directly; OIDC-compliant providers (Google, Keycloak)
// use github.com/coreos/go-oidc/v3.
package oidc

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const stateTTL = 5 * time.Minute

// StateStore holds short-lived CSRF states for OAuth2 flows.
// States are single-use: Validate consumes the entry on success.
type StateStore struct {
	mu     sync.Mutex
	states map[string]time.Time
}

// NewStateStore returns an empty StateStore.
func NewStateStore() *StateStore {
	return &StateStore{states: make(map[string]time.Time)}
}

// Generate creates a random state value, stores it with a 5-minute TTL,
// and returns it for use in the OAuth2 authorization URL.
func (s *StateStore) Generate() string {
	state := uuid.NewString()
	s.mu.Lock()
	s.states[state] = time.Now().Add(stateTTL)
	s.mu.Unlock()
	return state
}

// Validate returns true if the state exists and has not expired.
// The entry is removed on first use regardless of validity.
func (s *StateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.states[state]
	if !ok {
		return false
	}
	delete(s.states, state)
	return time.Now().Before(exp)
}
