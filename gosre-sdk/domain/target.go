// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

// TargetType identifies the protocol used to monitor a target.
type TargetType string

const (
	TargetTypeHTTP TargetType = "http"
	TargetTypeTCP  TargetType = "tcp"
	TargetTypeDNS  TargetType = "dns"
	TargetTypeTLS  TargetType = "tls"
)

// Target represents something to monitor: a URL, TCP endpoint, DNS name, or TLS host.
type Target struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     TargetType        `json:"type"`
	Address  string            `json:"address"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
