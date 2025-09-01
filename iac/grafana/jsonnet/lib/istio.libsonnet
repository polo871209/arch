local g = import 'github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet';

{
  // Core Istio queries for Golden Signals monitoring
  queries:: {
    // Traffic - Request Rate
    requestRate(service_name, interval='5m')::
      'sum(irate(istio_requests_total{destination_service_name="' + service_name + '"}[' + interval + ']))',

    // Errors - gRPC Server Error Rate (combined)
    grpcErrorRate(service_name, interval='5m')::
      '(sum(irate(istio_requests_total{destination_service_name="' + service_name + '",grpc_response_status!="0",grpc_response_status!=""}[' + interval + '])) or vector(0)) / sum(irate(istio_requests_total{destination_service_name="' + service_name + '"}[' + interval + '])) * 100',

    // Errors - gRPC Server Error Rate (by status code)
    grpcErrorRateByCode(service_name, interval='5m')::
      'sum by (grpc_response_status) (irate(istio_requests_total{destination_service_name="' + service_name + '",grpc_response_status!="0",grpc_response_status!=""}[' + interval + '])) / ignoring(grpc_response_status) group_left sum(irate(istio_requests_total{destination_service_name="' + service_name + '"}[' + interval + '])) * 100',

    // Errors - HTTP Client Error Rate (for client-side monitoring)
    httpClientErrorRate(service_name, interval='5m')::
      '(sum(irate(istio_requests_total{destination_service_name="rpc-client",response_code=~"[45].."}[' + interval + '])) or vector(0)) / sum(irate(istio_requests_total{destination_service_name="rpc-client"}[' + interval + '])) * 100',

    // Errors - HTTP Client Error Rate (by response code)
    httpClientErrorRateByCode(service_name, interval='5m')::
      'sum by (response_code) (irate(istio_requests_total{destination_service_name="rpc-client",response_code=~"[45].."}[' + interval + '])) / ignoring(response_code) group_left sum(irate(istio_requests_total{destination_service_name="rpc-client"}[' + interval + '])) * 100',

    // Latency - Percentiles
    latencyP99(service_name, interval='5m')::
      'histogram_quantile(0.99, sum(irate(istio_request_duration_milliseconds_bucket{destination_service_name="' + service_name + '"}[' + interval + '])) by (le))',

    latencyP90(service_name, interval='5m')::
      'histogram_quantile(0.90, sum(irate(istio_request_duration_milliseconds_bucket{destination_service_name="' + service_name + '"}[' + interval + '])) by (le))',

    latencyP50(service_name, interval='5m')::
      'histogram_quantile(0.50, sum(irate(istio_request_duration_milliseconds_bucket{destination_service_name="' + service_name + '"}[' + interval + '])) by (le))',

    // Saturation - Resource Utilization
    cpuUtilization(service_name)::
      'avg(rate(container_cpu_usage_seconds_total{pod=~"' + service_name + '-.*"}[5m])) * 100',

    memoryUtilization(service_name)::
      'avg(container_memory_working_set_bytes{pod=~"' + service_name + '-.*"} / container_spec_memory_limit_bytes{pod=~"' + service_name + '-.*"}) * 100',
  },

  // Panel templates optimized for modern Grafana
  panels:: {
    // Traffic Panel
    requestRatePanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else 'Request Rate';
      g.panel.stat.new(panelTitle)
      + g.panel.stat.datasource.withType('prometheus')
      + g.panel.stat.datasource.withUid('prometheus')
      + g.panel.stat.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.requestRate(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('req/sec'),
      ])
      + g.panel.stat.standardOptions.withUnit('reqps')
      + g.panel.stat.gridPos.withH(std.get(gridPos, 'h', 6))
      + g.panel.stat.gridPos.withW(std.get(gridPos, 'w', 6))
      + g.panel.stat.gridPos.withX(std.get(gridPos, 'x', 0))
      + g.panel.stat.gridPos.withY(std.get(gridPos, 'y', 0)),

    // Error Rate Panel (gRPC Server)
    grpcErrorRatePanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else 'gRPC Server Errors';
      g.panel.stat.new(panelTitle)
      + g.panel.stat.datasource.withType('prometheus')
      + g.panel.stat.datasource.withUid('prometheus')
      + g.panel.stat.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.grpcErrorRate(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('Error Rate'),
      ])
      + g.panel.stat.standardOptions.withUnit('percent')
      + g.panel.stat.standardOptions.withMin(0)
      + g.panel.stat.standardOptions.thresholds.withSteps([
          g.panel.stat.standardOptions.threshold.step.withColor('green') + g.panel.stat.standardOptions.threshold.step.withValue(0),
          g.panel.stat.standardOptions.threshold.step.withColor('yellow') + g.panel.stat.standardOptions.threshold.step.withValue(1),
          g.panel.stat.standardOptions.threshold.step.withColor('red') + g.panel.stat.standardOptions.threshold.step.withValue(5),
      ])
      + g.panel.stat.gridPos.withH(std.get(gridPos, 'h', 6))
      + g.panel.stat.gridPos.withW(std.get(gridPos, 'w', 6))
      + g.panel.stat.gridPos.withX(std.get(gridPos, 'x', 0))
      + g.panel.stat.gridPos.withY(std.get(gridPos, 'y', 0)),

    // HTTP Client Error Rate Panel
    httpClientErrorRatePanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else 'HTTP Client Errors';
      g.panel.stat.new(panelTitle)
      + g.panel.stat.datasource.withType('prometheus')
      + g.panel.stat.datasource.withUid('prometheus')
      + g.panel.stat.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.httpClientErrorRate(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('Error Rate'),
      ])
      + g.panel.stat.standardOptions.withUnit('percent')
      + g.panel.stat.standardOptions.withMin(0)
      + g.panel.stat.standardOptions.thresholds.withSteps([
          g.panel.stat.standardOptions.threshold.step.withColor('green') + g.panel.stat.standardOptions.threshold.step.withValue(0),
          g.panel.stat.standardOptions.threshold.step.withColor('yellow') + g.panel.stat.standardOptions.threshold.step.withValue(2),
          g.panel.stat.standardOptions.threshold.step.withColor('red') + g.panel.stat.standardOptions.threshold.step.withValue(10),
      ])
      + g.panel.stat.gridPos.withH(std.get(gridPos, 'h', 6))
      + g.panel.stat.gridPos.withW(std.get(gridPos, 'w', 6))
      + g.panel.stat.gridPos.withX(std.get(gridPos, 'x', 0))
      + g.panel.stat.gridPos.withY(std.get(gridPos, 'y', 0)),

    // Latency Distribution Panel
    latencyPanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else 'Latency Distribution';
      g.panel.timeSeries.new(panelTitle)
      + g.panel.timeSeries.datasource.withType('prometheus')
      + g.panel.timeSeries.datasource.withUid('prometheus')
      + g.panel.timeSeries.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.latencyP50(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('P50'),
          g.query.prometheus.new('prometheus', $.queries.latencyP90(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('P90'),
          g.query.prometheus.new('prometheus', $.queries.latencyP99(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('P99'),
      ])
      + g.panel.timeSeries.standardOptions.withUnit('ms')
      + g.panel.timeSeries.standardOptions.withMin(0)
      + g.panel.timeSeries.gridPos.withH(std.get(gridPos, 'h', 6))
      + g.panel.timeSeries.gridPos.withW(std.get(gridPos, 'w', 12))
      + g.panel.timeSeries.gridPos.withX(std.get(gridPos, 'x', 0))
      + g.panel.timeSeries.gridPos.withY(std.get(gridPos, 'y', 0)),

    // CPU Utilization Panel
    cpuUtilizationPanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else 'CPU Utilization';
      g.panel.stat.new(panelTitle)
      + g.panel.stat.datasource.withType('prometheus')
      + g.panel.stat.datasource.withUid('prometheus')
      + g.panel.stat.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.cpuUtilization(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('CPU Usage'),
      ])
      + g.panel.stat.standardOptions.withUnit('percent')
      + g.panel.stat.standardOptions.withMin(0)
      + g.panel.stat.standardOptions.withMax(100)
      + g.panel.stat.standardOptions.thresholds.withSteps([
          g.panel.stat.standardOptions.threshold.step.withColor('green') + g.panel.stat.standardOptions.threshold.step.withValue(0),
          g.panel.stat.standardOptions.threshold.step.withColor('yellow') + g.panel.stat.standardOptions.threshold.step.withValue(70),
          g.panel.stat.standardOptions.threshold.step.withColor('red') + g.panel.stat.standardOptions.threshold.step.withValue(90),
      ])
      + g.panel.stat.gridPos.withH(std.get(gridPos, 'h', 6))
      + g.panel.stat.gridPos.withW(std.get(gridPos, 'w', 6))
      + g.panel.stat.gridPos.withX(std.get(gridPos, 'x', 0))
      + g.panel.stat.gridPos.withY(std.get(gridPos, 'y', 0)),

    // Memory Utilization Panel
    memoryUtilizationPanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else 'Memory Utilization';
      g.panel.stat.new(panelTitle)
      + g.panel.stat.datasource.withType('prometheus')
      + g.panel.stat.datasource.withUid('prometheus')
      + g.panel.stat.queryOptions.withTargets([
          g.query.prometheus.new('prometheus', $.queries.memoryUtilization(service_name))
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('Memory Usage'),
      ])
      + g.panel.stat.standardOptions.withUnit('percent')
      + g.panel.stat.standardOptions.withMin(0)
      + g.panel.stat.standardOptions.withMax(100)
      + g.panel.stat.standardOptions.thresholds.withSteps([
          g.panel.stat.standardOptions.threshold.step.withColor('green') + g.panel.stat.standardOptions.threshold.step.withValue(0),
          g.panel.stat.standardOptions.threshold.step.withColor('yellow') + g.panel.stat.standardOptions.threshold.step.withValue(80),
          g.panel.stat.standardOptions.threshold.step.withColor('red') + g.panel.stat.standardOptions.threshold.step.withValue(95),
      ])
      + g.panel.stat.gridPos.withH(std.get(gridPos, 'h', 6))
      + g.panel.stat.gridPos.withW(std.get(gridPos, 'w', 6))
      + g.panel.stat.gridPos.withX(std.get(gridPos, 'x', 0))
      + g.panel.stat.gridPos.withY(std.get(gridPos, 'y', 0)),
  },

  // Row generators for structured dashboards
  rows:: {
    // Golden Signals Row
    goldenSignalsRow(y=0)::
      g.panel.row.new('Golden Signals')
      + g.panel.row.gridPos.withH(1)
      + g.panel.row.gridPos.withW(24)
      + g.panel.row.gridPos.withX(0)
      + g.panel.row.gridPos.withY(y),

    // Saturation Row  
    saturationRow(y=7)::
      g.panel.row.new('Saturation')
      + g.panel.row.gridPos.withH(1)
      + g.panel.row.gridPos.withW(24)
      + g.panel.row.gridPos.withX(0)
      + g.panel.row.gridPos.withY(y),

    // Trends Row
    trendsRow(y=14)::
      g.panel.row.new('Trends')
      + g.panel.row.gridPos.withH(1)
      + g.panel.row.gridPos.withW(24)
      + g.panel.row.gridPos.withX(0)
      + g.panel.row.gridPos.withY(y),
  },
}