# AGENTS.md — Quick Guide for Agentic Coding

## Environment Intro
- Local Kubernetes via Orbstack; in-cluster services are reachable by service name (and full DNS), e.g., http://grafana or http://grafana.observability.svc.cluster.local
- Stack: Istio; PostgreSQL (CloudNativePG); Valkey; Grafana/Prometheus/Jaeger/OpenTelemetry; EFK; ArgoCD

## Build/Lint/Test
- Go: build `go build ./...`; tests `go test ./...`; single pkg `go test ./internal/server`; single test `go test ./internal/server -run '^TestName$'`
- Go lint: `golangci-lint run` (or `go vet ./...`); SQL gen `sqlc generate`; Proto `just proto`
- Python (client/): deps `cd client && uv sync`; tests `uv run pytest`; single `uv run pytest client/path/test_x.py::test_y -q`
- Python lint/format: `uv run ruff check [--fix]`; `uv run ruff format`
- Infra: `just infra-bootstrap`; `just infra`; app `just start`; migration `just migration`

## Code Style — Go
- Imports: stdlib, external, local (blank lines); layout: `internal/` private, `pkg/` public; context.Context first param
- Errors: gRPC with `status.Errorf(codes.X, ...)`; prefer `repository.Err*`; wrap with `%w`; logging via `slog` (e.g., `logging.UserID`, `logging.Error`)
- Naming: CamelCase exported, camelCase unexported

## Code Style — Python
- Ruff formatting (double quotes, 4 spaces); Python >= 3.13; use type hints
- Imports: isort order (stdlib, third-party, local); exclude `client/proto/` from lint/format
- Errors/IO: raise FastAPI `HTTPException` or custom; avoid bare `except`; use `async`/`await`; structured logging

## Editor/AI Rules
- No Cursor (.cursor/rules, .cursorrules) or Copilot (.github/copilot-instructions.md) rules found; if added, follow alongside this guide.
