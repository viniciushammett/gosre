// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import "time"

// DependencyKind describes the nature of a directed dependency between services.
type DependencyKind string

const (
	DependencyKindHTTP     DependencyKind = "http"
	DependencyKindGRPC     DependencyKind = "grpc"
	DependencyKindDatabase DependencyKind = "database"
	DependencyKindQueue    DependencyKind = "queue"
	DependencyKindGeneric  DependencyKind = "generic"
)

// Dependency represents a directed runtime dependency from one Service to another.
type Dependency struct {
	ID              string         `json:"id"`
	SourceServiceID string         `json:"source_service_id"`
	TargetServiceID string         `json:"target_service_id"`
	Kind            DependencyKind `json:"kind"`
	CreatedAt       time.Time      `json:"created_at"`
}
