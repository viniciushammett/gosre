// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

func TestDNSChecker_Execute(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		params     map[string]string
		ctxFn      func() (context.Context, context.CancelFunc)
		wantStatus domain.CheckStatus
	}{
		{
			name:    "valid host — A record (default)",
			address: "localhost",
			params:  nil,
			ctxFn: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			wantStatus: domain.StatusOK,
		},
		{
			name:    "valid host — explicit A record",
			address: "localhost",
			params:  map[string]string{"record_type": "A"},
			ctxFn: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			wantStatus: domain.StatusOK,
		},
		{
			name:    "nonexistent domain",
			address: "does-not-exist.invalid",
			params:  nil,
			ctxFn: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			wantStatus: domain.StatusFail,
		},
		{
			name:    "timeout — context already cancelled",
			address: "localhost",
			params:  nil,
			ctxFn: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // cancel before Execute is called
				return ctx, func() {}
			},
			wantStatus: domain.StatusTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.ctxFn()
			defer cancel()

			checker := NewDNSChecker()
			target := domain.Target{ID: "t1", Address: tt.address}
			cfg := domain.CheckConfig{ID: "c1", Params: tt.params}

			result := checker.Execute(ctx, target, cfg)

			assert.Equal(t, tt.wantStatus, result.Status)
			assert.Equal(t, "t1", result.TargetID)
			assert.Equal(t, "c1", result.CheckID)
			assert.False(t, result.Timestamp.IsZero())
			assert.Greater(t, result.Duration, time.Duration(0))

			if tt.wantStatus == domain.StatusOK {
				assert.NotEmpty(t, result.Metadata["resolved"])
			}
		})
	}
}
