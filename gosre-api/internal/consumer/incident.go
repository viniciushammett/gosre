// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	natsjets "github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/viniciushammett/gosre/gosre-api/internal/service"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"
	"github.com/viniciushammett/gosre/gosre-events/events"
	gosrejs "github.com/viniciushammett/gosre/gosre-events/jetstream"
)

// StartIncidentDetector subscribes to gosre.results.created and runs incident
// detection asynchronously for every persisted result. It replaces the inline
// incidentSvc.Process call that was previously inside CheckService.Run.
func StartIncidentDetector(ctx context.Context, js natsjets.JetStream, incidentSvc *service.IncidentService, logger *zap.Logger) error {
	return gosrejs.Subscribe(ctx, js, events.SubjectResultsCreated, "incident-detector",
		func(ctx context.Context, data []byte) error {
			var payload events.ResultCreatedPayload
			if err := json.Unmarshal(data, &payload); err != nil {
				logger.Error("incident-detector: unmarshal", zap.Error(err))
				return fmt.Errorf("unmarshal result payload: %w", err)
			}
			r := domain.Result{
				ID:         payload.ResultID,
				CheckID:    payload.CheckID,
				TargetID:   payload.TargetID,
				TargetName: payload.TargetName,
				AgentID:    payload.AgentID,
				Status:     domain.CheckStatus(payload.Status),
				Duration:   time.Duration(payload.DurationMS) * time.Millisecond,
				Error:      payload.Error,
				Timestamp:  payload.Timestamp,
				Metadata:   payload.Metadata,
			}
			if err := incidentSvc.Process(ctx, r); err != nil {
				logger.Error("incident-detector: process", zap.Error(err))
				return err
			}
			return nil
		},
	)
}
