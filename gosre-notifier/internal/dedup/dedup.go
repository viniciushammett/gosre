// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package dedup prevents duplicate notifications for the same incident within a time window.
package dedup

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const dedupTTL = 5 * time.Minute

// Deduplicator uses Redis to gate repeated notifications for the same incident.
type Deduplicator struct {
	rdb *redis.Client
}

// New constructs a Deduplicator backed by the given Redis client.
func New(rdb *redis.Client) *Deduplicator {
	return &Deduplicator{rdb: rdb}
}

// HasBeenNotified returns true if this incident+eventKind was already sent within dedupTTL.
func (d *Deduplicator) HasBeenNotified(ctx context.Context, incidentID, eventKind string) (bool, error) {
	key := "gosre:notifier:dedup:" + eventKind + ":" + incidentID
	n, err := d.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// MarkNotified records that this incident+eventKind has been notified, expiring after dedupTTL.
func (d *Deduplicator) MarkNotified(ctx context.Context, incidentID, eventKind string) error {
	key := "gosre:notifier:dedup:" + eventKind + ":" + incidentID
	return d.rdb.Set(ctx, key, "1", dedupTTL).Err()
}
