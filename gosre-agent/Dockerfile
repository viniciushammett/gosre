# Build context: ~/gosre (monorepo root)
# docker build -f gosre-agent/Dockerfile -t gosre-agent:local .

# Stage 1: build
FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY gosre-sdk/   /src/gosre-sdk/
COPY gosre-agent/ /src/gosre-agent/

WORKDIR /src/gosre-agent

RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -trimpath -ldflags="-s -w" -o /bin/gosre-agent ./cmd/agent

# Stage 2: runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/gosre-agent /gosre-agent

ENTRYPOINT ["/gosre-agent"]
