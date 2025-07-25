#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REGISTRY="localhost:5000"
NAMESPACE="rpc"
KUSTOMIZE_DIR="k8s/overlays/development"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to wait for deployment to be ready
wait_for_deployment() {
    local deployment=$1
    local namespace=${2:-$NAMESPACE}
    
    print_status "Waiting for deployment $deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/$deployment -n $namespace
    print_success "Deployment $deployment is ready"
}

# Function to wait for job to complete
wait_for_job() {
    local job=$1
    local namespace=${2:-$NAMESPACE}
    
    print_status "Waiting for job $job to complete..."
    kubectl wait --for=condition=complete --timeout=300s job/$job -n $namespace
    print_success "Job $job completed successfully"
}

# Function to check if resource exists
resource_exists() {
    local resource_type=$1
    local resource_name=$2
    local namespace=${3:-$NAMESPACE}
    
    kubectl get $resource_type $resource_name -n $namespace &>/dev/null
}

# Function to delete job if it exists
cleanup_job() {
    local job=$1
    local namespace=${2:-$NAMESPACE}
    
    if resource_exists job $job $namespace; then
        print_status "Cleaning up existing job $job..."
        kubectl delete job $job -n $namespace --ignore-not-found=true
        sleep 2
    fi
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed or not in PATH"
    exit 1
fi

# Check if kustomize is available
if ! command -v kustomize &> /dev/null; then
    print_error "kustomize is not installed or not in PATH"
    print_status "Installing kustomize..."
    curl -s "https://raw.githubusercontent.com/kubectl-kustomize/kustomize/master/hack/install_kustomize.sh" | bash
    sudo mv kustomize /usr/local/bin/
fi

# Check if we're in the right directory
if [[ ! -d "$KUSTOMIZE_DIR" ]]; then
    print_error "Kustomize directory $KUSTOMIZE_DIR not found. Please run from the project root."
    exit 1
fi

print_status "Starting RPC application deployment with Kustomize..."

# Phase 1: Build and push Docker images
print_status "Phase 1: Building and pushing Docker images..."

print_status "Building gRPC server image..."
docker build -t $REGISTRY/rpc-server:latest .
docker push $REGISTRY/rpc-server:latest

print_status "Building FastAPI client image..."
docker build -t $REGISTRY/rpc-client:latest -f client/Dockerfile client/
docker push $REGISTRY/rpc-client:latest

print_status "Building migration image..."
docker build -t $REGISTRY/rpc-migration:latest -f Dockerfile.migration .
docker push $REGISTRY/rpc-migration:latest

print_success "All Docker images built and pushed successfully"

# Phase 2: Deploy infrastructure (PostgreSQL, Valkey)
print_status "Phase 2: Deploying infrastructure..."

# Create or ensure namespace exists
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Apply base infrastructure resources first
print_status "Deploying PostgreSQL and Valkey..."
kustomize build $KUSTOMIZE_DIR | kubectl apply -f - --selector="component in (database,cache)"

# Wait for PostgreSQL to be ready
wait_for_deployment "dev-postgres"

# Wait for Valkey to be ready  
wait_for_deployment "dev-valkey"

print_success "Infrastructure deployment completed"

# Phase 3: Run database migration
print_status "Phase 3: Running database migration..."

# Clean up any existing migration job
cleanup_job "dev-db-migration"

# Deploy migration job
print_status "Deploying migration job..."
kustomize build $KUSTOMIZE_DIR | kubectl apply -f - --selector="app=migration"

# Wait for migration to complete
wait_for_job "dev-db-migration"

print_success "Database migration completed"

# Phase 4: Deploy applications (gRPC server, FastAPI client)
print_status "Phase 4: Deploying applications..."

print_status "Deploying gRPC server..."
kustomize build $KUSTOMIZE_DIR | kubectl apply -f - --selector="component=backend"

# Wait for gRPC server to be ready
wait_for_deployment "dev-grpc-server"

print_status "Deploying FastAPI client..."
kustomize build $KUSTOMIZE_DIR | kubectl apply -f - --selector="component=frontend"

# Wait for FastAPI client to be ready
wait_for_deployment "dev-fastapi-client"

print_success "Application deployment completed"

# Phase 5: Verification and status
print_status "Phase 5: Deployment verification..."

print_status "Checking deployment status:"
kubectl get all -n $NAMESPACE

print_status "Checking service endpoints:"
kubectl get services -n $NAMESPACE

# Get LoadBalancer endpoint for FastAPI client
FASTAPI_SERVICE=$(kubectl get service dev-fastapi-client-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [[ -n "$FASTAPI_SERVICE" ]]; then
    print_success "FastAPI client available at: http://$FASTAPI_SERVICE:8000"
    print_status "Health check: http://$FASTAPI_SERVICE:8000/health"
    print_status "API docs: http://$FASTAPI_SERVICE:8000/docs"
else
    # For local development, show port-forward command
    print_warning "LoadBalancer IP not yet assigned. For local access, use:"
    echo "kubectl port-forward -n $NAMESPACE service/dev-fastapi-client-service 8000:8000"
fi

print_success "🎉 RPC application deployment completed successfully!"

print_status "Useful commands:"
echo "  View logs: kubectl logs -n $NAMESPACE -l app=<app-name> --tail=100 -f"
echo "  Port forward: kubectl port-forward -n $NAMESPACE service/dev-fastapi-client-service 8000:8000"
echo "  Scale deployment: kubectl scale -n $NAMESPACE deployment/dev-grpc-server --replicas=3"
echo "  Update deployment: kubectl rollout restart -n $NAMESPACE deployment/dev-grpc-server"
echo ""
print_status "To clean up: kubectl delete namespace $NAMESPACE"