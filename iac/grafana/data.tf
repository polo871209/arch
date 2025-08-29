data "jsonnet_file" "dashboards" {
  for_each = fileset(path.module, "jsonnet/dashboards/**/*.jsonnet")
  source   = each.value
}
