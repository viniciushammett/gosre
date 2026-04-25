// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package config

import "github.com/spf13/viper"

// Config holds all runtime configuration for gosre-auth.
type Config struct {
	Port               string
	JWTSecret          string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
}

// Load reads configuration from environment variables.
func Load() Config {
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("JWT_SECRET", "dev-secret-change-in-production")
	viper.SetDefault("GITHUB_REDIRECT_URL", "http://localhost:8081/auth/github/callback")
	viper.AutomaticEnv()

	return Config{
		Port:               viper.GetString("PORT"),
		JWTSecret:          viper.GetString("JWT_SECRET"),
		GitHubClientID:     viper.GetString("GOSRE_GITHUB_CLIENT_ID"),
		GitHubClientSecret: viper.GetString("GOSRE_GITHUB_CLIENT_SECRET"),
		GitHubRedirectURL:  viper.GetString("GITHUB_REDIRECT_URL"),
	}
}
