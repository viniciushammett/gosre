# Build context: ~/gosre (monorepo root)
# docker build -f gosre-scheduler/Dockerfile -t gosre-scheduler:local .

# Stage 1: build
FROM golang:1.26-bookworm AS builder

WORKDIR /src

# Copy local replace directive dependencies alongside the module being built.
COPY gosre-sdk/       /src/gosre-sdk/
COPY gosre-events/    /src/gosre-events/
COPY gosre-scheduler/ /src/gosre-scheduler/

WORKDIR /src/gosre-scheduler

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/gosre-scheduler ./cmd/server

# Stage 2: runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/gosre-scheduler /gosre-scheduler

ENTRYPOINT ["/gosre-scheduler"]
