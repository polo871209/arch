default: local-build

timestamp := `date +%Y%m%d%H%M%S`
registry := "localhost:5000"

image-server := "rpc-server"
image-migration := "rpc-migration"
image-client := "rpc-client"

# For local development only
local-build-deploy image dockerfile context:
    @docker build -t {{registry}}/{{image}}:{{timestamp}} -f {{dockerfile}} {{context}}

[parallel]
local-build:
    @just local-build-deploy {{image-server}} Dockerfile .
    @just local-build-deploy {{image-client}} ./client/Dockerfile ./client
    @just local-build-deploy {{image-migration}} Dockerfile.migration .

[working-directory: 'argos/bootstrap']
argos-bootstrap:
    @kustomize build . | kubectl apply -f -

[working-directory: 'argos']
argos:
    @kustomize build . | kubectl apply -f -

proto:
    @protoc -Iproto --go_out=pkg/pb --go_opt=paths=source_relative --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative ./proto/user.proto
    @cd client/proto && uv run python -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. --pyi_out=. ../../proto/user.proto
    @echo "Please manually fix the import of python after proto generation."

[working-directory: 'iac/kibana']
kibana: 
    @uv run main.py

[working-directory: 'iac/grafana']
grafana-init:
    @jb install
    @tofu init

[working-directory: 'iac/grafana']
grafana-update:
    @tofu apply -auto-approve



