# AGENTS.md - Guidelines for Agentic Coding in this Repository

## Build/Test/Lint Commands

**Go (root directory):**
- Build: `go build ./...`
- Test: `go test ./...` (single package: `go test ./internal/server`)
- Lint: `golangci-lint run` or `go vet ./...`
- Generate code: `go generate ./...`

**Python (client/ directory):**
- Install deps: `cd client && uv sync`
- Lint: `cd client && uv run ruff check` (fix: `uv run ruff check --fix`)
- Format: `cd client && uv run ruff format`
- Test: `cd client && uv run pytest` (single test: `uv run pytest test_file.py::test_function`)

**Docker/K8s:**
- Full deploy: `just start` (builds & deploys to k8s)
- Proto generation: `just proto`

## Code Style Guidelines

**Go:**
- Use structured logging with `slog` and custom fields (`logging.UserID`, `logging.Error`)
- Error handling: Return `status.Errorf(codes.X, "msg")` for gRPC; use `repository.ErrX` constants
- Package structure: `internal/` for private code, `pkg/` for public
- Imports: std lib, external, local (grouped with blank lines)
- Naming: CamelCase for exported, camelCase for unexported

**Python:**
- Format with Ruff (double quotes, 4 spaces)
- Type hints required (Python 3.13+)
- Imports: follow `isort` style
- Error handling: Use FastAPI `HTTPException` or custom exceptions
- Async/await for I/O operations