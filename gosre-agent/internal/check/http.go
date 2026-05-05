// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// HTTPChecker executes an HTTP GET check against a Target.
type HTTPChecker struct{}

// NewHTTPChecker returns a ready-to-use HTTPChecker.
func NewHTTPChecker() *HTTPChecker {
	return &HTTPChecker{}
}

// Execute performs an HTTP GET to t.Address, respecting ctx and cfg.Timeout.
// It implements domain.Checker.
func (c *HTTPChecker) Execute(ctx context.Context, t domain.Target, cfg domain.CheckConfig) domain.Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.Address, nil)
	if err != nil {
		return domain.Result{
			ID:        fmt.Sprintf("%d", start.UnixNano()),
			CheckID:   cfg.ID,
			TargetID:  t.ID,
			Status:    domain.StatusFail,
			Duration:  time.Since(start),
			Error:     fmt.Errorf("build request: %w", err).Error(),
			Timestamp: time.Now().UTC(),
		}
	}

	client := &http.Client{Timeout: cfg.Timeout}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		status := domain.StatusFail
		var urlErr *url.Error
		if errors.Is(ctx.Err(), context.DeadlineExceeded) ||
			(errors.As(err, &urlErr) && urlErr.Timeout()) {
			status = domain.StatusTimeout
		}
		return domain.Result{
			ID:        fmt.Sprintf("%d", start.UnixNano()),
			CheckID:   cfg.ID,
			TargetID:  t.ID,
			Status:    status,
			Duration:  duration,
			Error:     err.Error(),
			Timestamp: time.Now().UTC(),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	status := domain.StatusOK
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status = domain.StatusFail
	}

	return domain.Result{
		ID:        fmt.Sprintf("%d", start.UnixNano()),
		CheckID:   cfg.ID,
		TargetID:  t.ID,
		Status:    status,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
		Metadata:  map[string]string{"status_code": fmt.Sprintf("%d", resp.StatusCode)},
	}
}
