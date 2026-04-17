// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/gosre/gosre-cli/internal/config"
)

var targetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "Manage targets from ~/.gosre.yaml",
}

var targetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List targets configured in ~/.gosre.yaml",
	RunE:  runTargetsList,
}

func init() {
	targetsCmd.AddCommand(targetsListCmd)
	rootCmd.AddCommand(targetsCmd)
}

func runTargetsList(cmd *cobra.Command, _ []string) error {
	quiet, _ := cmd.Flags().GetBool("quiet")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if quiet {
		return nil
	}

	if len(cfg.Targets) == 0 {
		fmt.Println("no targets configured in ~/.gosre.yaml")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tTYPE\tADDRESS\tTAGS"); err != nil {
		return err
	}
	for _, t := range cfg.Targets {
		tags := strings.Join(t.Tags, ",")
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", t.Name, t.Type, t.Address, tags); err != nil {
			return err
		}
	}
	return tw.Flush()
}
