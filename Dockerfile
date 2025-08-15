# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/wolfi-base:latest AS builder

RUN apk add --no-cache \
      ca-certificates \
      file \
      git \
      go

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

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

RUN file server | grep -q "statically linked"

# Runtime stage
FROM cgr.dev/chainguard/wolfi-base:latest

USER nonroot

WORKDIR /app

COPY --from=builder /app/server /usr/local/bin/server

EXPOSE 50051
