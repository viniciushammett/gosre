// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gosre/gosre-sdk/client"
)

// writeJSON writes a JSON response to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// apiOK returns a standard API envelope with the given data payload.
func apiOK(data any) map[string]any {
	return map[string]any{"data": data, "error": nil}
}

// apiErr returns a standard API envelope with an error payload.
func apiErr(code, message string) map[string]any {
	return map[string]any{
		"data":  nil,
		"error": map[string]string{"code": code, "message": message},
	}
}

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/healthz" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "0.1.0"})
	}))
	defer srv.Close()

	c := client.New(srv.URL, "")
	got, err := c.Healthz(context.Background())
	if err != nil {
		t.Fatalf("Healthz() error = %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("Status = %q, want %q", got.Status, "ok")
	}
	if got.Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", got.Version, "0.1.0")
	}
}

func TestListTargets(t *testing.T) {
	want := []client.Target{{ID: "t1", Name: "my-api", Type: "http", Address: "https://example.com"}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/targets" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if key := r.Header.Get("X-API-Key"); key != "test-key" {
			t.Errorf("X-API-Key = %q, want %q", key, "test-key")
		}
		writeJSON(w, http.StatusOK, apiOK(want))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "test-key")
	got, err := c.ListTargets(context.Background())
	if err != nil {
		t.Fatalf("ListTargets() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != "t1" || got[0].Name != "my-api" {
		t.Errorf("got %+v, want %+v", got[0], want[0])
	}
}

func TestCreateTarget(t *testing.T) {
	req := client.CreateTargetRequest{Name: "new-api", Type: "http", Address: "https://new.example.com"}
	created := client.Target{ID: "t2", Name: req.Name, Type: req.Type, Address: req.Address}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/targets" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		var body client.CreateTargetRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if body.Name != req.Name || body.Address != req.Address || body.Type != req.Type {
			t.Errorf("body = %+v, want %+v", body, req)
		}
		writeJSON(w, http.StatusCreated, apiOK(created))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "test-key")
	got, err := c.CreateTarget(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTarget() error = %v", err)
	}
	if got.ID != "t2" {
		t.Errorf("ID = %q, want t2", got.ID)
	}
	if got.Name != req.Name {
		t.Errorf("Name = %q, want %q", got.Name, req.Name)
	}
}

func TestGetTarget(t *testing.T) {
	tests := []struct {
		name       string
		serverCode int
		serverBody any
		wantErr    string
	}{
		{
			name:       "found",
			serverCode: http.StatusOK,
			serverBody: apiOK(client.Target{ID: "t1", Name: "my-api", Type: "http", Address: "https://example.com"}),
		},
		{
			name:       "not found",
			serverCode: http.StatusNotFound,
			serverBody: apiErr("target_not_found", "target not found"),
			wantErr:    "target_not_found: target not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, tt.serverCode, tt.serverBody)
			}))
			defer srv.Close()

			c := client.New(srv.URL, "test-key")
			got, err := c.GetTarget(context.Background(), "t1")

			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != "t1" {
				t.Errorf("ID = %q, want t1", got.ID)
			}
		})
	}
}

func TestListResults(t *testing.T) {
	want := []client.Result{{ID: "r1", TargetID: "t1", Status: "ok"}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/results" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("target_id"); got != "t1" {
			t.Errorf("target_id query param = %q, want %q", got, "t1")
		}
		writeJSON(w, http.StatusOK, apiOK(want))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "test-key")
	got, err := c.ListResults(context.Background(), client.ListResultsParams{TargetID: "t1"})
	if err != nil {
		t.Fatalf("ListResults() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "r1" {
		t.Errorf("got %+v, want 1 result with ID r1", got)
	}
}

func TestListIncidents(t *testing.T) {
	want := []client.Incident{{ID: "i1", TargetID: "t1", State: "open"}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/incidents" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("state"); got != "open" {
			t.Errorf("state query param = %q, want %q", got, "open")
		}
		writeJSON(w, http.StatusOK, apiOK(want))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "test-key")
	got, err := c.ListIncidents(context.Background(), client.ListIncidentsParams{State: "open"})
	if err != nil {
		t.Fatalf("ListIncidents() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "i1" {
		t.Errorf("got %+v, want 1 incident with ID i1", got)
	}
}

func TestRetry(t *testing.T) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			writeJSON(w, http.StatusInternalServerError, apiErr("internal_error", "server error"))
			return
		}
		writeJSON(w, http.StatusOK, apiOK([]client.Target{{ID: "t1", Name: "my-api", Type: "http", Address: "https://example.com"}}))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "test-key")
	got, err := c.ListTargets(context.Background())
	if err != nil {
		t.Fatalf("ListTargets() after retry error = %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("server called %d times, want 3", n)
	}
}
