// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package channel defines the Sender interface and delivery implementations.
package channel

import (
	"context"

	"github.com/gosre/gosre-sdk/domain"
)

// Message holds the human-readable content sent to a notification channel.
type Message struct {
	Subject string
	Body    string
}

// Sender delivers a notification to a configured channel.
type Sender interface {
	Send(ctx context.Context, ch domain.NotificationChannel, msg Message) error
}
