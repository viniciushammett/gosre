# gosre-events

NATS JetStream wrapper for the GoSRE event bus.

## Overview

All GoSRE events flow through a single JetStream stream named `GOSRE`
covering subjects `gosre.>`.

## Event subjects

| Subject | Published by | Consumed by |
|---------|-------------|-------------|
| `gosre.results.created` | gosre-api (result handler) | incident detector, gosre-notifier |
| `gosre.incidents.opened` | gosre-api (incident service) | gosre-notifier |
| `gosre.incidents.resolved` | gosre-api (incident service) | gosre-notifier |
| `gosre.agents.heartbeat` | gosre-agent (heartbeat goroutine) | gosre-scheduler |
| `gosre.checks.assigned` | gosre-scheduler | gosre-agent |

## Usage

```go
import (
    "github.com/nats-io/nats.go"
    natsjt "github.com/nats-io/nats.go/jetstream"
    "github.com/viniciushammett/gosre/gosre-events/events"
    "github.com/viniciushammett/gosre/gosre-events/jetstream"
)

nc, _ := nats.Connect(nats.DefaultURL)
js, _ := natsjt.New(nc)

// Ensure stream exists (idempotent — call on startup).
_ = jetstream.EnsureStream(ctx, js)

// Publish
pub := jetstream.NewPublisher(js)
_ = pub.Publish(ctx, events.SubjectResultsCreated, events.ResultCreatedPayload{
    EventEnvelope: events.NewEnvelope(),
    ResultID:      result.ID,
    // ...
})

// Subscribe
_ = jetstream.Subscribe(ctx, js, events.SubjectResultsCreated, "incident-detector",
    func(ctx context.Context, data []byte) error {
        var p events.ResultCreatedPayload
        return json.Unmarshal(data, &p)
    },
)
```
