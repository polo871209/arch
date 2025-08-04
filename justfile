default: start

timestamp := `date +%Y%m%d%H%M%S`
registry := "localhost:5000"

image-server := "rpc-server"
image-migration := "rpc-migration"
image-client := "rpc-client"

overlay := "kustomize/overlays/dev"

build-deploy image dockerfile context kustomize_subpath:
    @docker build -t {{registry}}/{{image}}:{{timestamp}} -f {{dockerfile}} {{context}}
    @cd {{overlay}}/{{kustomize_subpath}} && kustomize edit set image {{image}}={{registry}}/{{image}}:{{timestamp}}

[parallel]
build:
    @just build-deploy {{image-server}} Dockerfile . app
    @just build-deploy {{image-client}} ./client/Dockerfile ./client app

migration:
    @just build-deploy {{image-migration}} Dockerfile.migration . app/migration
    cd {{overlay}}/app/migration && kustomize edit set namesuffix -- -{{timestamp}} 
    @kustomize build {{overlay}}/app/migration | kubectl apply -f -

start:
    @just build
    @kustomize build {{overlay}}/app | kubectl apply -f -

[working-directory: 'infra']
infra:
    @kustomize build . | kubectl apply -f - | grep -v 'unchanged'

proto:
    @protoc -Iproto --go_out=pkg/pb --go_opt=paths=source_relative --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative ./proto/user.proto
    @cd client/proto && uv run python -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. --pyi_out=. ../../proto/user.proto
    @echo "Please manually fix the import of python after proto generation."
