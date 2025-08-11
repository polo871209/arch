# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/wolfi-base:latest AS builder

RUN apk add --no-cache \
      ca-certificates \
      py3.13-pip \
      python-3.13 \
      uv

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

WORKDIR /app

COPY pyproject.toml uv.lock ./

RUN uv sync --frozen

COPY app/ ./app/
COPY proto/ ./proto/

# Runtime stage
FROM cgr.dev/chainguard/wolfi-base:latest

RUN apk add --no-cache \
      ca-certificates \
      python-3.13 \
      tzdata

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH="/app"

RUN adduser -D -u 1001 -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /app/.venv /app/.venv

ENV PATH="/app/.venv/bin:$PATH"

COPY --from=builder --chown=appuser:appuser /app/app/ ./app/
COPY --from=builder --chown=appuser:appuser /app/proto/ ./proto/

USER appuser

EXPOSE 8000/tcp

