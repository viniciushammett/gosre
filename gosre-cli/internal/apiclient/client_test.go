// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// ── stableID ─────────────────────────────────────────────────────────────────

func TestStableID(t *testing.T) {
	t.Run("deterministic — same input always yields same output", func(t *testing.T) {
		inputs := [][]string{
			{"https://example.com"},
			{"https://example.com", "http"},
			{""},
			{"a", "b", "c"},
		}
		for _, parts := range inputs {
			a := stableID(parts...)
			b := stableID(parts...)
			if a != b {
				t.Errorf("stableID(%v) not deterministic: %q != %q", parts, a, b)
			}
			if a == "" {
				t.Errorf("stableID(%v) returned empty string", parts)
			}
		}
	})

	t.Run("different addresses produce different IDs", func(t *testing.T) {
		a := stableID("https://example.com")
		b := stableID("https://other.com")
		if a == b {
			t.Errorf("stableID: distinct addresses produced same ID %q", a)
		}
	})

	t.Run("one part vs two parts produce different IDs", func(t *testing.T) {
		one := stableID("https://example.com")
		two := stableID("https://example.com", "http")
		if one == two {
			t.Errorf("stableID: one part and two parts produced same ID %q", one)
		}
	})

	t.Run("separator matters — concatenated parts are not equivalent", func(t *testing.T) {
		// "ab" "c" must differ from "a" "bc"
		abc := stableID("ab", "c")
		abc2 := stableID("a", "bc")
		if abc == abc2 {
			t.Errorf("stableID: different split points produced same ID %q", abc)
		}
	})
}

// ── RunCheck ─────────────────────────────────────────────────────────────────

// apiEnvelope mirrors the gosre-api standard response wrapper.
type apiEnvelope struct {
	Data  any     `json:"data"`
	Error *apiErr `json:"error"`
}

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeEnvelope(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, apiEnvelope{Data: data})
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, apiEnvelope{Error: &apiErr{Code: code, Message: msg}})
}

// successMux builds an httptest.Server that handles the three endpoints that
// RunCheck hits: POST /api/v1/targets, POST /api/v1/checks, POST /api/v1/checks/:id/run.
func successServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/targets", func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(w, http.StatusCreated, map[string]any{
			"id":      "tgt-1",
			"name":    "test-target",
			"type":    "http",
			"address": "https://example.com/healthz",
		})
	})

	mux.HandleFunc("/api/v1/checks", func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(w, http.StatusCreated, map[string]any{
			"id":        "chk-1",
			"type":      "http",
			"target_id": "tgt-1",
		})
	})

	// Matches both POST /api/v1/checks/{id}/run and the listing above — the
	// listing handler is registered first so the mux uses longest-prefix.
	mux.HandleFunc("/api/v1/checks/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/run") {
			http.NotFound(w, r)
			return
		}
		writeEnvelope(w, http.StatusCreated, map[string]any{
			"id":          "result-1",
			"check_id":    "chk-1",
			"target_id":   "tgt-1",
			"status":      "ok",
			"duration_ms": int64(50),
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"metadata":    map[string]string{},
		})
	})

	return httptest.NewServer(mux)
}

func TestRunCheck(t *testing.T) {
	target := domain.Target{
		Name:    "test-api",
		Type:    domain.TargetTypeHTTP,
		Address: "https://example.com/healthz",
	}
	cfg := domain.CheckConfig{
		Type:    domain.CheckTypeHTTP,
		Timeout: 5 * time.Second,
	}

	tests := []struct {
		name      string
		newServer func(t *testing.T) *httptest.Server
		closeNow  bool // close before calling RunCheck (unreachable)
		wantErr   bool
		wantOK    bool // result status should be "ok"
	}{
		{
			name:      "success — 200 + valid Result JSON",
			newServer: successServer,
			wantErr:   false,
			wantOK:    true,
		},
		{
			name: "error — server returns 500 on upsert target",
			newServer: func(t *testing.T) *httptest.Server {
				t.Helper()
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					writeError(w, http.StatusInternalServerError, "internal_error", "forced failure")
				}))
			},
			wantErr: true,
		},
		{
			name: "error — server unreachable",
			newServer: func(t *testing.T) *httptest.Server {
				t.Helper()
				// Server is started and immediately closed; connection refused.
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			closeNow: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.newServer(t)
			if tt.closeNow {
				srv.Close()
			} else {
				t.Cleanup(srv.Close)
			}

			c := New(srv.URL, "")
			result, err := c.RunCheck(context.Background(), target, cfg)

			if (err != nil) != tt.wantErr {
				t.Fatalf("RunCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantOK {
				if result.Status != domain.StatusOK {
					t.Errorf("RunCheck() Status = %q, want %q", result.Status, domain.StatusOK)
				}
				if result.ID == "" {
					t.Error("RunCheck() Result.ID is empty")
				}
			}
		})
	}
}
