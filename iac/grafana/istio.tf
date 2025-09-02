# Grafana provided default dashboards for Istio
data "local_file" "istio_dashboards" {
  for_each = fileset(path.module, "json/istio/*.json")
  filename = "${path.module}/${each.value}"
}

resource "grafana_folder" "istio" {
  title = "istio"
}

resource "grafana_dashboard" "istio_dashboards" {
  for_each    = data.local_file.istio_dashboards
  config_json = each.value.content
  folder      = grafana_folder.istio.id
}
