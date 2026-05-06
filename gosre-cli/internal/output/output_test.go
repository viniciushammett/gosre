// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

var (
	fixedTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	resultOK = domain.Result{
		ID:         "r1",
		CheckID:    "c1",
		TargetID:   "t1",
		TargetName: "my-api",
		Status:     domain.StatusOK,
		Duration:   50 * time.Millisecond,
		Timestamp:  fixedTime,
		Metadata:   map[string]string{},
	}

	resultFail = domain.Result{
		ID:        "r2",
		CheckID:   "c2",
		TargetID:  "t2",
		Status:    domain.StatusFail,
		Duration:  200 * time.Millisecond,
		Error:     "connection refused",
		Timestamp: fixedTime,
		Metadata:  map[string]string{},
	}
)

func TestTable(t *testing.T) {
	tests := []struct {
		name     string
		results  []domain.Result
		contains []string
		absent   []string
	}{
		{
			name:    "header always present",
			results: []domain.Result{},
			contains: []string{
				"TIMESTAMP", "TARGET", "STATUS", "DURATION",
			},
		},
		{
			name:     "empty results — no data rows",
			results:  []domain.Result{},
			contains: []string{"TIMESTAMP"},
			absent:   []string{"my-api", "ok"},
		},
		{
			name:    "single ok result with TargetName",
			results: []domain.Result{resultOK},
			contains: []string{
				"my-api",
				"ok",
				"50ms",
				"12:00:00",
			},
		},
		{
			name:    "result without TargetName falls back to TargetID",
			results: []domain.Result{resultFail},
			contains: []string{
				"t2",
				"fail",
				"200ms",
				"connection refused",
			},
			absent: []string{"my-api"},
		},
		{
			name:    "multiple results all appear",
			results: []domain.Result{resultOK, resultFail},
			contains: []string{
				"my-api", "ok",
				"t2", "fail",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Table(&buf, tt.results); err != nil {
				t.Fatalf("Table() error = %v", err)
			}
			got := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("Table() output missing %q\ngot:\n%s", s, got)
				}
			}
			for _, s := range tt.absent {
				if strings.Contains(got, s) {
					t.Errorf("Table() output should not contain %q\ngot:\n%s", s, got)
				}
			}
		})
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name    string
		results []domain.Result
	}{
		{
			name:    "empty list encodes as empty array",
			results: []domain.Result{},
		},
		{
			name:    "single ok result round-trips",
			results: []domain.Result{resultOK},
		},
		{
			name:    "multiple results round-trip",
			results: []domain.Result{resultOK, resultFail},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := JSON(&buf, tt.results); err != nil {
				t.Fatalf("JSON() error = %v", err)
			}

			var got []domain.Result
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("JSON() produced invalid JSON: %v\noutput: %s", err, buf.String())
			}

			if len(got) != len(tt.results) {
				t.Fatalf("JSON() round-trip: got %d results, want %d", len(got), len(tt.results))
			}
			for i, r := range tt.results {
				if got[i].ID != r.ID {
					t.Errorf("result[%d].ID = %q, want %q", i, got[i].ID, r.ID)
				}
				if got[i].Status != r.Status {
					t.Errorf("result[%d].Status = %q, want %q", i, got[i].Status, r.Status)
				}
				if got[i].TargetID != r.TargetID {
					t.Errorf("result[%d].TargetID = %q, want %q", i, got[i].TargetID, r.TargetID)
				}
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		format  Format
		wantErr bool
	}{
		{
			name:    "table format",
			format:  FormatTable,
			wantErr: false,
		},
		{
			name:    "json format",
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:    "unknown format returns error",
			format:  Format("yaml"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Write(&buf, tt.format, []domain.Result{resultOK})
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
