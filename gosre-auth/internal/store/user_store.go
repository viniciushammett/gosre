// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package store

import (
	"context"
	"errors"
	"sync"

	"github.com/gosre/gosre-auth/internal/domain"
)

var (
	// ErrUserNotFound is returned when a lookup finds no matching user.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailTaken is returned when registering an email that already exists.
	ErrEmailTaken = errors.New("email already taken")
)

// UserStore is the persistence contract for user identity records.
type UserStore interface {
	Create(ctx context.Context, u domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
}

// MemoryStore is an in-memory UserStore used in tests.
type MemoryStore struct {
	mu      sync.RWMutex
	users   map[string]domain.User
	byEmail map[string]string // email → id
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:   make(map[string]domain.User),
		byEmail: make(map[string]string),
	}
}

// Create persists a new user. Returns ErrEmailTaken if the email is already registered.
func (m *MemoryStore) Create(_ context.Context, u domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byEmail[u.Email]; ok {
		return ErrEmailTaken
	}
	m.users[u.ID] = u
	m.byEmail[u.Email] = u.ID
	return nil
}

// GetByEmail returns the user with the given email or ErrUserNotFound.
func (m *MemoryStore) GetByEmail(_ context.Context, email string) (domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.byEmail[email]
	if !ok {
		return domain.User{}, ErrUserNotFound
	}
	return m.users[id], nil
}

// GetByID returns the user with the given ID or ErrUserNotFound.
func (m *MemoryStore) GetByID(_ context.Context, id string) (domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return domain.User{}, ErrUserNotFound
	}
	return u, nil
}
