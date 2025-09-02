# AGENTS.md — Quick Guide for Agentic Coding

## Environment Intro

- Local Kubernetes via Orbstack; in-cluster services are reachable by service name, e.g., http://grafana.observability.svc.cluster.local
- App manifests: check `argocd/` directory for Kustomize configurations; check `~/app/arch-manifest` when examining app deployment structure
- Stack: Istio; PostgreSQL; Valkey; Grafana/Prometheus/Jaeger/OpenTelemetry; EFK; ArgoCD

## Code Style

### Go (rpc-server/)

- Module: `grpc-server`, Go 1.24; imports: stdlib first, external, then local `grpc-server/internal/...`
- Types: PascalCase structs/interfaces, camelCase fields, receiver names single letter of type
- Errors: return explicit errors, use `fmt.Errorf` for wrapping; log with structured `slog.Logger`

### Python (rpc-client/)

- Python 3.13+; FastAPI app structure; imports: stdlib, external, local relative `.`
- Style: `ruff format` (double quotes, 4-space indent); type hints with `pydantic` models
- Async: prefer async/await, use `asynccontextmanager` for lifespans
- Config: centralized in `core/config.py`; structured logging with stdlib `logging`
