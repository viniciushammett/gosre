# gosre-exporter

Prometheus exporter for the GoSRE platform. Scrapes metrics from
[gosre-api](../gosre-api) and exposes them at `/metrics`.

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `gosre_target_up` | Gauge | `target_id`, `target_name`, `target_type` | 1 if the latest check result was ok, 0 otherwise |
| `gosre_check_result_total` | Counter | `target_id`, `check_type`, `status` | Total check results by status |
| `gosre_check_duration_seconds` | Histogram | `target_id`, `check_type` | Check execution duration |
| `gosre_incident_total` | Gauge | `target_id`, `state` | Incidents per target and state |
| `gosre_agent_up` | Gauge | — | Agent availability (stub, Phase 4) |

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOSRE_API_URL` | ✅ | — | gosre-api base URL (e.g. `http://localhost:8080`) |
| `GOSRE_API_KEY` | ❌ | — | API key for the `X-API-Key` header |
| `GOSRE_POLL_INTERVAL` | ❌ | `30s` | API call timeout per scrape (match your Prometheus scrape interval) |
| `GOSRE_METRICS_PORT` | ❌ | `9090` | Port to expose `/metrics` on |

## Running Locally

```bash
GOSRE_API_URL=http://localhost:8080 go run ./cmd/exporter
```

## Docker

```bash
# Build from monorepo root
docker build -f gosre-exporter/Dockerfile -t gosre-exporter:local .

# Run
docker run --rm \
  -e GOSRE_API_URL=http://host.docker.internal:8080 \
  -p 9090:9090 \
  gosre-exporter:local
```

## Prometheus Scrape Config

```yaml
scrape_configs:
  - job_name: gosre
    scrape_interval: 30s
    static_configs:
      - targets: ["localhost:9090"]
```
