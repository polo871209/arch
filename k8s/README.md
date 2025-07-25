# RPC Application Kubernetes Deployment with Kustomize

This directory contains Kustomize configurations for deploying the RPC application (Go gRPC server + Python FastAPI client) to Kubernetes.

## Directory Structure

```
k8s/
├── base/                           # Base Kubernetes resources
│   ├── namespace.yaml             # RPC namespace
│   ├── postgres.yaml              # PostgreSQL database
│   ├── valkey.yaml                # Valkey cache
│   ├── migration.yaml             # Database migration job
│   ├── grpc-server.yaml           # Go gRPC server
│   ├── fastapi-client.yaml        # Python FastAPI client
│   └── kustomization.yaml         # Base kustomization
├── overlays/
│   └── development/               # Development environment
│       ├── kustomization.yaml     # Dev overlay
│       └── deployment-patches.yaml # Resource adjustments for dev
├── deploy.sh                      # Automated deployment script
└── README.md                      # This file
```

## Prerequisites

- **Docker**: For building and pushing images
- **kubectl**: Kubernetes CLI tool
- **kustomize**: For managing Kubernetes configurations
- **Kubernetes cluster**: OrbStack, minikube, or any K8s cluster
- **Container registry**: Default is `localhost:5000`

## Quick Start

### 1. Automated Deployment

Run the deployment script from the project root:

```bash
./k8s/deploy.sh
```

This script will:
1. Build and push Docker images (`rpc-server`, `rpc-client`, `rpc-migration`)
2. Deploy infrastructure (PostgreSQL, Valkey)
3. Run database migrations using goose
4. Deploy applications (gRPC server, FastAPI client)
5. Verify deployment status

### 2. Manual Deployment

If you prefer manual control:

```bash
# Build and push images first
just build
docker build -t localhost:5000/rpc-server:latest .
docker build -t localhost:5000/rpc-client:latest -f client/Dockerfile client/
docker build -t localhost:5000/rpc-migration:latest -f Dockerfile.migration .
docker push localhost:5000/rpc-server:latest
docker push localhost:5000/rpc-client:latest  
docker push localhost:5000/rpc-migration:latest

# Deploy with kustomize
kubectl apply -k k8s/overlays/development
```

## Application Components

### Infrastructure
- **PostgreSQL 17**: Primary database with persistent storage (1Gi)
- **Valkey 8.0**: Redis-compatible cache with persistent storage (512Mi)

### Applications
- **gRPC Server**: Go application exposing user management API on port 50051
- **FastAPI Client**: Python web API frontend on port 8000 with LoadBalancer service
- **Migration Job**: Goose-based database migration runner

### Secrets
- `postgres-secret`: Database credentials and connection URL
- `grpc-server-secret`: Server configuration (DB and cache URLs)
- `fastapi-client-secret`: Client configuration (gRPC connection details)

## Accessing the Application

### FastAPI Client (Web API)
- **Health check**: `http://<EXTERNAL-IP>:8000/health`
- **API docs**: `http://<EXTERNAL-IP>:8000/docs`
- **User endpoints**: `http://<EXTERNAL-IP>:8000/api/v1/users`

### Local Development Access
If LoadBalancer IP is not available:
```bash
kubectl port-forward -n rpc service/dev-fastapi-client-service 8000:8000
```
Then access at `http://localhost:8000`

## Useful Commands

### Monitoring
```bash
# Check all resources
kubectl get all -n rpc

# View application logs
kubectl logs -n rpc -l app=dev-grpc-server --tail=100 -f
kubectl logs -n rpc -l app=dev-fastapi-client --tail=100 -f

# Check migration job logs
kubectl logs -n rpc job/dev-db-migration
```

### Scaling
```bash
# Scale gRPC server
kubectl scale -n rpc deployment/dev-grpc-server --replicas=3

# Scale FastAPI client
kubectl scale -n rpc deployment/dev-fastapi-client --replicas=2
```

### Updates
```bash
# Restart deployments (useful after image changes)
kubectl rollout restart -n rpc deployment/dev-grpc-server
kubectl rollout restart -n rpc deployment/dev-fastapi-client

# Update with new images
kustomize build k8s/overlays/development | kubectl apply -f -
```

### Database Access
```bash
# Connect to PostgreSQL
kubectl exec -it -n rpc deployment/dev-postgres -- psql -U rpc_user -d rpc_dev

# Run database commands
kubectl exec -it -n rpc deployment/dev-postgres -- pg_dump -U rpc_user rpc_dev
```

### Debugging
```bash
# Get detailed information about resources
kubectl describe -n rpc deployment/dev-grpc-server
kubectl describe -n rpc pod/<pod-name>

# Check events
kubectl get events -n rpc --sort-by=.metadata.creationTimestamp

# Exec into containers
kubectl exec -it -n rpc deployment/dev-grpc-server -- /bin/sh
```

## Configuration

### Development Environment (overlays/development)
- Reduced resource limits for local development
- Single replica for most services
- `dev-` prefix for all resources

### Customization
To modify configurations:
1. Edit base YAML files in `base/` directory
2. Add patches in `overlays/development/deployment-patches.yaml`
3. Update image tags in `kustomization.yaml`

## Cleanup

Remove all resources:
```bash
kubectl delete namespace rpc
```

Or use kustomize:
```bash
kustomize build k8s/overlays/development | kubectl delete -f -
```

## Troubleshooting

### Common Issues

1. **Images not found**: Ensure Docker registry is accessible and images are pushed
2. **Migration fails**: Check PostgreSQL is ready and credentials are correct
3. **Services not ready**: Check resource limits and probe configurations
4. **Network issues**: Verify service DNS names and port configurations

### Health Checks
- PostgreSQL: `pg_isready -U rpc_user -d rpc_dev`
- Valkey: `valkey-cli ping`
- gRPC Server: TCP socket check on port 50051
- FastAPI Client: HTTP GET `/health` on port 8000

## Production Considerations

For production deployment:
1. Create a production overlay in `overlays/production/`
2. Use proper resource limits and requests
3. Configure persistent volume storage classes
4. Set up monitoring and logging
5. Use secrets management (e.g., Sealed Secrets, External Secrets)
6. Configure ingress controllers instead of LoadBalancer
7. Enable TLS/SSL certificates
8. Set up backup strategies for PostgreSQL