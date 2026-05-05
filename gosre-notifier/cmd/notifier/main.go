// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Command notifier subscribes to NATS JetStream incident events and dispatches
// notifications to configured channels (Slack, Email, Webhook).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	nats "github.com/nats-io/nats.go"
	natsjets "github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
	"github.com/viniciushammett/gosre/gosre-events/events"
	gosrejs "github.com/viniciushammett/gosre/gosre-events/jetstream"

	"github.com/viniciushammett/gosre/gosre-notifier/internal/apiclient"
	"github.com/viniciushammett/gosre/gosre-notifier/internal/channel"
	"github.com/viniciushammett/gosre/gosre-notifier/internal/dedup"
)

type config struct {
	NatsURL       string
	APIURL        string
	APIKey        string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func loadConfig() config {
	v := viper.New()
	v.SetConfigName("notifier")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/gosre")
	v.SetEnvPrefix("gosre_notifier")
	v.AutomaticEnv()

	v.SetDefault("nats_url", "nats://localhost:4222")
	v.SetDefault("api_url", "http://localhost:8080")
	v.SetDefault("api_key", "")
	v.SetDefault("redis_addr", "localhost:6379")
	v.SetDefault("redis_password", "")
	v.SetDefault("redis_db", 0)

	_ = v.ReadInConfig()

	return config{
		NatsURL:       v.GetString("nats_url"),
		APIURL:        v.GetString("api_url"),
		APIKey:        v.GetString("api_key"),
		RedisAddr:     v.GetString("redis_addr"),
		RedisPassword: v.GetString("redis_password"),
		RedisDB:       v.GetInt("redis_db"),
	}
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("init zap: " + err.Error())
	}
	defer func() { _ = logger.Sync() }()

	cfg := loadConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		logger.Fatal("redis ping", zap.Error(err))
	}

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Fatal("nats connect", zap.Error(err))
	}
	defer func() { _ = nc.Drain() }()

	js, err := natsjets.New(nc)
	if err != nil {
		logger.Fatal("jetstream client", zap.Error(err))
	}

	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()
	if err := gosrejs.EnsureStream(initCtx, js); err != nil {
		logger.Fatal("ensure nats stream", zap.Error(err))
	}

	apiClient := apiclient.New(cfg.APIURL, cfg.APIKey)
	deduplicator := dedup.New(rdb)

	senders := map[domain.NotificationChannelKind]channel.Sender{
		domain.ChannelKindSlack:   &channel.SlackSender{},
		domain.ChannelKindEmail:   &channel.EmailSender{},
		domain.ChannelKindWebhook: &channel.WebhookSender{},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	subscriptions := []struct {
		subject string
		durable string
	}{
		{events.SubjectIncidentsOpened, "notifier-incidents-opened"},
		{events.SubjectIncidentsResolved, "notifier-incidents-resolved"},
	}

	for _, sub := range subscriptions {
		handler := buildHandler(apiClient, deduplicator, senders, logger, sub.subject)
		if err := gosrejs.Subscribe(ctx, js, sub.subject, sub.durable, handler); err != nil {
			logger.Fatal("subscribe", zap.String("subject", sub.subject), zap.Error(err))
		}
		logger.Info("subscribed", zap.String("subject", sub.subject))
	}

	logger.Info("notifier started",
		zap.String("nats_url", cfg.NatsURL),
		zap.String("api_url", cfg.APIURL),
		zap.String("redis_addr", cfg.RedisAddr),
	)

	<-ctx.Done()
	stop()

	if err := rdb.Close(); err != nil {
		logger.Error("redis close", zap.Error(err))
	}
	logger.Info("notifier stopped")
}

func buildHandler(
	api *apiclient.Client,
	dd *dedup.Deduplicator,
	senders map[domain.NotificationChannelKind]channel.Sender,
	logger *zap.Logger,
	subject string,
) gosrejs.HandlerFunc {
	return func(ctx context.Context, data []byte) error {
		var payload events.IncidentPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			logger.Error("unmarshal incident payload", zap.Error(err))
			return nil // ack malformed messages to prevent infinite redelivery
		}

		alreadySent, err := dd.HasBeenNotified(ctx, payload.IncidentID, subject)
		if err != nil {
			logger.Warn("dedup check failed; proceeding", zap.Error(err), zap.String("incident_id", payload.IncidentID))
		} else if alreadySent {
			logger.Debug("already notified, skipping",
				zap.String("incident_id", payload.IncidentID),
				zap.String("subject", subject),
			)
			return nil
		}

		rules, err := api.ListRulesByProject(ctx, payload.ProjectID)
		if err != nil {
			logger.Error("fetch rules", zap.Error(err), zap.String("project_id", payload.ProjectID))
			return err // nak: retry when api recovers
		}

		for _, rule := range rules {
			if rule.EventKind != subject {
				continue
			}
			ch, err := api.GetChannel(ctx, rule.ChannelID)
			if err != nil {
				logger.Error("fetch channel", zap.Error(err), zap.String("channel_id", rule.ChannelID))
				continue
			}
			sender, ok := senders[ch.Kind]
			if !ok {
				logger.Warn("unknown channel kind", zap.String("kind", string(ch.Kind)))
				continue
			}
			msg := buildMessage(payload, subject)
			if err := sendWithRetry(ctx, sender, ch, msg); err != nil {
				logger.Error("send notification",
					zap.Error(err),
					zap.String("channel_id", ch.ID),
					zap.String("kind", string(ch.Kind)),
					zap.String("incident_id", payload.IncidentID),
				)
				continue
			}
			logger.Info("notification sent",
				zap.String("channel_id", ch.ID),
				zap.String("kind", string(ch.Kind)),
				zap.String("incident_id", payload.IncidentID),
			)
		}

		if err := dd.MarkNotified(ctx, payload.IncidentID, subject); err != nil {
			logger.Error("mark notified", zap.Error(err), zap.String("incident_id", payload.IncidentID))
		}

		return nil
	}
}

func buildMessage(payload events.IncidentPayload, subject string) channel.Message {
	if subject == events.SubjectIncidentsOpened {
		return channel.Message{
			Subject: fmt.Sprintf("[INCIDENT OPENED] Target %s", payload.TargetID),
			Body: fmt.Sprintf("Incident %s opened at %s. State: %s.",
				payload.IncidentID,
				payload.FirstSeen.UTC().Format(time.RFC3339),
				payload.State,
			),
		}
	}
	return channel.Message{
		Subject: fmt.Sprintf("[INCIDENT RESOLVED] Target %s", payload.TargetID),
		Body: fmt.Sprintf("Incident %s resolved at %s. State: %s.",
			payload.IncidentID,
			payload.LastSeen.UTC().Format(time.RFC3339),
			payload.State,
		),
	}
}

func sendWithRetry(ctx context.Context, sender channel.Sender, ch domain.NotificationChannel, msg channel.Message) error {
	backoffs := []time.Duration{time.Second, 2 * time.Second}
	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoffs[attempt-1]):
			}
		}
		if err := sender.Send(ctx, ch, msg); err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt+1, err)
			continue
		}
		return nil
	}
	return lastErr
}
