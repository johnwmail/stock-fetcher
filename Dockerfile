# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./
COPY web/ ./web/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o stock-fetcher .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/stock-fetcher .

# Expose port
EXPOSE 8080

# Default to serve mode
ENV PORT=8080

CMD ["./stock-fetcher", "-serve"]
