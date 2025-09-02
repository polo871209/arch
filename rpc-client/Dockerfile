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

# Runtime stage
FROM cgr.dev/chainguard/wolfi-base:latest

RUN apk add --no-cache \
      ca-certificates \
      python-3.13 \
      tzdata

USER nonroot

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH="/app"

WORKDIR /app

COPY --from=builder /app/.venv /app/.venv

ENV PATH="/app/.venv/bin:$PATH"

COPY app/ ./app/
COPY proto/ ./proto/

EXPOSE 8000/tcp
