# gRPC Service

Modern gRPC service with **Go 1.24 server** and **Python 3.13 FastAPI client** using best practice project structure.

## 🚀 Quick Start

### One-Command Setup
```bash
just install    # Install deps & generate proto
just start      # Start both services
```

### Common Commands
```bash
just start      # Start both services (in-memory)
just start-db   # Start both services (PostgreSQL)
just stop       # Stop all services  
just logs       # View logs
just test       # Test API endpoints
just status     # Check service status
just help       # Show all available commands
```

**Service URLs:**
- gRPC Server: `localhost:50051`
- REST API: `http://localhost:8000`
- API Docs: `http://localhost:8000/docs`

## 📁 Project Structure

```
rpc/
├── cmd/server/           # Go server entry point
├── internal/             # Private Go packages
│   ├── server/          # gRPC server implementation
│   ├── models/          # Domain models
│   ├── repository/      # Data access layer
│   │   ├── memory/     # In-memory implementation
│   │   └── postgres/   # PostgreSQL implementation
│   ├── database/        # Database layer
│   │   ├── migrations/ # SQL migration files
│   │   ├── queries/    # SQL query files
│   │   └── generated/  # SQLC generated code
│   ├── config/          # Configuration
│   └── validation/      # Business validation
├── pkg/pb/              # Generated protobuf files
├── client/              # Python FastAPI client
│   ├── app/            # Application package
│   │   ├── api/        # API routes (versioned)
│   │   ├── core/       # Core utilities & config
│   │   ├── models/     # Pydantic models
│   │   ├── services/   # Business logic
│   │   └── grpc_client/ # gRPC client
│   └── proto/          # Generated Python proto
├── proto/              # Protobuf definitions
├── scripts/            # Build scripts
├── sqlc.yaml           # SQLC configuration
└── justfile            # Task runner commands
```

## ✨ Features

**Go 1.24 Server:**
- Clean architecture with dependency injection
- Structured logging with `slog`
- Configuration from environment variables
- Repository pattern for data access
- Comprehensive validation layer
- **Dual storage**: In-memory or PostgreSQL
- **Database migrations** with Goose
- **Type-safe SQL** with SQLC

**Python 3.13 Client:**
- FastAPI with modern async patterns
- Versioned API routes (`/api/v1/`)
- Pydantic models with validation
- Service layer architecture
- Dependency injection
- Type hints with `|` union syntax

## 📋 Prerequisites

- Go 1.24+, Python 3.13+, [Just](https://github.com/casey/just), `protoc`

## 🗄️ Database Setup

**In-Memory (Default):**
```bash
just start      # Uses in-memory storage
```

**PostgreSQL:**
```bash
# 1. Install PostgreSQL
# 2. Create database and user
createdb grpc_db
createuser -s user
psql -c "ALTER USER user PASSWORD 'password';"

# 3. Run migrations
just db-up

# 4. Start with PostgreSQL
just start-db
```

**Database Commands:**
```bash
just db-up       # Run migrations
just db-down     # Rollback migrations  
just db-reset    # Reset database
just db-status   # Check migration status
just sqlc        # Generate Go code from SQL
```

## 🔌 API Endpoints

**Example User Service:**
- `POST /v1/users` - Create user
- `GET /v1/users/{id}` - Get user  
- `PUT /v1/users/{id}` - Update user
- `DELETE /v1/users/{id}` - Delete user
- `GET /v1/users?page=1&limit=10` - List users
- `GET /health` - Health check

## 🛠 Development

```bash
just server     # Go server (foreground)
just client     # Python client (foreground + reload)
just proto      # Regenerate protobuf
just build      # Build Go binary
just fmt        # Format code
just lint       # Lint code
just docs       # Open API documentation
```

Run `just` to see all commands.

## 🏗 Architecture Benefits

- **Scalable**: Easy to add new services and features
- **Maintainable**: Clear separation of concerns  
- **Testable**: Proper dependency injection
- **Production-ready**: Configuration management, logging, validation
- **Team-friendly**: Clear boundaries and conventions
- **Flexible**: Generic structure supports any gRPC service