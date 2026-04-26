// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package jetstream

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// HandlerFunc is called for each message received on a subscribed subject.
// Returning nil causes the message to be acknowledged; returning a non-nil error
// causes a negative acknowledgment so the message is redelivered.
type HandlerFunc func(ctx context.Context, data []byte) error

// Subscribe creates a durable pull consumer on subject and calls handler for each
// message. durableName uniquely identifies the consumer across restarts — use a
// stable, descriptive name (e.g. "incident-detector", "notifier-incidents-opened").
//
// Subscribe is non-blocking: it returns as soon as the consumer is created.
// A goroutine stops the consumer when ctx is cancelled.
func Subscribe(ctx context.Context, js jetstream.JetStream, subject, durableName string, handler HandlerFunc) error {
	cons, err := js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Name:          durableName,
		Durable:       durableName,
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverNewPolicy,
	})
	if err != nil {
		return fmt.Errorf("create consumer %s: %w", durableName, err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		if herr := handler(ctx, msg.Data()); herr != nil {
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("start consume for %s: %w", durableName, err)
	}

	go func() {
		<-ctx.Done()
		cc.Stop()
	}()

	return nil
}
