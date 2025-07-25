# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/wolfi-base:latest AS builder

# Install build dependencies in single layer, sorted alphabetically
RUN apk add --no-cache \
      ca-certificates \
      py3.13-pip \
      python-3.13

# Set Python environment variables for build
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# Set working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY pyproject.toml uv.lock ./

# Install uv for faster dependency management
RUN pip install --no-cache-dir uv

# Install dependencies using uv sync
RUN uv sync --frozen

# Copy application code excluding unnecessary files
COPY app/ ./app/
COPY proto/ ./proto/

# Runtime stage
FROM cgr.dev/chainguard/wolfi-base:latest

# Install only runtime Python dependencies
RUN apk add --no-cache \
      ca-certificates \
      python-3.13 \
      tzdata

# Set Python environment variables for runtime
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH="/app"

# Create non-root user with specific UID for consistency
RUN adduser -D -u 1001 -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy virtual environment from builder stage
COPY --from=builder /app/.venv /app/.venv

# Update PATH to use virtual environment
ENV PATH="/app/.venv/bin:$PATH"

# Copy application code
COPY --from=builder --chown=appuser:appuser /app/app/ ./app/
COPY --from=builder --chown=appuser:appuser /app/proto/ ./proto/

# Switch to non-root user
USER appuser

# Expose FastAPI port with protocol specification
EXPOSE 8000/tcp

# Health check with proper error handling
HEALTHCHECK --interval=30s \
            --timeout=5s \
            --start-period=15s \
            --retries=3 \
            CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health', timeout=3)" || exit 1

# Set entrypoint and default command
ENTRYPOINT ["python", "-m", "uvicorn"]
CMD ["app.main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "1"]
