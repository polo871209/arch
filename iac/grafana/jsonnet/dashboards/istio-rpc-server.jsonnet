local istio = import '../lib/istio.libsonnet';
local g = import 'github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet';

local dashboard = g.dashboard.new('Istio Service Metrics')
+ g.dashboard.withDescription('Dashboard showing Istio service metrics with modular approach')
+ g.dashboard.withTags(['istio', 'microservices', 'sla'])
+ g.dashboard.time.withFrom('now-1h')
+ g.dashboard.time.withTo('now')
+ g.dashboard.withRefresh('30s')
+ g.dashboard.withTimezone('browser')
+ g.dashboard.withPanels([
  // RPC Server Success Rate Stat Panel
  istio.panels.successRatePanel(
    'rpc-server',
    'RPC Server Success Rate',
    {x: 0, y: 0, w: 6, h: 8}
  ),

  // RPC Server Success Rate Time Series
  istio.panels.successRateTimeSeriesPanel(
    'rpc-server',
    'RPC Server Success Rate Trend',
    {x: 6, y: 0, w: 18, h: 8}
  ),

  // Request Rate Panel
  g.panel.stat.new('RPC Server Request Rate')
  + g.panel.stat.datasource.withType('prometheus')
  + g.panel.stat.datasource.withUid('prometheus')
  + g.panel.stat.queryOptions.withTargets([
      g.query.prometheus.new(
        'prometheus',
        istio.queries.requestRate('rpc-server')
      )
      + g.query.prometheus.withFormat('time_series')
      + g.query.prometheus.withLegendFormat('Requests/sec'),
  ])
  + g.panel.stat.standardOptions.withUnit('reqps')
  + g.panel.stat.gridPos.withH(8)
  + g.panel.stat.gridPos.withW(6)
  + g.panel.stat.gridPos.withX(0)
  + g.panel.stat.gridPos.withY(8),

  // Error Rate Panel
  g.panel.stat.new('RPC Server Error Rate')
  + g.panel.stat.datasource.withType('prometheus')
  + g.panel.stat.datasource.withUid('prometheus')
  + g.panel.stat.queryOptions.withTargets([
      g.query.prometheus.new(
        'prometheus',
        istio.queries.errorRate('rpc-server')
      )
      + g.query.prometheus.withFormat('time_series')
      + g.query.prometheus.withLegendFormat('Error Rate'),
  ])
  + g.panel.stat.standardOptions.withUnit('percent')
  + g.panel.stat.standardOptions.withMin(0)
  + g.panel.stat.standardOptions.thresholds.withSteps([
      g.panel.stat.standardOptions.threshold.step.withColor('green') +
      g.panel.stat.standardOptions.threshold.step.withValue(0),
      g.panel.stat.standardOptions.threshold.step.withColor('yellow') +
      g.panel.stat.standardOptions.threshold.step.withValue(1),
      g.panel.stat.standardOptions.threshold.step.withColor('red') +
      g.panel.stat.standardOptions.threshold.step.withValue(5),
  ])
  + g.panel.stat.gridPos.withH(8)
  + g.panel.stat.gridPos.withW(6)
  + g.panel.stat.gridPos.withX(6)
  + g.panel.stat.gridPos.withY(8),

  // Request Rate Time Series
  g.panel.timeSeries.new('RPC Server Request Rate Over Time')
  + g.panel.timeSeries.datasource.withType('prometheus')
  + g.panel.timeSeries.datasource.withUid('prometheus')
  + g.panel.timeSeries.queryOptions.withTargets([
      g.query.prometheus.new(
        'prometheus',
        istio.queries.requestRate('rpc-server')
      )
      + g.query.prometheus.withFormat('time_series')
      + g.query.prometheus.withLegendFormat('{{destination_service_name}} Requests/sec'),
  ])
  + g.panel.timeSeries.standardOptions.withUnit('reqps')
  + g.panel.timeSeries.gridPos.withH(8)
  + g.panel.timeSeries.gridPos.withW(12)
  + g.panel.timeSeries.gridPos.withX(12)
  + g.panel.timeSeries.gridPos.withY(8),
]);

dashboard

