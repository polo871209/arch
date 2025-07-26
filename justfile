default: start

timestamp := `date +%Y%m%d%H%M%S`
registry := "localhost:5000"

image-server := "rpc-server"
image-migration := "rpc-migration"
image-client := "rpc-client"

overlay := "k8s/overlays/development"

build-push image dockerfile context:
    @set -euo pipefail
    @docker build -t {{registry}}/{{image}}:{{timestamp}} -f {{dockerfile}} {{context}}
    @docker push {{registry}}/{{image}}:{{timestamp}}
    @cd {{overlay}}/app && kustomize edit set image {{image}}={{registry}}/{{image}}:{{timestamp}}

[parallel]
build:
    @just build-push {{image-server}} Dockerfile .
    @just build-push {{image-migration}} Dockerfile.migration .
    @just build-push {{image-client}} ./client/Dockerfile ./client

start:
    @just build
    @kustomize build {{overlay}}/app | kubectl apply -f -

infra:
    @kustomize build {{overlay}}/infra | kubectl apply -f -

install:
    @go mod tidy
    @cd client && uv sync

proto:
    @protoc -Iproto --go_out=pkg/pb --go_opt=paths=source_relative --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative ./proto/user.proto
    @cd client/proto && uv run python -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. --pyi_out=. ../../proto/user.proto
    @echo "Please manually fix the import of python after proto generation."

sqlc:
    @sqlc generate

