// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlite

import (
	"context"
	"testing"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(":memory:")
	require.NoError(t, err)
	return s
}

func TestTargetStore(t *testing.T) {
	target := domain.Target{
		ID:      "t1",
		Name:    "my-api",
		Type:    domain.TargetTypeHTTP,
		Address: "https://api.example.com/healthz",
		Tags:    []string{"production", "api"},
		Metadata: map[string]string{
			"team": "platform",
		},
	}

	t.Run("save and get", func(t *testing.T) {
		s := newTestStore(t)
		ctx := context.Background()

		require.NoError(t, s.Save(ctx, target))

		got, err := s.Get(ctx, target.ID)
		require.NoError(t, err)
		assert.Equal(t, target, got)
	})

	t.Run("get not found", func(t *testing.T) {
		s := newTestStore(t)
		ctx := context.Background()

		_, err := s.Get(ctx, "nonexistent")
		assert.ErrorIs(t, err, domain.ErrTargetNotFound)
	})

	t.Run("list empty", func(t *testing.T) {
		s := newTestStore(t)
		ctx := context.Background()

		targets, err := s.List(ctx)
		require.NoError(t, err)
		assert.NotNil(t, targets)
		assert.Empty(t, targets)
	})

	t.Run("list returns all", func(t *testing.T) {
		s := newTestStore(t)
		ctx := context.Background()

		second := domain.Target{
			ID:       "t2",
			Name:     "db",
			Type:     domain.TargetTypeTCP,
			Address:  "db.example.com:5432",
			Tags:     []string{},
			Metadata: map[string]string{},
		}

		require.NoError(t, s.Save(ctx, target))
		require.NoError(t, s.Save(ctx, second))

		targets, err := s.List(ctx)
		require.NoError(t, err)
		assert.Len(t, targets, 2)
	})

	t.Run("delete existing", func(t *testing.T) {
		s := newTestStore(t)
		ctx := context.Background()

		require.NoError(t, s.Save(ctx, target))
		require.NoError(t, s.Delete(ctx, target.ID))

		_, err := s.Get(ctx, target.ID)
		assert.ErrorIs(t, err, domain.ErrTargetNotFound)
	})

	t.Run("delete not found", func(t *testing.T) {
		s := newTestStore(t)
		ctx := context.Background()

		err := s.Delete(ctx, "nonexistent")
		assert.ErrorIs(t, err, domain.ErrTargetNotFound)
	})

	t.Run("save replaces existing", func(t *testing.T) {
		s := newTestStore(t)
		ctx := context.Background()

		require.NoError(t, s.Save(ctx, target))

		updated := target
		updated.Name = "my-api-updated"
		updated.Address = "https://api2.example.com/healthz"
		require.NoError(t, s.Save(ctx, updated))

		got, err := s.Get(ctx, target.ID)
		require.NoError(t, err)
		assert.Equal(t, "my-api-updated", got.Name)
		assert.Equal(t, "https://api2.example.com/healthz", got.Address)
	})
}
