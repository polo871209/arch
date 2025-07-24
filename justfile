# gRPC Service - Justfile

# Default: Start both servers
default: start

# Show available commands
help:
    @echo "🎯 Available commands:"
    @echo ""
    @echo "📦 Setup & Dependencies:"
    @echo "  install      Install all dependencies"
    @echo ""
    @echo "🔧 Code Generation:"
    @echo "  proto        Generate protobuf files"
    @echo "  sqlc         Generate SQL code with sqlc"
    @echo ""
    @echo "🗄️ Database:"
    @echo "  db-up        Run database migrations (Docker PostgreSQL)"
    @echo "  db-down      Rollback database migrations (Docker PostgreSQL)"
    @echo "  db-reset     Reset database (down all, then up)"
    @echo "  db-status    Show migration status"
    @echo "  db-migrate   Create new migration file"
    @echo ""
    @echo "🚀 Services:"
    @echo "  start        Start both servers with PostgreSQL"
    @echo "  server       Start Go server only (foreground)"
    @echo "  client       Start Python client only (foreground)"
    @echo "  stop         Stop all services"
    @echo ""
    @echo "🧪 Testing & Quality:"
    @echo "  test         Test API endpoints"
    @echo "  build        Build Go server"
    @echo "  fmt          Format code"
    @echo "  lint         Lint code"
    @echo ""
    @echo "📊 Monitoring:"
    @echo "  status       Check service status"
    @echo "  logs         Show recent logs"
    @echo "  docs         Open API documentation"
    @echo ""
    @echo "🧹 Cleanup:"
    @echo "  clean        Clean up logs and build artifacts"

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

# Generate SQL code using sqlc
sqlc:
    @echo "🗄️ Generating SQL code..."
    sqlc generate
    @echo "✅ SQL code generated"

# Database migration up
db-up:
    @echo "⬆️ Running database migrations..."
    goose -dir internal/database/migrations postgres "postgres://rpc_user:rpc_password@localhost:5433/rpc_dev?sslmode=disable" up
    @echo "✅ Migrations applied"

# Database migration down
db-down:
    @echo "⬇️ Rolling back database migrations..."
    goose -dir internal/database/migrations postgres "postgres://rpc_user:rpc_password@localhost:5433/rpc_dev?sslmode=disable" down
    @echo "✅ Migrations rolled back"

# Create a new migration
db-migrate name:
    @echo "📝 Creating migration: {{name}}"
    goose -dir internal/database/migrations create {{name}} sql
    @echo "✅ Migration created"

# Reset database (down all, then up)
db-reset:
    @echo "🔄 Resetting database..."
    goose -dir internal/database/migrations postgres "postgres://rpc_user:rpc_password@localhost:5433/rpc_dev?sslmode=disable" reset
    goose -dir internal/database/migrations postgres "postgres://rpc_user:rpc_password@localhost:5433/rpc_dev?sslmode=disable" up
    @echo "✅ Database reset complete"

# Show migration status
db-status:
    @echo "📊 Database migration status:"
    goose -dir internal/database/migrations postgres "postgres://rpc_user:rpc_password@localhost:5433/rpc_dev?sslmode=disable" status

# Build Go server
build:
    @echo "🏗️ Building server..."
    go build -o bin/server cmd/server/main.go
    @echo "✅ Server built"

# Start both servers with PostgreSQL
start:
    @echo "🚀 Starting services with PostgreSQL..."
    @echo "   Starting PostgreSQL container..."
    docker-compose up -d postgres
    @echo "   Waiting for PostgreSQL to be ready..."
    ./scripts/wait_for_postgres.sh
    @echo "   Running database migrations..."
    @just db-up
    @mkdir -p logs
    nohup sh -c 'go run cmd/server/main.go' > logs/server.log 2>&1 & echo $! > logs/server.pid
    @sleep 2
    nohup sh -c 'cd client && uv run uvicorn app.main:app --host 0.0.0.0 --port 8000' > logs/client.log 2>&1 & echo $! > logs/client.pid
    @echo "✅ Services started with PostgreSQL!"
    @echo "   PostgreSQL: localhost:5433"
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
    @echo "   Stopping PostgreSQL container..."
    docker-compose down postgres
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