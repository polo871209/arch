local g = import 'github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet';

{
  // Core Istio queries for Golden Signals monitoring
  queries:: {
    // Version hash filter - handles All (.*) and specific versions using regex operator
    versionFilter(version_hash='')::
      if version_hash == '' || version_hash == '$__all' || version_hash == 'All'
      then ',rollouts_pod_template_hash=~".*"'
      else ',rollouts_pod_template_hash=~"' + version_hash + '"',

    // Traffic: Request Rate
    requestRate(service_name, interval='5m', version_hash='')::
      'sum(irate(istio_requests_total{destination_service_name="' + service_name + '"' + self.versionFilter(version_hash) + '}[' + interval + ']))',

    // Errors: gRPC Server Error Rate  
    grpcErrorRate(service_name, interval='5m', version_hash='')::
      local filter = self.versionFilter(version_hash);
      '(sum(irate(istio_requests_total{destination_service_name="' + service_name + '",grpc_response_status!="0",grpc_response_status!=""' + filter + '}[' + interval + '])) or vector(0)) / sum(irate(istio_requests_total{destination_service_name="' + service_name + '"' + filter + '}[' + interval + '])) * 100',

    // Errors: gRPC Server Error Rate by Status Code
    grpcErrorRateByCode(service_name, interval='5m', version_hash='')::
      local filter = self.versionFilter(version_hash);
      'sum by (grpc_response_status) (irate(istio_requests_total{destination_service_name="' + service_name + '",grpc_response_status!="0",grpc_response_status!=""' + filter + '}[' + interval + '])) / ignoring(grpc_response_status) group_left sum(irate(istio_requests_total{destination_service_name="' + service_name + '"' + filter + '}[' + interval + '])) * 100',

    // Errors: HTTP Client Error Rate
    httpClientErrorRate(service_name, interval='5m', version_hash='')::
      local filter = self.versionFilter(version_hash);
      '(sum(irate(istio_requests_total{destination_service_name="rpc-client",response_code=~"[45].."' + filter + '}[' + interval + '])) or vector(0)) / sum(irate(istio_requests_total{destination_service_name="rpc-client"' + filter + '}[' + interval + '])) * 100',

    // Errors: HTTP Client Error Rate by Response Code
    httpClientErrorRateByCode(service_name, interval='5m', version_hash='')::
      local filter = self.versionFilter(version_hash);
      'sum by (response_code) (irate(istio_requests_total{destination_service_name="rpc-client",response_code=~"[45].."' + filter + '}[' + interval + '])) / ignoring(response_code) group_left sum(irate(istio_requests_total{destination_service_name="rpc-client"' + filter + '}[' + interval + '])) * 100',

    // Latency: Percentiles
    latencyP99(service_name, interval='5m', version_hash='')::
      'histogram_quantile(0.99, sum(irate(istio_request_duration_milliseconds_bucket{destination_service_name="' + service_name + '"' + self.versionFilter(version_hash) + '}[' + interval + '])) by (le))',

    latencyP90(service_name, interval='5m', version_hash='')::
      'histogram_quantile(0.90, sum(irate(istio_request_duration_milliseconds_bucket{destination_service_name="' + service_name + '"' + self.versionFilter(version_hash) + '}[' + interval + '])) by (le))',

    latencyP50(service_name, interval='5m', version_hash='')::
      'histogram_quantile(0.50, sum(irate(istio_request_duration_milliseconds_bucket{destination_service_name="' + service_name + '"' + self.versionFilter(version_hash) + '}[' + interval + '])) by (le))',

    // Saturation: Resource Utilization (pod-level, no version filtering)
    cpuUtilization(service_name)::
      'avg(rate(container_cpu_usage_seconds_total{pod=~"' + service_name + '-.*"}[5m])) * 100',

    memoryUtilization(service_name)::
      'avg(container_memory_working_set_bytes{pod=~"' + service_name + '-.*"} / container_spec_memory_limit_bytes{pod=~"' + service_name + '-.*"}) * 100',

    // Rollout: Version distribution (shows all versions, no filtering)
    rolloutVersionsPercentage(service_name, interval='5m')::
      'sum by (rollouts_pod_template_hash) (rate(istio_requests_total{destination_service_name="' + service_name + '",rollouts_pod_template_hash!=""}[' + interval + '])) / on() group_left() sum(rate(istio_requests_total{destination_service_name="' + service_name + '",rollouts_pod_template_hash!=""}[' + interval + '])) * 100',

    // Template variable: Available version hashes
    availableVersions(service_name)::
      'label_values(istio_requests_total{destination_service_name="' + service_name + '",rollouts_pod_template_hash!=""}, rollouts_pod_template_hash)',
  },

  // Panel templates with shared configuration
  panels:: {
    local commonDatasource = { type: 'prometheus', uid: 'prometheus' },
    local baseGridPos(pos) = {
      h: std.get(pos, 'h', 6),
      w: std.get(pos, 'w', 6), 
      x: std.get(pos, 'x', 0),
      y: std.get(pos, 'y', 0),
    },

    // Helper for stat panels with thresholds
    statPanel(title, query, unit, thresholds, gridPos)::
      g.panel.stat.new(title)
      + g.panel.stat.datasource.withType(commonDatasource.type)
      + g.panel.stat.datasource.withUid(commonDatasource.uid)
      + g.panel.stat.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', query)
          + g.query.prometheus.withLegendFormat(title),
      ])
      + g.panel.stat.standardOptions.withUnit(unit)
      + (if std.objectHas(thresholds, 'min') then g.panel.stat.standardOptions.withMin(thresholds.min) else {})
      + (if std.objectHas(thresholds, 'max') then g.panel.stat.standardOptions.withMax(thresholds.max) else {})
      + g.panel.stat.standardOptions.thresholds.withSteps(thresholds.steps)
      + g.panel.stat.gridPos.withH(baseGridPos(gridPos).h)
      + g.panel.stat.gridPos.withW(baseGridPos(gridPos).w)
      + g.panel.stat.gridPos.withX(baseGridPos(gridPos).x)
      + g.panel.stat.gridPos.withY(baseGridPos(gridPos).y),

    // Golden Signals panels
    requestRatePanel(service_name, title='Request Rate', gridPos={}, version_hash='')::
      self.statPanel(title, $.queries.requestRate(service_name, '5m', version_hash), 'reqps', {
        steps: [
          { color: 'green', value: 0 },
        ]
      }, gridPos),

    grpcErrorRatePanel(service_name, title='gRPC Server Errors', gridPos={}, version_hash='')::
      self.statPanel(title, $.queries.grpcErrorRate(service_name, '5m', version_hash), 'percent', {
        min: 0,
        steps: [
          { color: 'green', value: 0 },
          { color: 'yellow', value: 1 },
          { color: 'red', value: 5 },
        ]
      }, gridPos),

    httpClientErrorRatePanel(service_name, title='HTTP Client Errors', gridPos={}, version_hash='')::
      self.statPanel(title, $.queries.httpClientErrorRate(service_name, '5m', version_hash), 'percent', {
        min: 0,
        steps: [
          { color: 'green', value: 0 },
          { color: 'yellow', value: 2 },
          { color: 'red', value: 10 },
        ]
      }, gridPos),

    latencyP99Panel(service_name, title='P99 Response Time', gridPos={}, version_hash='')::
      self.statPanel(title, $.queries.latencyP99(service_name, '5m', version_hash), 'ms', {
        steps: [
          { color: 'green', value: 0 },
          { color: 'yellow', value: 100 },
          { color: 'red', value: 500 },
        ]
      }, gridPos),

    // Saturation panels  
    cpuUtilizationPanel(service_name, title='CPU Utilization', gridPos={})::
      self.statPanel(title, $.queries.cpuUtilization(service_name), 'percent', {
        min: 0, max: 100,
        steps: [
          { color: 'green', value: 0 },
          { color: 'yellow', value: 70 },
          { color: 'red', value: 90 },
        ]
      }, gridPos),

    memoryUtilizationPanel(service_name, title='Memory Utilization', gridPos={})::
      self.statPanel(title, $.queries.memoryUtilization(service_name), 'percent', {
        min: 0, max: 100,
        steps: [
          { color: 'green', value: 0 },
          { color: 'yellow', value: 80 },
          { color: 'red', value: 95 },
        ]
      }, gridPos),

    // Time series panels
    latencyPanel(service_name, title='Latency Distribution', gridPos={}, version_hash='')::
      g.panel.timeSeries.new(title)
      + g.panel.timeSeries.datasource.withType(commonDatasource.type)
      + g.panel.timeSeries.datasource.withUid(commonDatasource.uid)
      + g.panel.timeSeries.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.latencyP50(service_name, '5m', version_hash))
          + g.query.prometheus.withLegendFormat('P50'),
          g.query.prometheus.new('prometheus', $.queries.latencyP90(service_name, '5m', version_hash))
          + g.query.prometheus.withLegendFormat('P90'),
          g.query.prometheus.new('prometheus', $.queries.latencyP99(service_name, '5m', version_hash))
          + g.query.prometheus.withLegendFormat('P99'),
      ])
      + g.panel.timeSeries.standardOptions.withUnit('ms')
      + g.panel.timeSeries.standardOptions.withMin(0)
      + g.panel.timeSeries.gridPos.withH(baseGridPos(gridPos).h)
      + g.panel.timeSeries.gridPos.withW(std.get(gridPos, 'w', 24))
      + g.panel.timeSeries.gridPos.withX(baseGridPos(gridPos).x)
      + g.panel.timeSeries.gridPos.withY(baseGridPos(gridPos).y),

    rolloutVersionsPanel(service_name, title='Rollout Versions Distribution', gridPos={})::
      g.panel.timeSeries.new(title)
      + g.panel.timeSeries.datasource.withType(commonDatasource.type)
      + g.panel.timeSeries.datasource.withUid(commonDatasource.uid)
      + g.panel.timeSeries.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.rolloutVersionsPercentage(service_name))
          + g.query.prometheus.withLegendFormat('{{rollouts_pod_template_hash}}'),
      ])
      + g.panel.timeSeries.standardOptions.withUnit('percent')
      + g.panel.timeSeries.standardOptions.withMin(0)
      + g.panel.timeSeries.standardOptions.withMax(100)
      + g.panel.timeSeries.gridPos.withH(baseGridPos(gridPos).h)
      + g.panel.timeSeries.gridPos.withW(baseGridPos(gridPos).w)
      + g.panel.timeSeries.gridPos.withX(baseGridPos(gridPos).x)
      + g.panel.timeSeries.gridPos.withY(baseGridPos(gridPos).y),
  },

  // Row generators
  rows:: {
    local rowPanel(title, y) =
      g.panel.row.new(title)
      + g.panel.row.gridPos.withH(1)
      + g.panel.row.gridPos.withW(24) 
      + g.panel.row.gridPos.withX(0)
      + g.panel.row.gridPos.withY(y),

    goldenSignalsRow(y=0):: rowPanel('Golden Signals', y),
    saturationRow(y=7):: rowPanel('Saturation', y),
    trendsRow(y=14):: rowPanel('Trends', y),
  },
}