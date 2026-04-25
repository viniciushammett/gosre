// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrSessionNotFound is returned when a refresh token is not found or has expired.
var ErrSessionNotFound = errors.New("session not found")

// SessionStore persists refresh tokens mapped to user IDs.
type SessionStore interface {
	Save(ctx context.Context, refreshToken, userID string, ttl time.Duration) error
	Get(ctx context.Context, refreshToken string) (userID string, err error)
	Delete(ctx context.Context, refreshToken string) error
}

// RedisSessionStore stores sessions in Redis.
type RedisSessionStore struct {
	client *redis.Client
}

// NewRedisSessionStore parses redisURL and returns a store ready to use.
func NewRedisSessionStore(redisURL string) (*RedisSessionStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}
	return &RedisSessionStore{client: redis.NewClient(opts)}, nil
}

// Ping verifies the Redis connection. Used at startup to decide whether to fall back.
func (r *RedisSessionStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Save stores a refresh token → userID mapping with the given TTL.
func (r *RedisSessionStore) Save(ctx context.Context, refreshToken, userID string, ttl time.Duration) error {
	return r.client.Set(ctx, refreshToken, userID, ttl).Err()
}

// Get retrieves the userID for the given refresh token.
// Returns ErrSessionNotFound if the token is missing or expired.
func (r *RedisSessionStore) Get(ctx context.Context, refreshToken string) (string, error) {
	val, err := r.client.Get(ctx, refreshToken).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrSessionNotFound
	}
	if err != nil {
		return "", fmt.Errorf("redis get: %w", err)
	}
	return val, nil
}

// Delete removes a refresh token from Redis.
func (r *RedisSessionStore) Delete(ctx context.Context, refreshToken string) error {
	return r.client.Del(ctx, refreshToken).Err()
}

// MemorySessionStore is an in-memory SessionStore for tests and local dev fallback.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]memoryEntry
}

type memoryEntry struct {
	userID    string
	expiresAt time.Time
}

// NewMemorySessionStore returns an empty MemorySessionStore.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: make(map[string]memoryEntry)}
}

// Save stores a refresh token → userID mapping that expires after ttl.
func (m *MemorySessionStore) Save(_ context.Context, refreshToken, userID string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[refreshToken] = memoryEntry{userID: userID, expiresAt: time.Now().Add(ttl)}
	return nil
}

// Get returns the userID for the token, or ErrSessionNotFound if missing or expired.
func (m *MemorySessionStore) Get(_ context.Context, refreshToken string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.sessions[refreshToken]
	if !ok || time.Now().After(e.expiresAt) {
		return "", ErrSessionNotFound
	}
	return e.userID, nil
}

// Delete removes a refresh token from the store.
func (m *MemorySessionStore) Delete(_ context.Context, refreshToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, refreshToken)
	return nil
}
