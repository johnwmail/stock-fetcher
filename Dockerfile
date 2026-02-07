# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./
COPY web/ ./web/

# Build arguments for version info
ARG VERSION=vDev
ARG BUILD_TIME=timeless
ARG COMMIT_HASH=sha-unknown

# Build binary with version info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}" \
    -o stock-fetcher .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# Create non-root user with UID/GID 8080
RUN addgroup -g 8080 appgroup && \
    adduser -D -u 8080 -G appgroup appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/stock-fetcher .

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER 8080:8080

# Expose port
EXPOSE 8080

# Default to serve mode
ENV PORT=8080

CMD ["./stock-fetcher", "-serve"]
