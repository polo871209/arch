terraform {
  required_providers {
    grafana = {
      source  = "grafana/grafana"
      version = ">=4.5.3"
    }
    jsonnet = {
      source  = "alxrem/jsonnet"
      version = ">=2.5.0"
    }
  }
}

provider "grafana" {
  url  = "http://kube-prometheus-stack-grafana.observability.svc.cluster.local"
  auth = "glsa_jKQgmPRU96KtnoZ2tSqsYolALtM7nHwr_644423d5"
}

provider "jsonnet" {
  jsonnet_path = "${path.module}/jsonnet/vendor"
}
