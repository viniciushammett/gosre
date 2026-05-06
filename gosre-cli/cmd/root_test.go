// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package cmd

import (
	"testing"
	"time"
)

// TestRootFlagParsing verifies that rootCmd's persistent flags are registered
// correctly and that ParseFlags stores values without error. Tests run
// sequentially; t.Cleanup restores defaults so subtests are independent.
func TestRootFlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantOutput  string
		wantTimeout string
		wantQuiet   bool
	}{
		{
			name:        "--output table is accepted",
			args:        []string{"--output", "table"},
			wantOutput:  "table",
			wantTimeout: "10s",
			wantQuiet:   false,
		},
		{
			name:        "--output json is accepted",
			args:        []string{"--output", "json"},
			wantOutput:  "json",
			wantTimeout: "10s",
			wantQuiet:   false,
		},
		{
			name:        "--timeout 30s is stored as string for later duration parse",
			args:        []string{"--timeout", "30s"},
			wantOutput:  "table",
			wantTimeout: "30s",
			wantQuiet:   false,
		},
		{
			name:        "--output unrecognized value accepted at flag level (validated at runtime)",
			args:        []string{"--output", "yaml"},
			wantOutput:  "yaml",
			wantTimeout: "10s",
			wantQuiet:   false,
		},
		{
			name:        "--quiet enables quiet mode",
			args:        []string{"--quiet"},
			wantOutput:  "table",
			wantTimeout: "10s",
			wantQuiet:   true,
		},
		{
			name:        "combined --output and --timeout",
			args:        []string{"--output", "json", "--timeout", "5s"},
			wantOutput:  "json",
			wantTimeout: "5s",
			wantQuiet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Restore persistent flag defaults after each subtest.
			t.Cleanup(func() {
				_ = rootCmd.PersistentFlags().Set("output", "table")
				_ = rootCmd.PersistentFlags().Set("timeout", "10s")
				_ = rootCmd.PersistentFlags().Set("quiet", "false")
			})

			err := rootCmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFlags(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			gotOutput, err := rootCmd.PersistentFlags().GetString("output")
			if err != nil {
				t.Fatalf("GetString(output): %v", err)
			}
			if gotOutput != tt.wantOutput {
				t.Errorf("--output = %q, want %q", gotOutput, tt.wantOutput)
			}

			gotTimeout, err := rootCmd.PersistentFlags().GetString("timeout")
			if err != nil {
				t.Fatalf("GetString(timeout): %v", err)
			}
			if gotTimeout != tt.wantTimeout {
				t.Errorf("--timeout = %q, want %q", gotTimeout, tt.wantTimeout)
			}

			gotQuiet, err := rootCmd.PersistentFlags().GetBool("quiet")
			if err != nil {
				t.Fatalf("GetBool(quiet): %v", err)
			}
			if gotQuiet != tt.wantQuiet {
				t.Errorf("--quiet = %v, want %v", gotQuiet, tt.wantQuiet)
			}
		})
	}
}

// TestTimeoutFlagInterpretation verifies that valid --timeout values stored by
// ParseFlags can be successfully converted to time.Duration, matching the
// production behaviour in runCheckHTTP.
func TestTimeoutFlagInterpretation(t *testing.T) {
	validTimeouts := []struct {
		raw  string
		want time.Duration
	}{
		{"10s", 10 * time.Second},
		{"30s", 30 * time.Second},
		{"1m", time.Minute},
		{"500ms", 500 * time.Millisecond},
	}

	for _, tc := range validTimeouts {
		t.Run(tc.raw, func(t *testing.T) {
			t.Cleanup(func() {
				_ = rootCmd.PersistentFlags().Set("timeout", "10s")
			})

			if err := rootCmd.ParseFlags([]string{"--timeout", tc.raw}); err != nil {
				t.Fatalf("ParseFlags(--timeout %s) error = %v", tc.raw, err)
			}

			stored, _ := rootCmd.PersistentFlags().GetString("timeout")
			d, err := time.ParseDuration(stored)
			if err != nil {
				t.Fatalf("time.ParseDuration(%q) error = %v", stored, err)
			}
			if d != tc.want {
				t.Errorf("duration = %v, want %v", d, tc.want)
			}
		})
	}
}
