from grafanalib.core import *

dashboard = Dashboard(
    title='Kubernetes Control Plane Status',
    version=3,
    graphTooltip=0,
    schemaVersion=14,
    time=Time(start='now-6h'),
    timezone='browser',
    refresh=None,
    inputs=[
        {
            'name': 'DS_PROMETHEUS',
            'label': 'prometheus',
            'description': '',
            'type': 'datasource',
            'pluginId': 'prometheus',
            'pluginName': 'Prometheus'
        },
    ],
    rows=[
        Row(
            title='Dashboard Row', showTitle=False, titleSize='h6',
            panels=[
                SingleStat(
                    title='API Servers UP',
                    dataSource='${DS_PROMETHEUS}',
                    format='percent',
                    gauge=Gauge(
                        show=True,
                    ),
                    id=1,
                    span=3,
                    thresholds='50, 80',
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        }
                    ],
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        }
                    ],
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    targets=[
                        {
                            'expr': '(sum(up{job=\"apiserver\"} == 1) / '
                            'sum(up{job=\"apiserver\"})) * 100',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'refId': 'A',
                            'step': 600,
                        },
                    ]
                ),
                SingleStat(
                    title='Controller Managers UP',
                    dataSource='${DS_PROMETHEUS}',
                    format='percent',
                    gauge=Gauge(
                        show=True,
                    ),
                    id=2,
                    span=3,
                    thresholds='50, 80',
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        }
                    ],
                    rangeMaps=([
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ]),
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        }
                    ],
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    targets=[
                        {
                            'expr': '(sum(up{job=\"kube-controller-manager\"}'
                            ' == 1) / sum(up{job=\"kube-controller-manager\"'
                            '})) * 100',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'refId': 'A',
                            'step': 600,
                        }
                    ]
                ),
                SingleStat(
                    title='Schedulers UP',
                    dataSource='${DS_PROMETHEUS}',
                    format='percent',
                    gauge=Gauge(
                        show=True,
                    ),
                    id=3,
                    span=3,
                    thresholds='50, 80',
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        }
                    ],
                    rangeMaps=([
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ]),
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        }
                    ],
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    targets=[
                        {
                            'expr': '(sum(up{job=\"kube-scheduler\"} == 1) '
                            '/ sum(up{job=\"kube-scheduler\"})) * 100',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'refId': 'A',
                            'step': 600,
                        }
                    ]
                ),
                SingleStat(
                    title='API Server Request Error Rate',
                    dataSource='${DS_PROMETHEUS}',
                    format='percent',
                    gauge=Gauge(
                        show=True,
                    ),
                    id=4,
                    span=3,
                    thresholds='5, 10',
                    valueMaps=[
                        {
                            'op': '=',
                            'text': '0',
                            'value': 'null',
                        }
                    ],
                    rangeMaps=([
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ]),
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        }
                    ],
                    targets=[
                        {
                            'expr': 'max(sum by(instance) (rate('
                            'apiserver_request_count{code=~"5.."}[5m])) / '
                            'sum by(instance) (rate(apiserver_request_count'
                            '[5m]))) * 100',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ]
                ),
            ],
        ),
        Row(
            title='Dashboard Row', showTitle=False, titleSize='h6',
            panels=[
                Graph(
                    title='API Server Request Latency',
                    id=7,
                    dataSource='${DS_PROMETHEUS}',
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    lineWidth=1,
                    nullPointMode='null',
                    tooltip=Tooltip(
                        msResolution=False, valueType='individual',
                    ),
                    spaceLength=10,
                    yAxes=YAxes(
                        YAxis(format='short', min=None),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum by(verb) (rate(apiserver_latency_'
                            'seconds:quantile[5m]) >= 0)',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 30,
                        }
                    ],
                ),
            ],
        ),
        Row(
            title='Dashboard Row', showTitle=False, titleSize='h6',
            panels=[
                Graph(
                    title='End to End Scheduling Latency',
                    id=5,
                    dataSource='${DS_PROMETHEUS}',
                    isNew=False,
                    dashLength=10,
                    lineWidth=1,
                    nullPointMode="null",
                    spaceLength=10,
                    span=6,
                    dashes=False,
                    tooltip=Tooltip(
                        msResolution=False,
                        valueType='individual',
                    ),
                    yAxes=YAxes(
                        YAxis(format='short', min=None),
                        YAxis(format='dtdurations', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'cluster:scheduler_e2e_scheduling_'
                            'latency_seconds:quantile',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'refId': 'A',
                            'step': 60,
                        }
                    ],
                ),
                Graph(
                    title='API Server Request Rates',
                    id=6,
                    dataSource='${DS_PROMETHEUS}',
                    isNew=False,
                    dashLength=10,
                    lineWidth=1,
                    nullPointMode="null",
                    spaceLength=10,
                    span=6,
                    dashes=False,
                    tooltip=Tooltip(
                        msResolution=False,
                        valueType='individual',
                    ),
                    yAxes=YAxes(
                        YAxis(format='short', min=None),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum by(instance) (rate(apiserver_'
                            'request_count{code!~\"2..\"}[5m]))',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': 'Error Rate',
                            'refId': 'A',
                            'step': 60,
                        },
                        {
                            'expr': 'sum by(instance) (rate(apiserver_'
                            'request_count[5m]))',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': 'Request Rate',
                            'refId': 'B',
                            'step': 60,
                        },
                    ],
                ),
            ],
        ),
    ],
)
