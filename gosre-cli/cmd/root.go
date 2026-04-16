// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "gosre",
	Short:   "GoSRE — SRE platform CLI",
	Version: "0.1.0",
}

// Execute is the entry point called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("output", "table", "Output format: table|json")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress output; use exit code only")
	rootCmd.PersistentFlags().String("timeout", "10s", "Check timeout (e.g. 10s, 30s)")
}
