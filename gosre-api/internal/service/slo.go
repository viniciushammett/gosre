// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/gosre/gosre-sdk/store"
)

// BudgetResult holds the calculated error budget metrics for a given SLO.
// InsufficientData is true when the window has fewer than 1000 results (L-030).
type BudgetResult struct {
	SLOID            string  `json:"slo_id"`
	TargetID         string  `json:"target_id"`
	Compliance       float64 `json:"compliance"`
	BurnRate1h       float64 `json:"burn_rate_1h"`
	BurnRate6h       float64 `json:"burn_rate_6h"`
	BurnRate24h      float64 `json:"burn_rate_24h"`
	InsufficientData bool    `json:"insufficient_data"`
	TotalResults     int     `json:"total_results"`
}

const insufficientDataThreshold = 1000

// SLOService handles business logic for SLO entities and error budget calculations.
type SLOService struct {
	slos    store.SLOStore
	results store.ResultStore
}

// NewSLOService constructs a SLOService backed by the given stores.
func NewSLOService(slos store.SLOStore, results store.ResultStore) *SLOService {
	return &SLOService{slos: slos, results: results}
}

// Create validates, assigns a UUID if absent, and persists a new SLO.
func (svc *SLOService) Create(ctx context.Context, s domain.SLO) (domain.SLO, error) {
	if s.Name == "" {
		return domain.SLO{}, fmt.Errorf("name is required")
	}
	if s.TargetID == "" {
		return domain.SLO{}, fmt.Errorf("target_id is required")
	}
	if s.Metric == "" {
		return domain.SLO{}, fmt.Errorf("metric is required")
	}
	if s.Threshold <= 0 || s.Threshold >= 1 {
		return domain.SLO{}, fmt.Errorf("threshold must be between 0 and 1 exclusive")
	}
	if s.WindowSeconds <= 0 {
		return domain.SLO{}, fmt.Errorf("window_seconds must be positive")
	}
	if s.ID == "" {
		s.ID = newSLOID()
	}
	if err := svc.slos.Save(ctx, s); err != nil {
		return domain.SLO{}, fmt.Errorf("create slo: %w", err)
	}
	return s, nil
}

// Get retrieves a SLO by ID.
func (svc *SLOService) Get(ctx context.Context, id string) (domain.SLO, error) {
	return svc.slos.Get(ctx, id)
}

// ListByTarget returns all SLOs for the given target.
func (svc *SLOService) ListByTarget(ctx context.Context, targetID string) ([]domain.SLO, error) {
	return svc.slos.ListByTarget(ctx, targetID)
}

// Delete removes a SLO by ID.
func (svc *SLOService) Delete(ctx context.Context, id string) error {
	return svc.slos.Delete(ctx, id)
}

// Budget calculates error budget and burn rates for the given SLO.
// Returns InsufficientData=true when the window has fewer than 1000 results (L-030).
// BurnRate = actual_failure_rate / (1 - threshold). Values > 1 mean budget exhaustion.
func (svc *SLOService) Budget(ctx context.Context, sloID string) (BudgetResult, error) {
	slo, err := svc.slos.Get(ctx, sloID)
	if err != nil {
		return BudgetResult{}, fmt.Errorf("get slo: %w", err)
	}

	all, err := svc.results.ListByTarget(ctx, slo.TargetID)
	if err != nil {
		return BudgetResult{}, fmt.Errorf("list results for target %s: %w", slo.TargetID, err)
	}

	now := time.Now()
	window := time.Duration(slo.WindowSeconds) * time.Second
	windowResults := resultsSince(all, now.Add(-window))

	res := BudgetResult{
		SLOID:        slo.ID,
		TargetID:     slo.TargetID,
		TotalResults: len(windowResults),
	}

	if len(windowResults) < insufficientDataThreshold {
		res.InsufficientData = true
		return res, nil
	}

	failures := countFailures(windowResults)
	res.Compliance = 1 - float64(failures)/float64(len(windowResults))
	res.BurnRate1h = burnRate(all, now, time.Hour, slo.Threshold)
	res.BurnRate6h = burnRate(all, now, 6*time.Hour, slo.Threshold)
	res.BurnRate24h = burnRate(all, now, 24*time.Hour, slo.Threshold)

	return res, nil
}

// resultsSince filters results to those with Timestamp >= since.
func resultsSince(results []domain.Result, since time.Time) []domain.Result {
	out := make([]domain.Result, 0, len(results))
	for _, r := range results {
		if !r.Timestamp.Before(since) {
			out = append(out, r)
		}
	}
	return out
}

// countFailures counts results where Status != StatusOK.
func countFailures(results []domain.Result) int {
	n := 0
	for _, r := range results {
		if r.Status != domain.StatusOK {
			n++
		}
	}
	return n
}

// burnRate computes (failure_rate_in_window) / (1 - threshold).
// Returns 0 when the sub-window has no results or error budget is zero.
func burnRate(all []domain.Result, now time.Time, window time.Duration, threshold float64) float64 {
	inWindow := resultsSince(all, now.Add(-window))
	if len(inWindow) == 0 {
		return 0
	}
	errorBudget := 1 - threshold
	if errorBudget == 0 {
		return 0
	}
	return (float64(countFailures(inWindow)) / float64(len(inWindow))) / errorBudget
}

// newSLOID returns a random UUID v4 using crypto/rand.
func newSLOID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("gosre-api: crypto/rand unavailable: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
