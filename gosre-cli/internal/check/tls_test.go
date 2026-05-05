// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

func TestTLSChecker_Execute(t *testing.T) {
	// TLS test server with a self-signed certificate.
	srv := httptest.NewTLSServer(nil)
	defer srv.Close()

	// srv.Listener.Addr() returns "127.0.0.1:<port>" — strip the scheme.
	tlsAddr := srv.Listener.Addr().String()

	tests := []struct {
		name       string
		address    string
		params     map[string]string
		ctxFn      func() (context.Context, context.CancelFunc)
		wantStatus domain.CheckStatus
	}{
		{
			name:    "valid TLS — self-signed cert (insecure skip verify)",
			address: tlsAddr,
			params: map[string]string{
				"insecure": "true",
			},
			ctxFn: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			wantStatus: domain.StatusOK,
		},
		{
			name:    "invalid address — connection refused",
			address: "127.0.0.1:1",
			params:  nil,
			ctxFn: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			wantStatus: domain.StatusFail,
		},
		{
			name:    "timeout — context already cancelled",
			address: tlsAddr,
			params:  map[string]string{"insecure": "true"},
			ctxFn: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
			wantStatus: domain.StatusTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.ctxFn()
			defer cancel()

			checker := NewTLSChecker()
			target := domain.Target{ID: "t1", Address: tt.address}
			cfg := domain.CheckConfig{ID: "c1", Timeout: 5 * time.Second, Params: tt.params}

			result := checker.Execute(ctx, target, cfg)

			assert.Equal(t, tt.wantStatus, result.Status)
			assert.Equal(t, "t1", result.TargetID)
			assert.Equal(t, "c1", result.CheckID)
			assert.False(t, result.Timestamp.IsZero())
			assert.Greater(t, result.Duration, time.Duration(0))

			if tt.wantStatus == domain.StatusOK {
				assert.NotEmpty(t, result.Metadata["expiry_days"])
				// httptest certs have an empty CommonName — just verify the key exists.
				_, hasCN := result.Metadata["common_name"]
				assert.True(t, hasCN, "common_name key must be present in metadata")
			}
		})
	}
}
