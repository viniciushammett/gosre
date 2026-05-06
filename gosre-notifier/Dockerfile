# Build context: ~/gosre (monorepo root)
# docker build -f gosre-notifier/Dockerfile -t gosre-notifier:local .

# Stage 1: build
FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY gosre-sdk/      /src/gosre-sdk/
COPY gosre-events/   /src/gosre-events/
COPY gosre-notifier/ /src/gosre-notifier/

WORKDIR /src/gosre-notifier

RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -trimpath -ldflags="-s -w" -o /bin/gosre-notifier ./cmd/notifier

# Stage 2: runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/gosre-notifier /gosre-notifier

USER nonroot:nonroot

ENTRYPOINT ["/gosre-notifier"]
