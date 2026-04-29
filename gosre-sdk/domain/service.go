// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import "time"

// ServiceCriticality expresses the operational impact of a service going down.
type ServiceCriticality string

const (
	CriticalityLow      ServiceCriticality = "low"
	CriticalityMedium   ServiceCriticality = "medium"
	CriticalityHigh     ServiceCriticality = "high"
	CriticalityCritical ServiceCriticality = "critical"
)

// Service is a catalog entry representing a deployable software component.
type Service struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Owner       string             `json:"owner"`
	Criticality ServiceCriticality `json:"criticality"`
	RunbookURL  string             `json:"runbook_url,omitempty"`
	RepoURL     string             `json:"repo_url,omitempty"`
	ProjectID   string             `json:"project_id,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}
