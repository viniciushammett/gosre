# Build context: ~/gosre (monorepo root)
# docker build -f gosre-auth/Dockerfile -t gosre-auth:local .

# Stage 1: build
FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY gosre-sdk/   /src/gosre-sdk/
COPY gosre-auth/  /src/gosre-auth/

WORKDIR /src/gosre-auth

RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -trimpath -ldflags="-s -w" -o /bin/gosre-auth ./cmd/gosre-auth

# Stage 2: runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/gosre-auth /gosre-auth

EXPOSE 8081

ENTRYPOINT ["/gosre-auth"]
