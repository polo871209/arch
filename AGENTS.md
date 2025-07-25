# Agent Development Guidelines

## Build/Test/Lint Commands
- Install deps: `just install` (Go mod tidy, uv sync)
- Generate proto: `just proto` (protobuf compilation)
- Generate SQL: `just sqlc` (generate SQL code with sqlc)
- Database: `just db-up`, `just db-down`, `just db-reset` (PostgreSQL migrations)
- Start services: `just` or `just start` (default - both Go server and Python client with PostgreSQL)
- Individual servers: `just server` (Go foreground), `just client` (Python foreground)
- Test: `just test` (basic health check), specific tests: `cd client && uv run pytest path/to/test.py::test_name`
- Cache: `just cache-flush` (clear cache), `just cache-stats` (view cache info)
- Build: `just build` (Go server binary)
- Format: `just fmt` (Go fmt, black, isort)
- Lint: `just lint` (Go vet, flake8)
- Stop services: `just stop`
- View logs: `just logs`, status: `just status`
- Clean up: `just clean`

## Project Structure
- Go 1.24 gRPC server in root dir using structured logging (slog) with PostgreSQL/sqlc
- Python 3.13 FastAPI client in `client/` with modern async patterns and type hints
- Protocol buffers in `proto/` - regenerate with `just proto`
- Database: PostgreSQL with migrations in `internal/database/migrations/`

## Code Style Guidelines
- **Go**: Use `slog` for logging, `status.Errorf` for gRPC errors, proper error handling
- **Python**: Type hints with `|` union syntax, Pydantic models, async/await patterns
- **Imports**: Standard library first, third-party, then local imports (grouped with blank lines)
- **Error handling**: gRPC status codes in Go, HTTPException in Python FastAPI
- **Naming**: snake_case for Python, camelCase for Go, descriptive variable names
- **Types**: Strict typing - use Pydantic models for API, proper Go struct tags
- **Formatting**: Use `just fmt` before committing - black (line length 88), isort for Python