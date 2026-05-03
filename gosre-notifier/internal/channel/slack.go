// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gosre/gosre-sdk/domain"
)

// SlackSender sends notifications to a Slack incoming webhook.
// Config key: "webhook_url".
type SlackSender struct{}

type slackPayload struct {
	Text string `json:"text"`
}

// Send posts msg to the Slack webhook URL configured in ch.Config["webhook_url"].
func (s *SlackSender) Send(ctx context.Context, ch domain.NotificationChannel, msg Message) error {
	webhookURL, ok := ch.Config["webhook_url"]
	if !ok || webhookURL == "" {
		return fmt.Errorf("slack: channel %q missing webhook_url", ch.ID)
	}

	payload := slackPayload{Text: "*" + msg.Subject + "*\n" + msg.Body}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack: POST webhook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}
