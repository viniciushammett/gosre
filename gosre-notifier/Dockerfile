FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o notifier ./cmd/notifier

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/notifier /app/notifier
USER nonroot:nonroot
ENTRYPOINT ["/app/notifier"]
