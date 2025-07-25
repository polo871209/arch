# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/wolfi-base:latest AS builder

# Install build dependencies in single layer, sorted alphabetically
RUN apk add --no-cache \
      ca-certificates \
      file \
      git \
      go

# Set working directory with absolute path
WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies with verification
RUN go mod download && \
    go mod verify

# Copy source code excluding unnecessary files 
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build with security and optimization flags
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
      -a \
      -installsuffix cgo \
      -ldflags='-w -s -extldflags "-static"' \
      -tags netgo \
      -o server \
      ./cmd/server

# Verify binary is statically linked
RUN file server | grep -q "statically linked"

# Runtime stage 
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder stage
COPY --from=builder /app/server /usr/local/bin/server

# Use distroless nonroot user (uid 65532)
USER 65532:65532

# Expose gRPC port with documentation
EXPOSE 50051/tcp

# Health check with proper timeout
HEALTHCHECK --interval=30s \
            --timeout=5s \
            --start-period=10s \
            --retries=3 \
            CMD ["/usr/local/bin/server", "--health-check"]

# Set entrypoint and default command
ENTRYPOINT ["/usr/local/bin/server"]
CMD ["--config", "/etc/server/config.yaml"]
