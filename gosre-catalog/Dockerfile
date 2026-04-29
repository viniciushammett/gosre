# Build context: ~/gosre (monorepo root)
# docker build -f gosre-catalog/Dockerfile -t gosre-catalog:local .

# Stage 1: build
FROM golang:1.26-bookworm AS builder

WORKDIR /src

# go-sqlite3 requires CGO; copy both modules to satisfy replace directive
COPY gosre-sdk/     /src/gosre-sdk/
COPY gosre-catalog/ /src/gosre-catalog/

WORKDIR /src/gosre-catalog

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/gosre-catalog ./cmd/server

# Stage 2: runtime
FROM gcr.io/distroless/cc-debian12:nonroot

COPY --from=builder /bin/gosre-catalog /gosre-catalog

EXPOSE 8081

ENTRYPOINT ["/gosre-catalog"]
