# Monorepo for leaning

~~Just some rpc learning~~

## 🏗️ Architecture Overview

<img width="1535" height="1292" alt="Untitled-2025-03-26-1749" src="https://github.com/user-attachments/assets/a0df4883-9c86-44ce-96ab-c98d26291158" />

### Core Services

- **gRPC Server** (`cmd/server`, `internal/`) - Go-based user service with PostgreSQL persistence
- **API Gateway** (`client/`) - Python FastAPI REST facade over gRPC
- **Database** - PostgreSQL with CloudNativePG operator and sqlc code generation
- **Cache** - Valkey (Redis-compatible) for session/user caching

### Infrastructure Stack

- **Kubernetes** - Local cluster via Orbstack with Kustomize manifests
- **Service Mesh** - Istio with mTLS, traffic management, and telemetry
- **GitOps** - ArgoCD for declarative deployments and automated sync
- **Observability** - Complete O11y stack with correlation:
  - **Metrics** - Prometheus + Grafana dashboards
  - **Traces** - OpenTelemetry + Jaeger with B3 propagation
  - **Logs** - EFK stack (Elasticsearch, Fluent Bit, Kibana)
- **Security** - Wolfi base images, Istio mTLS, cert-manager
- **Protocols** - gRPC with Protobuf, REST APIs, OpenTelemetry OTLP

### Architecture & SRE Best Practices

- [ ] Design and document system architecture diagrams (service mesh, data flow, dependencies)
- [ ] Document SLOs, SLIs, and SLAs for all critical services
- [ ] Practice incident response
- [ ] Design for high availability and disaster recovery (multi-zone, backup/restore)
- [ ] Set up horizontal and vertical pod autoscaling in Kubernetes

### SRE Platform Engineering Roadmap

#### Infrastructure & GitOps

- [x] Kustomize bases/overlays + ArgoCD GitOps deployment
- [x] Wolfi secure container images + CloudNativePG operator
- [ ] Infrastructure as code (Terraform/Pulumi) for multi-environment
- [ ] CI/CD pipelines with automated rollback and canary strategies
- [ ] Secrets management and configuration drift detection

#### Service Mesh & Traffic Management

- [x] Istio service mesh with mTLS and ingress gateway
- [x] **Complete OpenTelemetry + Jaeger integration with propagation**
- [x] **Unified trace correlation: Istio sidecar ↔ application spans**
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
- [ ] Cost optimization, capacity planning, and automated remediation
