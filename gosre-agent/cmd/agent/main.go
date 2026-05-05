// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	natsjets "github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/viniciushammett/gosre/gosre-agent/internal/agent"
	gosrejs "github.com/viniciushammett/gosre/gosre-events/jetstream"
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

	connectNATS(ctx, a, logger)

	a.Run(ctx, 10*time.Second)
}

// connectNATS attempts to connect to NATS and wires up the publisher and
// assignment subscription on the agent. All errors are treated as warnings —
// the agent continues in HTTP-only mode if NATS is unavailable.
func connectNATS(ctx context.Context, a *agent.Agent, logger *zap.Logger) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Warn("nats connect failed; operating in HTTP-only mode",
			zap.String("url", natsURL), zap.Error(err))
		return
	}

	js, err := natsjets.New(nc)
	if err != nil {
		_ = nc.Drain()
		logger.Warn("nats jetstream init failed; operating in HTTP-only mode", zap.Error(err))
		return
	}

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := gosrejs.EnsureStream(initCtx, js); err != nil {
		_ = nc.Drain()
		logger.Warn("nats stream ensure failed; operating in HTTP-only mode", zap.Error(err))
		return
	}

	a.SetPublisher(gosrejs.NewPublisher(js))

	if err := a.SubscribeAssignments(ctx, js); err != nil {
		logger.Warn("nats assignment subscription failed; HTTP poll active", zap.Error(err))
		// publisher still active for heartbeat even if subscription failed
	} else {
		logger.Info("nats connected; push assignments active", zap.String("url", natsURL))
	}
}
