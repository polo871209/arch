# Monorepo: Python Client & Go Server

This repository contains a Python FastAPI client and a Go gRPC server, with supporting infrastructure for local development and deployment.

## Build & Development Commands

- **Build all services:** `just build`
- **Run DB migration:** `just migration`
- **Start full stack (dev):** `just start`
- **Python install deps:** `cd client && uv sync`
- **Python lint:** `cd client && uv run black . && uv run isort . && uv run flake8 .`
- **Python type check:** `cd client && uv run pyright`
- **Go build:** `go build ./cmd/server`

## Code Style Guidelines

- **Python:**
  - Use `black` and `isort` for formatting.
  - Type annotations for all public functions/classes.
  - Use Pydantic models for API schemas.
  - Use `async`/`await` for FastAPI endpoints and gRPC calls.
  - Use snake_case for functions/variables, PascalCase for classes.
- **Go:**
  - Use struct types for config, models, and repositories.
  - Return errors explicitly; wrap with context where helpful.
  - Use slog for logging.
  - Use PascalCase for exported names, camelCase for locals.
- **General:**
  - Follow existing patterns in each language and directory.
  - Keep code idiomatic and well-documented.

## Project Structure

- `client/` — Python FastAPI client
- `cmd/server/` — Go gRPC server
- `internal/` — Go internal packages
- `proto/` — Protobuf definitions and generated code
- `infra/`, `kustomize/` — Kubernetes, ArgoCD, Istio, and Postgres manifests

For more, see [AGENTS.md](./AGENTS.md).
