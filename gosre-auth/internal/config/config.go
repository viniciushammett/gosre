// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package config

import "github.com/spf13/viper"

// Config holds all runtime configuration for gosre-auth.
type Config struct {
	Port      string
	JWTSecret string
}

// Load reads configuration from environment variables.
func Load() Config {
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("JWT_SECRET", "dev-secret-change-in-production")
	viper.AutomaticEnv()

	return Config{
		Port:      viper.GetString("PORT"),
		JWTSecret: viper.GetString("JWT_SECRET"),
	}
}
