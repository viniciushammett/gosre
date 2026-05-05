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
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	eventsjetstream "github.com/viniciushammett/gosre/gosre-events/jetstream"
	"github.com/viniciushammett/gosre/gosre-scheduler/internal/scheduler"
	redisstore "github.com/viniciushammett/gosre/gosre-scheduler/internal/store/redis"
	sdkclient "github.com/viniciushammett/gosre/gosre-sdk/client"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync() //nolint:errcheck

	apiURL := env("GOSRE_API_URL", "http://localhost:8080")
	apiKey := env("GOSRE_API_KEY", "")
	redisURL := env("REDIS_URL", "redis://localhost:6379")
	natsURL := env("NATS_URL", nats.DefaultURL)

	interval, err := time.ParseDuration(env("SCHEDULER_INTERVAL", "30s"))
	if err != nil {
		log.Fatal("invalid SCHEDULER_INTERVAL", zap.Error(err))
	}

	// Redis.
	redisopts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("invalid REDIS_URL", zap.Error(err))
	}
	rdb := redis.NewClient(redisopts)
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("redis ping failed", zap.Error(err))
	}

	// NATS JetStream.
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("nats connect failed", zap.Error(err))
	}
	defer nc.Drain() //nolint:errcheck

	js, err := natsjets.New(nc)
	if err != nil {
		log.Fatal("jetstream init failed", zap.Error(err))
	}
	if err := eventsjetstream.EnsureStream(ctx, js); err != nil {
		log.Fatal("ensure GOSRE stream failed", zap.Error(err))
	}

	pub := eventsjetstream.NewPublisher(js)
	store := redisstore.NewAssignmentStore(rdb)
	apiCli := sdkclient.New(apiURL, apiKey)

	sched := scheduler.New(apiCli, store, pub, interval, log)

	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := sched.Run(runCtx); err != nil {
		log.Error("scheduler exited with error", zap.Error(err))
		os.Exit(1)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
