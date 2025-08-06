# AGENTS.md - Guidelines for Agentic Coding in this Repository

## Architecture Overview

**Core Services:**

- `cmd/server/` - Go gRPC server (main service)
- `client/` - Python FastAPI client (HTTP gateway)
- `pkg/pb/` - Generated protobuf Go code
- `client/proto/` - Generated protobuf Python code

**Infrastructure Stack:**

- **Orbstack:** Deploy on Mac local kubernetes environment, direct access through service account are available. eg: http://grafana.observability.svc.cluster.local
- **Service Mesh:** Istio (traffic management, security, observability)
- **Database:** PostgreSQL (CloudNativePG operator)
- **Cache:** Valkey (Redis-compatible)
- **Observability:** Grafana, Prometheus, Jaeger, OpenTelemetry
- **Logging:** EFK stack (Elasticsearch, Fluent Bit, Kibana)
- **GitOps:** ArgoCD for continuous deployment

## Build/Test/Lint Commands

**Go (root directory):**

- Build: `go build ./...`
- Test: `go test ./...` (single package: `go test ./internal/server`)
- Lint: `golangci-lint run` or `go vet ./...`
- Generate code: `go generate ./...`
- SQL generation: `sqlc generate`

**Python (client/ directory):**

- Install deps: `cd client && uv sync`
- Lint: `cd client && uv run ruff check` (fix: `uv run ruff check --fix`)
- Format: `cd client && uv run ruff format`
- Test: `cd client && uv run pytest` (single test: `uv run pytest test_file.py::test_function`)

**Infrastructure:**

- Bootstrap infra: `just infra-bootstrap` (ArgoCD, namespaces)
- Deploy infra: `just infra` (all infrastructure components)
- App deploy: `just start` (builds & deploys app services)
- Migration: `just migration` (database migrations)
- Proto generation: `just proto`

**Kustomize Structure:**

- Base configs: `kustomize/base/app/` (rpc-server, rpc-client)
- Overlays: `kustomize/overlays/dev/` (environment-specific configs)
- Infrastructure: `infra/` (all platform components)

## Code Style Guidelines

**Go:**

- Use structured logging with `slog` and custom fields (`logging.UserID`, `logging.Error`)
- Error handling: Return `status.Errorf(codes.X, "msg")` for gRPC; use `repository.ErrX` constants
- Package structure: `internal/` for private code, `pkg/` for public
- Imports: std lib, external, local (grouped with blank lines)
- Naming: CamelCase for exported, camelCase for unexported
- Context: Always pass `context.Context` as first parameter

**Python:**

- Format with Ruff (double quotes, 4 spaces)
- Type hints required (Python 3.13+)
- Imports: follow `isort` style, exclude proto files from formatting
- Error handling: Use FastAPI `HTTPException` or custom exceptions
- Async/await for I/O operations
- Logging: Use structured logging with consistent formatter

