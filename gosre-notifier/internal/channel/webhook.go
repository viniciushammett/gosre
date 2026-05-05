// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// WebhookSender sends a generic JSON POST to a configured URL.
// Config keys: "url", optionally "authorization" (sent as Authorization header).
type WebhookSender struct{}

type webhookPayload struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Send HTTP POSTs the notification to ch.Config["url"].
func (s *WebhookSender) Send(ctx context.Context, ch domain.NotificationChannel, msg Message) error {
	url, ok := ch.Config["url"]
	if !ok || url == "" {
		return fmt.Errorf("webhook: channel %q missing url", ch.ID)
	}

	payload := webhookPayload{Subject: msg.Subject, Body: msg.Body}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader := ch.Config["authorization"]; authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: POST %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: endpoint returned HTTP %d", resp.StatusCode)
	}
	return nil
}
