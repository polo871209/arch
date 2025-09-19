default: build-deploy

timestamp := `date +%Y%m%d%H%M%S`
registry := "localhost:5000"
dev_overlay := "kustomize/app/overlays/dev"
migration_overlay := "kustomize/app/overlays/dev/migration"

build-tag image dockerfile context overlay:
    @docker build -t {{registry}}/{{image}}:{{timestamp}} -f {{dockerfile}} {{context}} && \
        cd {{overlay}} && \
        kustomize edit set image {{image}}={{registry}}/{{image}}:{{timestamp}}

build-deploy: (build-tag "rpc-server" "rpc-server/Dockerfile" "rpc-server" dev_overlay) (build-tag "rpc-client" "rpc-client/Dockerfile" "rpc-client" dev_overlay)
    @kustomize build {{dev_overlay}}/ | kubectl apply -f -

migration: (build-tag "rpc-migration" "rpc-server/Dockerfile.migration" "rpc-server" migration_overlay)
    @cd {{migration_overlay}} && kustomize edit set namesuffix {{timestamp}}
    @kustomize build {{migration_overlay}} | kubectl apply -f -

infra:
    @kustomize build ./kustomize/infra | kubectl apply -f -
    @helm upgrade --install cert-manager oci://quay.io/jetstack/charts/cert-manager \
        --version v1.18.2 \
        -n cert-manager \
        --set crds.enabled=true \
        --wait
    @helm repo add istio https://istio-release.storage.googleapis.com/charts || true
    @helm repo update istio
    @helm upgrade --install istio-base istio/base \
        -n istio-system \
        --set defaultRevision=default
    @helm upgrade --install istiod istio/istiod \
        -n istio-system \
        --wait \
        --values ./kustomize/infra/values/istiod.yaml
    @helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts || true
    @helm repo update open-telemetry
    @helm upgrade --install opentelemetry-operator open-telemetry/opentelemetry-operator \
        -n observability \
        --values ./kustomize/infra/values/opentelemetry-operator.yaml \
        --wait
    @helm upgrade --install kube-prometheus-stack \
        -n observability \
        --version 77.13.0 \
        oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack \
        --values ./kustomize/infra/values/kube-prometheus-stack.yaml \
        --wait
    @helm repo add elastic https://helm.elastic.co || true
    @helm repo update elastic
    @helm upgrade --install eck-operator elastic/eck-operator \
        --version 3.1.0 \
        -n elastic-system \
        --values ./kustomize/infra/values/eck-operator.yaml \
        --wait
    @helm upgrade --install eck-stack elastic/eck-stack \
        --version 0.16.0 \
        -n observability \
        --values ./kustomize/infra/values/eck-stack.yaml \
        --values ./kustomize/infra/values/eck-elasticsearch.yaml \
        --values ./kustomize/infra/values/eck-kibana.yaml \
        --wait
    @helm repo add fluent https://fluent.github.io/helm-charts || true
    @helm repo update fluent
    @helm upgrade --install fluent-bit fluent/fluent-bit \
        --version 0.50.0 \
        -n observability \
        --values ./kustomize/infra/values/fluent-bit.yaml \
        --wait


proto:
    @protoc -Iproto --go_out=rpc-server/pkg/pb --go_opt=paths=source_relative --go-grpc_out=rpc-server/pkg/pb --go-grpc_opt=paths=source_relative ./proto/user.proto
    @cd rpc-client/proto && uv run python -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. --pyi_out=. ../../proto/user.proto
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
