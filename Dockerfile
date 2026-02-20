# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY web/ ./web/

ARG VERSION=vDev
ARG BUILD_TIME=timeless
ARG COMMIT_HASH=sha-unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}" \
    -o stock-fetcher .

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 8080 appgroup && \
    adduser -D -u 8080 -G appgroup appuser

WORKDIR /app

# Create /data for cache volume
RUN mkdir -p /data && chown appuser:appgroup /data

COPY --from=builder /app/stock-fetcher .

USER 8080:8080

EXPOSE 8080

ENV PORT=8080

# /data is auto-detected by the app for cache.db
VOLUME ["/data"]

CMD ["./stock-fetcher"]
