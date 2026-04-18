# Build context: ~/gosre (monorepo root)
# docker build -f gosre-exporter/Dockerfile -t gosre-exporter:local .

# Stage 1: build
FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY gosre-sdk/       /src/gosre-sdk/
COPY gosre-exporter/  /src/gosre-exporter/

WORKDIR /src/gosre-exporter

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/gosre-exporter ./cmd/exporter

# Stage 2: runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/gosre-exporter /gosre-exporter

EXPOSE 9090

ENTRYPOINT ["/gosre-exporter"]
