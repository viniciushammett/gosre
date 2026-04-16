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
	"github.com/gosre/gosre-cli/internal/output"
)

var checkTCPCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Check TCP port connectivity",
	RunE:  runCheckTCP,
}

func init() {
	checkTCPCmd.Flags().StringP("address", "a", "", "host:port to check (required)")
	_ = checkTCPCmd.MarkFlagRequired("address")
	checkCmd.AddCommand(checkTCPCmd)
}

func runCheckTCP(cmd *cobra.Command, _ []string) error {
	address, _ := cmd.Flags().GetString("address")
	timeoutStr, _ := cmd.Flags().GetString("timeout")
	outputFmt, _ := cmd.Flags().GetString("output")
	quiet, _ := cmd.Flags().GetBool("quiet")

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return fmt.Errorf("invalid timeout %q: %w", timeoutStr, err)
	}

	target := domain.Target{
		ID:      address,
		Address: address,
		Type:    domain.TargetTypeTCP,
	}
	cfg := domain.CheckConfig{
		ID:      "cli",
		Type:    domain.CheckTypeTCP,
		Timeout: timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := check.NewTCPChecker().Execute(ctx, target, cfg)

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
