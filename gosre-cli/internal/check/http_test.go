// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
	"github.com/stretchr/testify/assert"
)

func TestHTTPChecker_Execute(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		address    func(serverURL string) string
		timeout    time.Duration
		wantStatus domain.CheckStatus
	}{
		{
			name: "reachable 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			address:    func(serverURL string) string { return serverURL },
			timeout:    5 * time.Second,
			wantStatus: domain.StatusOK,
		},
		{
			name: "server returns 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			address:    func(serverURL string) string { return serverURL },
			timeout:    5 * time.Second,
			wantStatus: domain.StatusFail,
		},
		{
			name: "timeout — server sleeps longer than cfg.Timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(200 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			address:    func(serverURL string) string { return serverURL },
			timeout:    50 * time.Millisecond,
			wantStatus: domain.StatusTimeout,
		},
		{
			name:       "unreachable — invalid address",
			handler:    nil,
			address:    func(_ string) string { return "http://invalid.localhost.invalid" },
			timeout:    2 * time.Second,
			wantStatus: domain.StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverURL := ""
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				serverURL = srv.URL
			}

			checker := NewHTTPChecker()
			target := domain.Target{ID: "t1", Address: tt.address(serverURL)}
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
