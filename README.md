# Monorepo for leaning

~~Just some rpc learning~~

### Current stats

<img width="1270" height="629" alt="image" src="https://github.com/user-attachments/assets/6d5878cb-f22c-4960-8935-3a06249227b9" />

### Architecture & SRE Best Practices

- [ ] Design and document system architecture diagrams (service mesh, data flow, dependencies)
- [ ] Document SLOs, SLIs, and SLAs for all critical services
- [ ] Practice incident response
- [ ] Design for high availability and disaster recovery (multi-zone, backup/restore)
- [ ] Set up horizontal and vertical pod autoscaling in Kubernetes

### Infrastructure as Code & CI/CD

- [ ] Implement infrastructure as code (IaC) for all environments (e.g., Terraform, Pulumi)
- [ ] Set up automated CI/CD pipelines with rollback and canary deployment strategies
- [ ] Enforce configuration management and secrets management best practices
- [ ] Implement blue/green and canary deployments in Kubernetes

### Containerization

- [x] Use Wolfi images for secure, minimal containers
- [ ] Learn debug with attach functions
- [ ] Build and push images to a container registry

### Kubernetes & GitOps

- [x] Write Kustomize bases and overlays for all environments
- [x] Use CloudNativePG for Postgres operator in k8s
- [ ] Deploy and manage services with Kustomize/ArgoCD

### ArgoCD

- [x] Use ArgoCD for GitOps deployment and automated sync
- [ ] Implement canary deployment with ArgoCD
- [ ] Set up auto rollback based on Prometheus metrics
- [ ] Configure ArgoCD notifications and health checks
- [ ] Manage application lifecycle and sync policies with ArgoCD

### Service Mesh

- [x] Configure Istio for service mesh
- [x] Set up Istio ingress gateway and traffic management
- [x] **Set up Istio telemetry (metrics, logs, traces) **
  - [x] Complete Istio + OpenTelemetry + Jaeger integration
  - [x] Unified trace correlation between Istio sidecar and application
  - [x] B3 header propagation for service mesh compatibility
  - [x] End-to-end request tracing from ingress to database operations
- [ ] Implement canary deployment with Istio
- [ ] Configure Istio traffic shifting (weight-based routing, header-based routing)
- [ ] Enable Istio mTLS for service-to-service encryption
- [ ] Set up Istio ingress and egress policies
- [ ] Implement Istio RBAC for service-level access control
- [ ] Configure Istio retries, timeouts, and circuit breaking
- [ ] Use Istio for A/B testing and progressive delivery
- [ ] Document Istio configuration, policies, and best practices

### Database & Persistence

- [x] Write SQL migrations for user table
- [x] Use sqlc to generate Go database code
- [x] Implement repository pattern for DB access in Go
- [x] Integrate Postgres with CloudNativePG in Kubernetes

### Caching

- [x] Integrate Valkey (Redis-compatible) for caching
- [x] Use Valkey for user/session caching in the application

### RPC (gRPC & Protobuf)

- [x] Define service and messages in `.proto` files
- [x] Generate gRPC code for Go and Python
- [x] Implement gRPC server in Go
- [x] Implement async gRPC client in Python
- [x] Use gRPC stubs for client-server communication
- [x] CRUD operations via gRPC (Create, Read, Update, Delete, List)
- [x] Integrate gRPC with FastAPI endpoints
- [x] Error handling and health checks in gRPC clients
- [x] Use Protobuf for message serialization

### Python API

- [x] Use FastAPI for REST endpoints
- [x] Use Pydantic models for validation
- [x] Organize API endpoints with routers and services

### Monitoring & Observability (Prometheus & Grafana)

- [x] Set up Prometheus for metrics scraping
- [x] Configure Grafana dashboards for application and infrastructure
- [x] Add basic tracing (e.g., OpenTelemetry, Jaeger, or Zipkin)
- [x] Configure Prometheus to scrape Fluent Bit and PostgreSQL metrics
- [ ] Set up Grafana dashboards for SLOs, latency, error rates, and resource usage
- [ ] Configure alerting rules for SRE golden signals (latency, traffic, errors, saturation)
- [x] Correlate logs, metrics, and traces for root cause analysis

### Telemetry & Advanced Observability

- [x] **Implement distributed tracing (OpenTelemetry, Jaeger)**
  - [x] Full OpenTelemetry integration with OTLP gRPC exporter
  - [x] Service name consistency between Istio and application (`rpc-server.rpc`)
  - [x] B3 propagation for Istio/Jaeger trace context compatibility
  - [x] Unified trace correlation: Istio sidecar + application spans in same trace
- [x] **Instrument code for custom metrics, traces, and logs**
  - [x] gRPC server instrumentation with automatic span creation
  - [x] Database operation tracing (PostgreSQL queries with context keys)
  - [x] Cache operation tracing (Valkey Redis operations)
  - [x] Optimized bulk operations (single span for cache invalidation vs 20+ spans)
- [x] **Export application and infrastructure metrics to Prometheus**
- [x] **Integrate tracing context propagation across gRPC, HTTP, and async jobs**
  - [x] gRPC metadata propagation for distributed tracing
  - [x] HTTP header propagation through Istio service mesh
  - [x] Trace context preservation across service boundaries
- [x] **Document observability architecture and data flows**
  - [x] Complete request flow: Istio → gRPC → Database/Cache operations
  - [x] All spans correlated under single trace ID for end-to-end visibility
- [ ] Automate telemetry collection and export in CI/CD pipelines
- [ ] Evaluate and implement anomaly detection for proactive incident response
- [x] **End-to-End Trace Correlation**: Fixed separation between Istio and application traces
- [x] **Service Mesh Integration**: Full Istio + OpenTelemetry compatibility with B3 headers
- [x] **Database Instrumentation**: PostgreSQL query tracing with typed context keys
- [ ] **Observability Stack**: Complete integration with Jaeger, Prometheus, Grafana, and EFK

### EFK Stack (Elasticsearch, Fluent Bit, Kibana)

- [x] Deploy Elasticsearch cluster in Kubernetes using ECK operator
- [x] Set up Fluent Bit for log collection and shipping from pods
- [x] Configure Fluent Bit to parse and forward Kubernetes logs to Elasticsearch
- [x] Index application, infrastructure, and audit logs in Elasticsearch (`kubernetes-*` indices)
- [x] Set up Kibana for log visualization and search
- [x] Secure EFK stack with TLS and authentication (elastic user)
- [x] Configure Fluent Bit metrics endpoint for Prometheus scraping
- [ ] Design log retention, rollover, and archiving policies
- [ ] Create custom Kibana dashboards for application-specific logs
- [ ] Set up alerting in Kibana for error rates, anomalies, and critical events
- [x] Integrate EFK with Prometheus/Grafana for unified observability
- [x] Automate log pipeline deployment via ArgoCD and Kustomize

### Message Queue & Worker System

- [ ] Evaluate and select a message queue technology (e.g., RabbitMQ, NATS, Kafka, Redis Streams)
- [ ] Design message flow and worker architecture (producer, queue, consumer/worker)
- [ ] Implement message producer in application (e.g., enqueue jobs/events)
- [ ] Implement worker service to consume and process messages
- [ ] Ensure idempotency and error handling in worker logic
- [ ] Set up monitoring and alerting for queue depth, worker failures, and throughput
- [ ] Document message formats, queue topics, and retry/backoff strategies
- [ ] Integrate message queue with existing services (API, database, etc.)
- [ ] Automate deployment and scaling of worker services
- [ ] Test end-to-end message flow and failure scenarios

### Multilanguage Compile

- [ ] Compile at least one different language with Golang
