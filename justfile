default: start

timestamp := `date +%Y%m%d%H%M%S`
registry := "localhost:5000"

image-server := "rpc-server"
image-migration := "rpc-migration"
image-client := "rpc-client"

overlay := "k8s/overlays/development"

build-push image dockerfile context:
    set -euo pipefail
    docker build -t {{registry}}/{{image}}:{{timestamp}} -f {{dockerfile}} {{context}}
    docker push {{registry}}/{{image}}:{{timestamp}}
    cd {{overlay}}/app && kustomize edit set image {{image}}={{registry}}/{{image}}:{{timestamp}}

[parallel]
build:
    just build-push {{image-server}} Dockerfile .
    just build-push {{image-migration}} Dockerfile.migration .
    just build-push {{image-client}} ./client/Dockerfile ./client

start:
    just build
    kustomize build {{overlay}}/app | kubectl apply -f -

infra:
    kustomize build {{overlay}}/infra | kubectl apply -f -

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

