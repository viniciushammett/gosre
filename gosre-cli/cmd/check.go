// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package cmd

import "github.com/spf13/cobra"

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run connectivity and health checks",
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
