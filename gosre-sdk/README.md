# gosre-sdk

Core domain types, interfaces, and store contracts for the GoSRE platform.

Imported by all other GoSRE repos. Has zero platform dependencies.

---

## Domain Model

### Target

Something to monitor: an HTTP URL, TCP endpoint, DNS name, or TLS host.

```go
type Target struct {
    ID       string
    Name     string
    Type     TargetType        // "http" | "tcp" | "dns" | "tls"
    Address  string
    Tags     []string
    Metadata map[string]string
}
```

### CheckConfig

A validation to run against a Target, on a given interval and timeout.

```go
type CheckConfig struct {
    ID       string
    Type     CheckType         // "http" | "tcp" | "dns" | "tls"
    TargetID string
    Interval time.Duration
    Timeout  time.Duration
    Params   map[string]string
}
```

### Result

Structured output of a single check execution.

```go
type Result struct {
    ID        string
    CheckID   string
    TargetID  string
    AgentID   string           // empty for local execution
    Status    CheckStatus      // "ok" | "fail" | "timeout" | "unknown"
    Duration  time.Duration    // serialized as duration_ms (milliseconds)
    Error     string
    Timestamp time.Time
    Metadata  map[string]string
}
```

### Incident

Derived operational state from repeated Result failures on a Target.

```go
type Incident struct {
    ID        string
    TargetID  string
    State     IncidentState    // "open" | "acknowledged" | "resolved"
    FirstSeen time.Time
    LastSeen  time.Time
    ResultIDs []string
}
```

---

## Checker Interface

Implemented by each check runner in `gosre-cli` and `gosre-agent`.

```go
type Checker interface {
    Execute(ctx context.Context, t Target, cfg CheckConfig) Result
}
```

`context.Context` is the first parameter on every method that does I/O — cancellation and timeout propagation are mandatory.

---

## Store Interfaces

Defined here so `gosre-api` and `gosre-agent` share a stable contract.
Implementations live in each repo's `internal/store/`.

```go
type TargetStore interface {
    Save(ctx context.Context, t domain.Target) error
    Get(ctx context.Context, id string) (domain.Target, error)
    List(ctx context.Context) ([]domain.Target, error)
    Delete(ctx context.Context, id string) error
}

type ResultStore interface {
    Save(ctx context.Context, r domain.Result) error
    Get(ctx context.Context, id string) (domain.Result, error)
    ListByTarget(ctx context.Context, targetID string) ([]domain.Result, error)
}

type IncidentStore interface {
    Save(ctx context.Context, i domain.Incident) error
    Get(ctx context.Context, id string) (domain.Incident, error)
    ListByState(ctx context.Context, state domain.IncidentState) ([]domain.Incident, error)
    Update(ctx context.Context, i domain.Incident) error
}
```

---

## Sentinel Errors

```go
var (
    ErrTargetNotFound   = errors.New("target not found")
    ErrCheckTimeout     = errors.New("check timed out")
    ErrIncidentNotFound = errors.New("incident not found")
)
```

Check with `errors.Is(err, domain.ErrTargetNotFound)` — never compare error strings directly.

---

## Install

```bash
go get github.com/gosre/gosre-sdk@latest
```

## License

Apache License 2.0 — Copyright 2026 Vinicius Teixeira
