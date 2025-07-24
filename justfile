# User Management gRPC Service - Justfile

# Default: Start both servers
default: start

# Install all dependencies
install:
    @echo "📦 Installing dependencies..."
    go mod tidy
    cd client && uv sync
    @echo "✅ Dependencies installed"

# Generate protobuf files
proto:
    @echo "🔧 Generating protobuf files..."
    ./scripts/generate_proto.sh
    @echo "✅ Protobuf files generated"

# Build Go server
build:
    @echo "🏗️ Building server..."
    go build -o bin/server cmd/server/main.go
    @echo "✅ Server built"

# Start both servers
start:
    @echo "🚀 Starting services..."
    @mkdir -p logs
    nohup sh -c 'go run cmd/server/main.go' > logs/server.log 2>&1 & echo $! > logs/server.pid
    @sleep 2
    nohup sh -c 'cd client && uv run uvicorn app.main:app --host 0.0.0.0 --port 8000' > logs/client.log 2>&1 & echo $! > logs/client.pid
    @echo "✅ Services started!"
    @echo "   gRPC Server: localhost:50051"
    @echo "   FastAPI Client: http://localhost:8000"

# Start Go server only (foreground)
server:
    go run cmd/server/main.go

# Start Python client only (foreground)
client:
    cd client && uv run uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload

# Stop all services
stop:
    @echo "🛑 Stopping services..."
    @if [ -f logs/server.pid ]; then kill $(cat logs/server.pid) 2>/dev/null || true; rm -f logs/server.pid; fi
    @if [ -f logs/client.pid ]; then kill $(cat logs/client.pid) 2>/dev/null || true; rm -f logs/client.pid; fi
    @echo "✅ Services stopped"

# Test API endpoints
test:
    @echo "🧪 Testing API..."
    @if ! curl -s http://localhost:8000/health > /dev/null 2>&1; then just start; sleep 3; fi
    curl -s http://localhost:8000/health | jq .
    curl -s http://localhost:8000/v1/users | jq .
    @echo "✅ Test completed"

# Show recent logs
logs:
    @echo "📋 Recent logs:"
    @if [ -f logs/server.log ]; then echo "=== Server ==="; tail -10 logs/server.log; fi
    @if [ -f logs/client.log ]; then echo "=== Client ==="; tail -10 logs/client.log; fi

# Check service status
status:
    @echo "📊 Service status:"
    @if [ -f logs/server.pid ] && kill -0 $(cat logs/server.pid) 2>/dev/null; then echo "✅ gRPC Server: Running (PID: $(cat logs/server.pid))"; else echo "❌ gRPC Server: Stopped"; fi
    @if [ -f logs/client.pid ] && kill -0 $(cat logs/client.pid) 2>/dev/null; then echo "✅ FastAPI Client: Running (PID: $(cat logs/client.pid))"; else echo "❌ FastAPI Client: Stopped"; fi

# Format code
fmt:
    @echo "🎨 Formatting code..."
    go fmt ./...
    cd client && uv run black app/
    cd client && uv run isort app/
    @echo "✅ Code formatted"

# Lint code
lint:
    @echo "🔍 Linting code..."
    go vet ./...
    cd client && uv run flake8 app/ --max-line-length=88
    @echo "✅ Code linted"

# Open API documentation
docs:
    @echo "📖 Opening API docs..."
    @if ! curl -s http://localhost:8000/health > /dev/null 2>&1; then just start; sleep 3; fi
    open http://localhost:8000/docs

# Clean up
clean:
    @just stop
    rm -rf logs/*.log logs/*.pid bin/ __pycache__ .pytest_cache