# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

# Keep CA certificates available to the final image for TLS database connections.
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/opspilot-api ./cmd/server

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 opspilot \
    && adduser -S -D -H -u 10001 -G opspilot opspilot

COPY --from=builder /out/opspilot-api /usr/local/bin/opspilot-api

ENV APP_ENV=production \
    APP_VERSION=dev \
    GIN_MODE=release \
    PORT=8080

USER 10001:10001

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/live" || exit 1

ENTRYPOINT ["/usr/local/bin/opspilot-api"]
