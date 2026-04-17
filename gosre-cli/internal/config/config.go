// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// TargetConfig represents a named target entry in ~/.gosre.yaml.
type TargetConfig struct {
	Name    string   `mapstructure:"name"`
	Type    string   `mapstructure:"type"`
	Address string   `mapstructure:"address"`
	Tags    []string `mapstructure:"tags"`
}

// APIConfig holds the API connection settings from ~/.gosre.yaml.
type APIConfig struct {
	URL string `mapstructure:"url"`
	Key string `mapstructure:"key"`
}

// DefaultsConfig holds default CLI flag values from ~/.gosre.yaml.
type DefaultsConfig struct {
	Timeout string `mapstructure:"timeout"`
	Output  string `mapstructure:"output"`
}

// Config is the top-level structure for ~/.gosre.yaml.
type Config struct {
	API      APIConfig      `mapstructure:"api"`
	Defaults DefaultsConfig `mapstructure:"defaults"`
	Targets  []TargetConfig `mapstructure:"targets"`
}

// Load reads ~/.gosre.yaml and returns the parsed Config.
// If the file does not exist, an empty Config is returned without error.
// If the file exists but is invalid, an error is returned with context.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("config: resolve home directory: %w", err)
	}

	v := viper.New()
	v.SetConfigName(".gosre")
	v.SetConfigType("yaml")
	v.AddConfigPath(home)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("config: read ~/.gosre.yaml: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal ~/.gosre.yaml: %w", err)
	}
	return &cfg, nil
}

// FindTarget searches the targets list by name (case-insensitive).
// Returns an error if no target with the given name is found.
func (c *Config) FindTarget(name string) (*TargetConfig, error) {
	for i := range c.Targets {
		if strings.EqualFold(c.Targets[i].Name, name) {
			return &c.Targets[i], nil
		}
	}
	return nil, fmt.Errorf("target %q not found in config", name)
}
