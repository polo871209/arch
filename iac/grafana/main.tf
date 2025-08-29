resource "grafana_dashboard" "dashboards" {
  for_each    = data.jsonnet_file.dashboards
  config_json = each.value.rendered
}
