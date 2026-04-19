# gosre-charts

Helm charts for the GoSRE platform.

## Charts

| Chart | Description | Status |
|-------|-------------|--------|
| `gosre-api` | REST API server | ready |
| `gosre-agent` | Distributed check executor | stub |
| `gosre-exporter` | Prometheus metrics exporter | stub |
| `gosre-web` | React frontend | stub |
| `gosre` | Umbrella chart | stub |

## Usage

```bash
helm install gosre-api charts/gosre-api \
  --namespace gosre \
  --create-namespace \
  --set apiKey=your-api-key
```
