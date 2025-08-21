# Monorepo for learning

## 🏗️ Architecture Overview

<img width="2646" height="1272" alt="image" src="https://github.com/user-attachments/assets/ae3d23a0-8d73-4e48-88a1-bac8d6903b2d" />

### Core Services

- **gRPC Server** (`cmd/server`, `internal/`) - Go-based user service with PostgreSQL persistence
- **API Gateway** (`client/`) - Python FastAPI REST facade over gRPC
- **Database** - PostgreSQL with CloudNativePG operator and sqlc code generation
- **Cache** - Valkey (Redis-compatible) for session/user caching

### Infrastructure Stack

- **Kubernetes** - Local cluster via Orbstack with GitOps deployment
- **Service Mesh** - Istio with mTLS, traffic management, and telemetry
- **GitOps** - ArgoCD for declarative deployments and automated sync
  - **Rollouts** - Argo rollouts with canary deployment
- **Observability** - Complete O11y stack with correlation:
  - **Metrics** - Prometheus + Grafana dashboards
  - **Traces** - OpenTelemetry + Jaeger with B3 propagation
  - **Logs** - EFK stack (Elasticsearch, Fluent Bit, Kibana)
- **Security** - Wolfi base images, Istio mTLS, cert-manager
- **Protocols** - gRPC with Protobuf, REST APIs, OpenTelemetry OTLP

## CI/CD & GitOps

### GitOps Architecture

- **Main Repository** (this repo) - Source code, Dockerfiles, CI/CD pipelines
- **Manifest Repository** - [arch-manifest](https://github.com/polo871209/arch-manifest) - Kubernetes manifests managed by ArgoCD

### Deployment Flow

1. **Code Push** → GitHub Actions builds Docker images → DockerHub
2. **Automated Update** → GitHub Actions updates image tags in manifest repo using `kustomize edit`
3. **GitOps Sync** → ArgoCD detects changes and deploys to Kubernetes cluster

### SRE Platform Engineering Roadmap

#### Infrastructure & GitOps

- [x] Kustomize bases/overlays + ArgoCD GitOps deployment
- [x] **CI/CD pipelines with automated image updates via kustomize edit**
- [x] **Separate manifest repository for GitOps workflow**
- [x] Wolfi secure container images + CloudNativePG operator
- [ ] Infrastructure as code (Terraform/Pulumi) for multi-environment
- [x] Canary deployments with automated rollback with analysis
- [ ] Secrets management and configuration drift detection

#### Service Mesh & Traffic Management

- [x] Istio service mesh with mTLS and ingress gateway
- [x] **Complete OpenTelemetry + Jaeger integration with propagation**
- [x] **Unified trace correlation: Istio sidecar ↔ application spans**
- [x] **Canary deployments with Argo Rollouts and traffic analysis**
- [ ] Canary deployments with traffic shifting (weight/header-based routing)
- [ ] Circuit breakers, retries, timeouts, and RBAC policies
- [ ] Fault injection for chaos engineering

#### Observability (SRE Golden Signals: LETS)

- [x] **Prometheus + Grafana dashboards for infrastructure/applications**
- [x] **EFK stack: Elasticsearch (ECK), Fluent Bit, Kibana with TLS**
- [x] **Distributed tracing with end-to-end correlation (gRPC → DB/Cache)**
- [ ] **SLI/SLO monitoring with error budgets and alerting**
- [ ] **Golden signals alerting (Latency, Errors, Traffic, Saturation)**
- [ ] Anomaly detection and synthetic monitoring

#### Data Platform & Persistence

- [x] **SQL migrations + sqlc code generation + repository pattern**
- [x] **PostgreSQL with CloudNativePG operator in Kubernetes**
- [x] **Valkey (Redis) caching with TTL policies and operation tracing**
- [ ] Database backup/restore, HA failover, and performance tuning
- [ ] Cache clustering and warming strategies

#### Application Services

- [x] **gRPC services with Protobuf + FastAPI REST gateway**
- [x] **Async gRPC clients with comprehensive error handling**
- [ ] Message queue system (NATS/Kafka/Redis Streams) with workers
- [ ] API versioning, authentication, and rate limiting
- [ ] Event-driven architecture with idempotency patterns

#### Reliability & Operations

- [ ] **Chaos engineering with Litmus/Chaos Mesh**
- [ ] **Load testing, disaster recovery automation**
- [ ] **Incident response with PagerDuty/Opsgenie integration**
- [ ] Policy as code (OPA/Gatekeeper) and compliance scanning
