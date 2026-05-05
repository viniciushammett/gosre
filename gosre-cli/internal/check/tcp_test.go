// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

func TestTCPChecker_Execute(t *testing.T) {
	// Start a local TCP listener for the "port open" case.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	openAddr := ln.Addr().String()

	tests := []struct {
		name       string
		address    string
		timeout    time.Duration
		wantStatus domain.CheckStatus
	}{
		{
			name:       "port open",
			address:    openAddr,
			timeout:    5 * time.Second,
			wantStatus: domain.StatusOK,
		},
		{
			name:       "timeout — unreachable address with short deadline",
			address:    "192.0.2.1:9999", // TEST-NET, routable but no host
			timeout:    50 * time.Millisecond,
			wantStatus: domain.StatusTimeout,
		},
		{
			name:       "port closed — connection refused",
			address:    "127.0.0.1:1", // port 1 is always closed
			timeout:    5 * time.Second,
			wantStatus: domain.StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewTCPChecker()
			target := domain.Target{ID: "t1", Address: tt.address}
			cfg := domain.CheckConfig{ID: "c1", Timeout: tt.timeout}

			result := checker.Execute(context.Background(), target, cfg)

			assert.Equal(t, tt.wantStatus, result.Status)
			assert.Equal(t, "t1", result.TargetID)
			assert.Equal(t, "c1", result.CheckID)
			assert.False(t, result.Timestamp.IsZero())
			assert.Greater(t, result.Duration, time.Duration(0))
		})
	}
}
