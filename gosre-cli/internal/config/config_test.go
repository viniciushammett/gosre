// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		fileYAML  string // empty = no file written
		wantErr   bool
		check     func(t *testing.T, cfg *Config)
	}{
		{
			name:     "absent config returns empty Config without error",
			fileYAML: "",
			wantErr:  false,
			check: func(t *testing.T, cfg *Config) {
				if cfg == nil {
					t.Fatal("Load() returned nil config, want empty Config")
				}
				if len(cfg.Targets) != 0 {
					t.Errorf("Targets = %d, want 0", len(cfg.Targets))
				}
				if cfg.API.URL != "" {
					t.Errorf("API.URL = %q, want empty", cfg.API.URL)
				}
			},
		},
		{
			name: "valid config populates all fields",
			fileYAML: `
api:
  url: http://localhost:8080
  key: secret-key
defaults:
  timeout: 5s
  output: json
targets:
  - name: my-api
    type: http
    address: https://example.com/healthz
    tags: [production, api]
  - name: db-primary
    type: tcp
    address: db.example.com:5432
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.API.URL != "http://localhost:8080" {
					t.Errorf("API.URL = %q, want %q", cfg.API.URL, "http://localhost:8080")
				}
				if cfg.API.Key != "secret-key" {
					t.Errorf("API.Key = %q, want %q", cfg.API.Key, "secret-key")
				}
				if cfg.Defaults.Timeout != "5s" {
					t.Errorf("Defaults.Timeout = %q, want %q", cfg.Defaults.Timeout, "5s")
				}
				if cfg.Defaults.Output != "json" {
					t.Errorf("Defaults.Output = %q, want %q", cfg.Defaults.Output, "json")
				}
				if len(cfg.Targets) != 2 {
					t.Fatalf("Targets length = %d, want 2", len(cfg.Targets))
				}
				if cfg.Targets[0].Name != "my-api" {
					t.Errorf("Targets[0].Name = %q, want %q", cfg.Targets[0].Name, "my-api")
				}
				if cfg.Targets[0].Type != "http" {
					t.Errorf("Targets[0].Type = %q, want %q", cfg.Targets[0].Type, "http")
				}
				if cfg.Targets[0].Address != "https://example.com/healthz" {
					t.Errorf("Targets[0].Address = %q, want %q", cfg.Targets[0].Address, "https://example.com/healthz")
				}
				if len(cfg.Targets[0].Tags) != 2 {
					t.Errorf("Targets[0].Tags length = %d, want 2", len(cfg.Targets[0].Tags))
				}
			},
		},
		{
			name:     "config with only api block",
			fileYAML: "api:\n  url: http://api.example.com\n",
			wantErr:  false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.API.URL != "http://api.example.com" {
					t.Errorf("API.URL = %q, want %q", cfg.API.URL, "http://api.example.com")
				}
				if len(cfg.Targets) != 0 {
					t.Errorf("Targets = %d, want 0", len(cfg.Targets))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Setenv("HOME", tmpDir)

			if tt.fileYAML != "" {
				path := filepath.Join(tmpDir, ".gosre.yaml")
				if err := os.WriteFile(path, []byte(tt.fileYAML), 0600); err != nil {
					t.Fatalf("setup: write config: %v", err)
				}
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestFindTarget(t *testing.T) {
	cfg := &Config{
		Targets: []TargetConfig{
			{Name: "my-api", Type: "http", Address: "https://example.com/healthz"},
			{Name: "db-primary", Type: "tcp", Address: "db.example.com:5432"},
		},
	}

	tests := []struct {
		name      string
		search    string
		wantName  string
		wantFound bool
	}{
		{
			name:      "found — exact match",
			search:    "my-api",
			wantName:  "my-api",
			wantFound: true,
		},
		{
			name:      "found — case-insensitive upper",
			search:    "MY-API",
			wantName:  "my-api",
			wantFound: true,
		},
		{
			name:      "found — case-insensitive mixed",
			search:    "Db-Primary",
			wantName:  "db-primary",
			wantFound: true,
		},
		{
			name:      "not found — unknown name",
			search:    "nonexistent",
			wantFound: false,
		},
		{
			name:      "not found — empty search string",
			search:    "",
			wantFound: false,
		},
		{
			name:      "not found — partial name",
			search:    "my",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.FindTarget(tt.search)
			if tt.wantFound {
				if err != nil {
					t.Fatalf("FindTarget(%q) error = %v, want nil", tt.search, err)
				}
				if got == nil {
					t.Fatal("FindTarget() = nil, want non-nil")
				}
				if got.Name != tt.wantName {
					t.Errorf("FindTarget() Name = %q, want %q", got.Name, tt.wantName)
				}
			} else {
				if err == nil {
					t.Errorf("FindTarget(%q) error = nil, want error", tt.search)
				}
				if got != nil {
					t.Errorf("FindTarget() = %v, want nil", got)
				}
			}
		})
	}
}

func TestFindTargetEmptyConfig(t *testing.T) {
	cfg := &Config{}
	got, err := cfg.FindTarget("anything")
	if err == nil {
		t.Error("FindTarget() on empty config error = nil, want error")
	}
	if got != nil {
		t.Errorf("FindTarget() = %v, want nil", got)
	}
}
