# Agent Development Guidelines

## Build/Test/Lint Commands
- Install deps: `just install` (Go mod tidy, uv sync)
- Generate proto: `just proto` (protobuf compilation)
- Start services: `just` or `just start` (default - both Go server and Python client)
- Individual servers: `just server` (Go foreground), `just client` (Python foreground)
- Test: `just test` (basic health check)
- Stop services: `just stop`
- View logs: `just logs`
- Clean up: `just clean`

## Project Structure
- Go 1.24 gRPC server in `go-server/` using structured logging (slog)
- Python 3.13 FastAPI client in `python-client/` with modern async patterns
- Protocol buffers in `proto/` - regenerate with `just generate-proto`

## Code Style Guidelines
- **Go**: Use `slog` for logging, `status.Errorf` for gRPC errors, proper mutex locking
- **Python**: Type hints with `|` union syntax, Pydantic models, async/await patterns
- **Imports**: Standard library first, third-party, then local imports
- **Error handling**: gRPC status codes in Go, HTTPException in Python FastAPI
- **Naming**: Snake_case for Python, camelCase for Go, descriptive variable names
- **Types**: Strict typing - use Pydantic models for API, proper Go struct tags