// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package channel

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/gosre/gosre-sdk/domain"
)

// EmailSender sends notifications via SMTP using net/smtp.
// Config keys: "smtp_host", "smtp_port", "smtp_from", "smtp_username", "smtp_password", "to".
type EmailSender struct{}

// Send delivers msg via the SMTP server configured in ch.Config.
func (s *EmailSender) Send(_ context.Context, ch domain.NotificationChannel, msg Message) error {
	host := ch.Config["smtp_host"]
	port := ch.Config["smtp_port"]
	from := ch.Config["smtp_from"]
	username := ch.Config["smtp_username"]
	password := ch.Config["smtp_password"]
	to := ch.Config["to"]

	if host == "" || from == "" || to == "" {
		return fmt.Errorf("email: channel %q missing smtp_host, smtp_from, or to", ch.ID)
	}
	if port == "" {
		port = "587"
	}

	addr := host + ":" + port
	recipients := strings.Split(to, ",")

	headers := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + msg.Subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n"
	body := []byte(headers + msg.Body + "\r\n")

	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	if err := smtp.SendMail(addr, auth, from, recipients, body); err != nil {
		return fmt.Errorf("email: send via %s: %w", addr, err)
	}
	return nil
}
