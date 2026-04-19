// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/gosre/gosre-agent/internal/agent"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	apiURL := os.Getenv("GOSRE_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	apiKey := os.Getenv("GOSRE_API_KEY")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a := agent.New(apiURL, apiKey, logger)

	if err := a.Register(ctx); err != nil {
		logger.Fatal("agent registration failed", zap.Error(err))
	}

	a.Run(ctx, 10*time.Second)
}
