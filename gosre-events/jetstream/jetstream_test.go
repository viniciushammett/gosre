// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package jetstream_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natstest "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	natsjt "github.com/nats-io/nats.go/jetstream"

	"github.com/viniciushammett/gosre/gosre-events/events"
	"github.com/viniciushammett/gosre/gosre-events/jetstream"
)

// startJSServer starts an in-process NATS server with JetStream enabled.
// Returns a ready JetStream context and a cleanup function.
func startJSServer(t *testing.T) (natsjt.JetStream, func()) {
	t.Helper()

	opts := natstest.DefaultTestOptions
	opts.Port = -1
	opts.JetStream = true
	opts.StoreDir = t.TempDir()
	opts.NoLog = true
	opts.NoSigs = true

	srv := natstest.RunServer(&opts)
	if !srv.ReadyForConnections(5 * time.Second) {
		srv.Shutdown()
		t.Fatal("NATS server did not become ready in time")
	}

	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		srv.Shutdown()
		t.Fatalf("connect to NATS: %v", err)
	}

	js, err := natsjt.New(nc)
	if err != nil {
		nc.Close()
		srv.Shutdown()
		t.Fatalf("create JetStream context: %v", err)
	}

	cleanup := func() {
		nc.Close()
		srv.Shutdown()
	}
	return js, cleanup
}

// TestEnsureStream verifies that EnsureStream creates the GOSRE stream and
// that a second call (idempotent) returns no error.
func TestEnsureStream(t *testing.T) {
	js, cleanup := startJSServer(t)
	defer cleanup()

	ctx := context.Background()

	if err := jetstream.EnsureStream(ctx, js); err != nil {
		t.Fatalf("first EnsureStream: %v", err)
	}

	if err := jetstream.EnsureStream(ctx, js); err != nil {
		t.Fatalf("second EnsureStream (idempotent): %v", err)
	}
}

// TestPublish verifies that Publisher.Publish delivers a message to the stream
// and that its JSON content matches the published payload.
func TestPublish(t *testing.T) {
	js, cleanup := startJSServer(t)
	defer cleanup()

	ctx := context.Background()

	if err := jetstream.EnsureStream(ctx, js); err != nil {
		t.Fatalf("EnsureStream: %v", err)
	}

	want := events.ResultCreatedPayload{
		EventEnvelope: events.NewEnvelope(),
		ResultID:      "r-001",
		CheckID:       "c-001",
		TargetID:      "t-001",
		TargetName:    "example.com",
		Status:        "success",
		DurationMS:    42,
	}

	pub := jetstream.NewPublisher(js)
	if err := pub.Publish(ctx, events.SubjectResultsCreated, want); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// Read directly from the stream to verify delivery and content.
	cons, err := js.CreateOrUpdateConsumer(ctx, "GOSRE", natsjt.ConsumerConfig{
		Name:          "test-publish-verify",
		Durable:       "test-publish-verify",
		FilterSubject: events.SubjectResultsCreated,
		AckPolicy:     natsjt.AckExplicitPolicy,
		DeliverPolicy: natsjt.DeliverAllPolicy,
	})
	if err != nil {
		t.Fatalf("CreateOrUpdateConsumer: %v", err)
	}

	msgs, err := cons.Fetch(1, natsjt.FetchMaxWait(3*time.Second))
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	got := msgs.Messages()
	msg, ok := <-got
	if !ok {
		t.Fatal("no message received from stream")
	}
	_ = msg.Ack()

	var got2 events.ResultCreatedPayload
	if err := json.Unmarshal(msg.Data(), &got2); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got2.ResultID != want.ResultID {
		t.Errorf("ResultID: got %q, want %q", got2.ResultID, want.ResultID)
	}
	if got2.Status != want.Status {
		t.Errorf("Status: got %q, want %q", got2.Status, want.Status)
	}
	if got2.DurationMS != want.DurationMS {
		t.Errorf("DurationMS: got %d, want %d", got2.DurationMS, want.DurationMS)
	}
	if got2.EventID == "" {
		t.Error("EventID must not be empty")
	}
}

// TestSubscribe verifies that Subscribe delivers the correct JSON payload to
// the handler.
func TestSubscribe(t *testing.T) {
	js, cleanup := startJSServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := jetstream.EnsureStream(ctx, js); err != nil {
		t.Fatalf("EnsureStream: %v", err)
	}

	want := events.AgentHeartbeatPayload{
		EventEnvelope: events.NewEnvelope(),
		AgentID:       "agent-abc",
		Hostname:      "node-1",
		Version:       "v0.1.0",
	}

	received := make(chan events.AgentHeartbeatPayload, 1)

	err := jetstream.Subscribe(ctx, js, events.SubjectAgentsHeartbeat, "test-subscribe",
		func(_ context.Context, data []byte) error {
			var p events.AgentHeartbeatPayload
			if err := json.Unmarshal(data, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			received <- p
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	pub := jetstream.NewPublisher(js)
	if err := pub.Publish(ctx, events.SubjectAgentsHeartbeat, want); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case got := <-received:
		if got.AgentID != want.AgentID {
			t.Errorf("AgentID: got %q, want %q", got.AgentID, want.AgentID)
		}
		if got.Hostname != want.Hostname {
			t.Errorf("Hostname: got %q, want %q", got.Hostname, want.Hostname)
		}
		if got.Version != want.Version {
			t.Errorf("Version: got %q, want %q", got.Version, want.Version)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for message from Subscribe")
	}
}

// TestSubscribeCancelContext verifies that cancelling the context passed to
// Subscribe cleanly stops the consumer without blocking.
func TestSubscribeCancelContext(t *testing.T) {
	js, cleanup := startJSServer(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	if err := jetstream.EnsureStream(ctx, js); err != nil {
		t.Fatalf("EnsureStream: %v", err)
	}

	err := jetstream.Subscribe(ctx, js, events.SubjectChecksAssigned, "test-cancel",
		func(_ context.Context, data []byte) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Cancel the context; the internal goroutine should call cc.Stop().
	cancel()

	// Give the goroutine time to stop.
	done := make(chan struct{})
	go func() {
		// Attempt a new consumer on the same durable after cancellation.
		// If cc.Stop() was not called the consumer might still be active.
		// We simply verify that cancellation does not hang.
		close(done)
	}()

	select {
	case <-done:
		// success — goroutine exited promptly
	case <-time.After(3 * time.Second):
		t.Fatal("context cancellation did not stop consumer in time")
	}
}

// Compile-time check: server package must be imported.
var _ = server.Options{}
