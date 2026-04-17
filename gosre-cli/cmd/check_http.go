// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gosre/gosre-sdk/domain"
	"github.com/spf13/cobra"

	"github.com/gosre/gosre-cli/internal/check"
	"github.com/gosre/gosre-cli/internal/config"
	"github.com/gosre/gosre-cli/internal/output"
)

var checkHTTPCmd = &cobra.Command{
	Use:   "http",
	Short: "Check HTTP/HTTPS endpoint reachability",
	RunE:  runCheckHTTP,
}

func init() {
	checkHTTPCmd.Flags().StringP("address", "a", "", "URL to check")
	checkHTTPCmd.Flags().StringP("target-name", "n", "", "target name from ~/.gosre.yaml (overrides --address)")
	checkCmd.AddCommand(checkHTTPCmd)
}

func runCheckHTTP(cmd *cobra.Command, _ []string) error {
	address, _ := cmd.Flags().GetString("address")
	targetName, _ := cmd.Flags().GetString("target-name")
	timeoutStr, _ := cmd.Flags().GetString("timeout")
	outputFmt, _ := cmd.Flags().GetString("output")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if targetName != "" {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		t, err := cfg.FindTarget(targetName)
		if err != nil {
			return err
		}
		address = t.Address
	} else if address == "" {
		return fmt.Errorf("--address or --target-name required")
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return fmt.Errorf("invalid timeout %q: %w", timeoutStr, err)
	}

	target := domain.Target{
		ID:      address,
		Address: address,
		Type:    domain.TargetTypeHTTP,
	}
	cfg := domain.CheckConfig{
		ID:      "cli",
		Type:    domain.CheckTypeHTTP,
		Timeout: timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	checker := check.NewHTTPChecker()
	result := checker.Execute(ctx, target, cfg)

	if !quiet {
		if err := output.Write(os.Stdout, output.Format(outputFmt), []domain.Result{result}); err != nil {
			return err
		}
	}

	if result.Status != domain.StatusOK {
		os.Exit(1)
	}
	return nil
}
