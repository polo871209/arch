local g = import 'github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet';

{
  // Common Istio queries
  queries:: {
    successRate(service_name, interval='5m')::
      '(sum(irate(istio_requests_total{destination_service_name="' + service_name + '",response_code!~"5.."}[' + interval + '])) / sum(irate(istio_requests_total{destination_service_name="' + service_name + '"}[' + interval + ']))) * 100',

    requestRate(service_name, interval='5m')::
      'sum(irate(istio_requests_total{destination_service_name="' + service_name + '"}[' + interval + ']))',

    errorRate(service_name, interval='5m')::
      '(sum(irate(istio_requests_total{destination_service_name="' + service_name + '",response_code=~"5.."}[' + interval + '])) / sum(irate(istio_requests_total{destination_service_name="' + service_name + '"}[' + interval + ']))) * 100',
  },

  // Panel templates for Istio metrics
  panels:: {
    successRatePanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else service_name + ' Success Rate';
      g.panel.stat.new(panelTitle)
      + g.panel.stat.datasource.withType('prometheus')
      + g.panel.stat.datasource.withUid('prometheus')
      + g.panel.stat.queryOptions.withTargets([
          g.query.prometheus.new(
            'prometheus',
            $.queries.successRate(service_name)
          )
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('Success Rate'),
      ])
      + g.panel.stat.standardOptions.withUnit('percent')
      + g.panel.stat.standardOptions.withMin(0)
      + g.panel.stat.standardOptions.withMax(100)
      + g.panel.stat.standardOptions.thresholds.withSteps([
          g.panel.stat.standardOptions.threshold.step.withColor('red') +
          g.panel.stat.standardOptions.threshold.step.withValue(0),
          g.panel.stat.standardOptions.threshold.step.withColor('yellow') +
          g.panel.stat.standardOptions.threshold.step.withValue(95),
          g.panel.stat.standardOptions.threshold.step.withColor('green') +
          g.panel.stat.standardOptions.threshold.step.withValue(99),
      ])
      + g.panel.stat.gridPos.withH(if std.objectHas(gridPos, 'h') then gridPos.h else 8)
      + g.panel.stat.gridPos.withW(if std.objectHas(gridPos, 'w') then gridPos.w else 12)
      + g.panel.stat.gridPos.withX(if std.objectHas(gridPos, 'x') then gridPos.x else 0)
      + g.panel.stat.gridPos.withY(if std.objectHas(gridPos, 'y') then gridPos.y else 0),

    successRateTimeSeriesPanel(service_name, title=null, gridPos={})::
      local panelTitle = if title != null then title else service_name + ' Success Rate Over Time';
      g.panel.timeSeries.new(panelTitle)
      + g.panel.timeSeries.datasource.withType('prometheus')
      + g.panel.timeSeries.datasource.withUid('prometheus')
      + g.panel.timeSeries.queryOptions.withTargets([
          g.query.prometheus.new(
            'prometheus',
            $.queries.successRate(service_name)
          )
          + g.query.prometheus.withFormat('time_series')
          + g.query.prometheus.withLegendFormat('{{destination_service_name}} Success Rate'),
      ])
      + g.panel.timeSeries.standardOptions.withUnit('percent')
      + g.panel.timeSeries.standardOptions.withMin(0)
      + g.panel.timeSeries.standardOptions.withMax(100)
      + g.panel.timeSeries.gridPos.withH(if std.objectHas(gridPos, 'h') then gridPos.h else 8)
      + g.panel.timeSeries.gridPos.withW(if std.objectHas(gridPos, 'w') then gridPos.w else 12)
      + g.panel.timeSeries.gridPos.withX(if std.objectHas(gridPos, 'x') then gridPos.x else 0)
      + g.panel.timeSeries.gridPos.withY(if std.objectHas(gridPos, 'y') then gridPos.y else 0),
  },
}

