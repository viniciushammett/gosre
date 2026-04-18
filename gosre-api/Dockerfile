# Build context: ~/gosre (monorepo root)
# docker build -f gosre-api/Dockerfile -t gosre-api:local .

# Stage 1: build
FROM golang:1.25-bookworm AS builder

WORKDIR /src

# go-sqlite3 requires CGO; copy both modules to satisfy replace directive
COPY gosre-sdk/  /src/gosre-sdk/
COPY gosre-api/  /src/gosre-api/

WORKDIR /src/gosre-api

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/gosre-api ./cmd/server

# Stage 2: runtime
FROM gcr.io/distroless/cc-debian12:nonroot

COPY --from=builder /bin/gosre-api /gosre-api

EXPOSE 8080

ENTRYPOINT ["/gosre-api"]
