// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

import "errors"

var (
	ErrTargetNotFound   = errors.New("target not found")
	ErrCheckTimeout     = errors.New("check timed out")
	ErrIncidentNotFound = errors.New("incident not found")
)
