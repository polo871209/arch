local istio = import '../lib/istio.libsonnet';
local g = import 'github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet';

local service_name = 'rpc-server';

local dashboard = g.dashboard.new('Service Overview')
+ g.dashboard.withDescription('Complete service monitoring: Golden Signals and resource utilization')
+ g.dashboard.withTags(['sre', 'istio', 'monitoring'])
+ g.dashboard.time.withFrom('now-1h')
+ g.dashboard.time.withTo('now')
+ g.dashboard.withRefresh('30s')
+ g.dashboard.withTimezone('browser')
+ g.dashboard.withPanels([
  // Golden Signals Section
  istio.rows.goldenSignalsRow(0),
  
  istio.panels.requestRatePanel(service_name, 'Request Rate', {x: 0, y: 1, w: 6, h: 6}),
  istio.panels.grpcErrorRatePanel(service_name, 'gRPC Server Errors', {x: 6, y: 1, w: 6, h: 6}),  
  istio.panels.httpClientErrorRatePanel(service_name, 'HTTP Client Errors', {x: 12, y: 1, w: 6, h: 6}),
  
  // P99 Latency Panel
  g.panel.stat.new('P99 Response Time')
  + g.panel.stat.datasource.withType('prometheus')
  + g.panel.stat.datasource.withUid('prometheus')
  + g.panel.stat.queryOptions.withTargets([
      g.query.prometheus.new('prometheus', istio.queries.latencyP99(service_name))
      + g.query.prometheus.withFormat('time_series')
      + g.query.prometheus.withLegendFormat('P99'),
  ])
  + g.panel.stat.standardOptions.withUnit('ms')
  + g.panel.stat.standardOptions.thresholds.withSteps([
      g.panel.stat.standardOptions.threshold.step.withColor('green') + g.panel.stat.standardOptions.threshold.step.withValue(0),
      g.panel.stat.standardOptions.threshold.step.withColor('yellow') + g.panel.stat.standardOptions.threshold.step.withValue(100),
      g.panel.stat.standardOptions.threshold.step.withColor('red') + g.panel.stat.standardOptions.threshold.step.withValue(500),
  ])
  + g.panel.stat.gridPos.withH(6)
  + g.panel.stat.gridPos.withW(6)
  + g.panel.stat.gridPos.withX(18)
  + g.panel.stat.gridPos.withY(1),

  // Saturation Section  
  istio.rows.saturationRow(7),
  
  istio.panels.cpuUtilizationPanel(service_name, 'CPU Utilization', {x: 0, y: 8, w: 6, h: 6}),
  istio.panels.memoryUtilizationPanel(service_name, 'Memory Utilization', {x: 6, y: 8, w: 6, h: 6}),
  istio.panels.latencyPanel(service_name, 'Latency Distribution', {x: 12, y: 8, w: 12, h: 6}),

  // Trends Section
  istio.rows.trendsRow(14),
  
  // Request Rate Over Time
  g.panel.timeSeries.new('Request Rate')
  + g.panel.timeSeries.datasource.withType('prometheus')
  + g.panel.timeSeries.datasource.withUid('prometheus')
  + g.panel.timeSeries.queryOptions.withTargets([
      g.query.prometheus.new('prometheus', istio.queries.requestRate(service_name))
      + g.query.prometheus.withFormat('time_series')
      + g.query.prometheus.withLegendFormat('req/sec'),
  ])
  + g.panel.timeSeries.standardOptions.withUnit('reqps')
  + g.panel.timeSeries.gridPos.withH(8)
  + g.panel.timeSeries.gridPos.withW(12)
  + g.panel.timeSeries.gridPos.withX(0)
  + g.panel.timeSeries.gridPos.withY(15),

  // Error Rates Over Time
  g.panel.timeSeries.new('Error Rates')
  + g.panel.timeSeries.datasource.withType('prometheus')
  + g.panel.timeSeries.datasource.withUid('prometheus')
  + g.panel.timeSeries.queryOptions.withTargets([
      g.query.prometheus.new('prometheus', istio.queries.grpcErrorRateByCode(service_name))
      + g.query.prometheus.withFormat('time_series')
      + g.query.prometheus.withLegendFormat('gRPC {{grpc_response_status}}'),
      g.query.prometheus.new('prometheus', istio.queries.httpClientErrorRateByCode(service_name))
      + g.query.prometheus.withFormat('time_series')
      + g.query.prometheus.withLegendFormat('HTTP {{response_code}}'),
  ])
  + g.panel.timeSeries.standardOptions.withUnit('percent')
  + g.panel.timeSeries.standardOptions.withMin(0)
  + g.panel.timeSeries.gridPos.withH(8)
  + g.panel.timeSeries.gridPos.withW(12)
  + g.panel.timeSeries.gridPos.withX(12)
  + g.panel.timeSeries.gridPos.withY(15),
]);

dashboard

