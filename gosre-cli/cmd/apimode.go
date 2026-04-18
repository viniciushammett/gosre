// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gosre/gosre-cli/internal/apiclient"
	"github.com/gosre/gosre-cli/internal/config"
)

// resolveAPIClient returns a configured API client if --api-url is set (flag or config).
// Returns nil when no API URL is found — caller should run the check locally.
func resolveAPIClient(cmd *cobra.Command) (*apiclient.Client, error) {
	apiURL, _ := cmd.Flags().GetString("api-url")
	apiKey, _ := cmd.Flags().GetString("api-key")

	if apiURL == "" {
		cfg, err := config.Load()
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
		apiURL = cfg.API.URL
		if apiKey == "" {
			apiKey = cfg.API.Key
		}
	}

	if apiURL == "" {
		return nil, nil
	}
	return apiclient.New(apiURL, apiKey), nil
}
