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
  url  = "http://grafana.observability.svc.cluster.local/"
  auth = "glsa_y96MIwbxRE6hWt5ynFBGZfaHFUyI7GlH_a0aa8967"
}

provider "jsonnet" {
  jsonnet_path = "${path.module}/jsonnet/vendor"
}
