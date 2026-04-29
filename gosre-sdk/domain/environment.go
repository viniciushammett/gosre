// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import "time"

// EnvironmentKind classifies an environment by its role in the delivery pipeline.
type EnvironmentKind string

const (
	EnvironmentKindDev     EnvironmentKind = "dev"
	EnvironmentKindStaging EnvironmentKind = "staging"
	EnvironmentKindProd    EnvironmentKind = "prod"
)

// Environment represents a deployment target (dev, staging, prod) scoped to a Project.
type Environment struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	ProjectID string          `json:"project_id"`
	Kind      EnvironmentKind `json:"kind"`
	CreatedAt time.Time       `json:"created_at"`
}
