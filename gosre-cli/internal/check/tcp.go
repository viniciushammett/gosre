// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// TCPChecker validates TCP connectivity to a Target.
type TCPChecker struct{}

// NewTCPChecker returns a ready-to-use TCPChecker.
func NewTCPChecker() *TCPChecker {
	return &TCPChecker{}
}

// Execute dials t.Address over TCP, respecting ctx and cfg.Timeout.
// It implements domain.Checker.
func (c *TCPChecker) Execute(ctx context.Context, t domain.Target, cfg domain.CheckConfig) domain.Result {
	start := time.Now()

	dialCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", t.Address)
	duration := time.Since(start)

	if err != nil {
		status := domain.StatusFail
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			status = domain.StatusTimeout
		} else if dialCtx.Err() == context.DeadlineExceeded {
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
	_ = conn.Close()

	return domain.Result{
		ID:        fmt.Sprintf("%d", start.UnixNano()),
		CheckID:   cfg.ID,
		TargetID:  t.ID,
		Status:    domain.StatusOK,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
	}
}
