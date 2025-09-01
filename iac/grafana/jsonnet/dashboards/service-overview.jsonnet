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
+ {
  templating: {
    list: [{
      type: 'query',
      name: 'version_hash',
      label: 'Version Hash',
      query: istio.queries.availableVersions(service_name),
      datasource: { type: 'prometheus', uid: 'prometheus' },
      refresh: 2,  
      sort: 1,     
      multi: false,
      includeAll: true,
      allValue: '.*',
      current: { text: 'All', value: '.*' }
    }]
  }
}
+ g.dashboard.withPanels([
  // Golden Signals
  istio.rows.goldenSignalsRow(0),
  istio.panels.requestRatePanel(service_name, 'Request Rate', {x: 0, y: 1, w: 6, h: 6}, '$version_hash'),
  istio.panels.grpcErrorRatePanel(service_name, 'gRPC Server Errors', {x: 6, y: 1, w: 6, h: 6}, '$version_hash'),  
  istio.panels.httpClientErrorRatePanel(service_name, 'HTTP Client Errors', {x: 12, y: 1, w: 6, h: 6}, '$version_hash'),
  istio.panels.latencyP99Panel(service_name, 'P99 Response Time', {x: 18, y: 1, w: 6, h: 6}, '$version_hash'),

  // Saturation  
  istio.rows.saturationRow(7),
  istio.panels.cpuUtilizationPanel(service_name, 'CPU Utilization', {x: 0, y: 8, w: 6, h: 6}),
  istio.panels.memoryUtilizationPanel(service_name, 'Memory Utilization', {x: 6, y: 8, w: 6, h: 6}),
  istio.panels.rolloutVersionsPanel(service_name, 'Rollout Versions Distribution', {x: 18, y: 8, w: 6, h: 6}),
  
  // Trends
  istio.rows.trendsRow(14),
  istio.panels.latencyPanel(service_name, 'Latency Distribution', {x: 0, y: 15, w: 24, h: 6}, '$version_hash'),
  
  // Time series trends
  g.panel.timeSeries.new('Request Rate Over Time')
  + g.panel.timeSeries.datasource.withType('prometheus')
  + g.panel.timeSeries.datasource.withUid('prometheus')
  + g.panel.timeSeries.queryOptions.withTargets([
      g.query.prometheus.new('prometheus', istio.queries.requestRate(service_name, '5m', '$version_hash'))
      + g.query.prometheus.withLegendFormat('req/sec'),
  ])
  + g.panel.timeSeries.standardOptions.withUnit('reqps')
  + g.panel.timeSeries.gridPos.withH(8)
  + g.panel.timeSeries.gridPos.withW(12)
  + g.panel.timeSeries.gridPos.withX(0)
  + g.panel.timeSeries.gridPos.withY(21),

  g.panel.timeSeries.new('Error Rates Over Time')
  + g.panel.timeSeries.datasource.withType('prometheus')
  + g.panel.timeSeries.datasource.withUid('prometheus')
  + g.panel.timeSeries.queryOptions.withTargets([
      g.query.prometheus.new('prometheus', istio.queries.grpcErrorRateByCode(service_name, '5m', '$version_hash'))
      + g.query.prometheus.withLegendFormat('gRPC {{grpc_response_status}}'),
      g.query.prometheus.new('prometheus', istio.queries.httpClientErrorRateByCode(service_name, '5m', '$version_hash'))
      + g.query.prometheus.withLegendFormat('HTTP {{response_code}}'),
  ])
  + g.panel.timeSeries.standardOptions.withUnit('percent')
  + g.panel.timeSeries.standardOptions.withMin(0)
  + g.panel.timeSeries.gridPos.withH(8)
  + g.panel.timeSeries.gridPos.withW(12)
  + g.panel.timeSeries.gridPos.withX(12)
  + g.panel.timeSeries.gridPos.withY(21),
]);

dashboard