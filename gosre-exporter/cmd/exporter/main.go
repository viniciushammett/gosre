// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/gosre/gosre-exporter/internal/apiclient"
	"github.com/gosre/gosre-exporter/internal/collector"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	apiURL := os.Getenv("GOSRE_API_URL")
	if apiURL == "" {
		logger.Fatal("GOSRE_API_URL is required")
	}

	apiKey := os.Getenv("GOSRE_API_KEY")

	pollInterval := 30 * time.Second
	if raw := os.Getenv("GOSRE_POLL_INTERVAL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			logger.Fatal("invalid GOSRE_POLL_INTERVAL", zap.String("value", raw), zap.Error(err))
		}
		pollInterval = d
	}

	port := os.Getenv("GOSRE_METRICS_PORT")
	if port == "" {
		port = "9090"
	}

	client := apiclient.New(apiURL, apiKey)
	col := collector.New(client, pollInterval, logger)

	reg := prometheus.NewRegistry()
	reg.MustRegister(col)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("exporter started",
			zap.String("port", port),
			zap.String("api_url", apiURL),
			zap.Duration("poll_interval", pollInterval),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
	logger.Info("exporter stopped")
}
