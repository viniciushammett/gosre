// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"

	"github.com/viniciushammett/gosre/gosre-cli/internal/check"
	"github.com/viniciushammett/gosre/gosre-cli/internal/config"
	"github.com/viniciushammett/gosre/gosre-cli/internal/output"
)

var checkDNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "Check DNS resolution for a hostname",
	RunE:  runCheckDNS,
}

func init() {
	checkDNSCmd.Flags().StringP("address", "a", "", "hostname to resolve")
	checkDNSCmd.Flags().StringP("target-name", "n", "", "target name from ~/.gosre.yaml (overrides --address)")
	checkDNSCmd.Flags().String("record-type", "A", "DNS record type: A, AAAA, CNAME, MX")
	checkCmd.AddCommand(checkDNSCmd)
}

func runCheckDNS(cmd *cobra.Command, _ []string) error {
	address, _ := cmd.Flags().GetString("address")
	targetName, _ := cmd.Flags().GetString("target-name")
	recordType, _ := cmd.Flags().GetString("record-type")
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
		Type:    domain.TargetTypeDNS,
	}
	cfg := domain.CheckConfig{
		ID:      "cli",
		Type:    domain.CheckTypeDNS,
		Timeout: timeout,
		Params:  map[string]string{"record_type": recordType},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := resolveAPIClient(cmd)
	if err != nil {
		return err
	}

	var result domain.Result
	if client != nil {
		result, err = client.RunCheck(ctx, target, cfg)
		if err != nil {
			return fmt.Errorf("api check: %w", err)
		}
	} else {
		result = check.NewDNSChecker().Execute(ctx, target, cfg)
	}

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
