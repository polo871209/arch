# gRPC Service

Modern gRPC service with **Go 1.24 server** and **Python 3.13 FastAPI client** using best practice project structure.

## 🚀 Quick Start

### Local Development

```bash
just install    # Install deps & generate proto
just start      # Start both services
```

### Kubernetes Deployment

```bash
cd k8s
./deploy.sh deploy  # Deploy to OrbStack Kubernetes
```

**Service URLs:**

- **Local**: `localhost:50051` (gRPC), `http://localhost:8000` (REST)
- **Kubernetes**: `http://<EXTERNAL-IP>:8000` (get IP: `kubectl get svc fastapi-client-service`)
- **API Docs**: `/docs` endpoint

### Common Commands

```bash
# Local Development
just start      # Start both services (in-memory)
just start-db   # Start both services (PostgreSQL)
just stop       # Stop all services
just logs       # View logs
just test       # Test API endpoints

# Kubernetes
cd k8s && ./deploy.sh check   # Check prerequisites
cd k8s && ./deploy.sh deploy  # Full deployment
cd k8s && ./deploy.sh status  # Show status
cd k8s && ./deploy.sh clean   # Clean up
```

