# AGENTS.md — Quick Guide for Agentic Coding

## Environment Intro

- Local Kubernetes via Orbstack; in-cluster services are reachable by service name, e.g., http://grafana.observability.svc.cluster.local
- App manifests: check `argos/` directory for Kustomize configurations; check `~/app/arch-manifest` when examining app deployment structure
- Stack: Istio; PostgreSQL; Valkey; Grafana/Prometheus/Jaeger/OpenTelemetry; EFK; ArgoCD

## Build/Lint/Test

- Go: build `go build ./...`; tests `go test ./...`; single pkg `go test ./internal/server`; single test `go test ./internal/server -run '^TestName$'`
- Go lint: `golangci-lint run` (or `go vet ./...`); SQL gen `sqlc generate`; Proto `just proto`
- Python (client/): deps `cd client && uv sync`; tests `uv run pytest`; single `uv run pytest client/path/test_x.py::test_y -q`
- Python lint/format: `uv run ruff check [--fix]`; `uv run ruff format`
- Infra: `just argos-bootstrap`; `just argos`; app `just local-build`; IaC `just kibana`
