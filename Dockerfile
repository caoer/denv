# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o denv ./cmd/denv

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates git bash

# Create non-root user
RUN addgroup -g 1000 denv && \
    adduser -D -u 1000 -G denv denv

# Copy binary from builder
COPY --from=builder /build/denv /usr/local/bin/denv

# Set ownership
RUN chown -R denv:denv /usr/local/bin/denv && \
    chmod +x /usr/local/bin/denv

# Switch to non-root user
USER denv

# Set working directory
WORKDIR /workspace

# Set entrypoint
ENTRYPOINT ["denv"]

# Default command
CMD ["--help"]