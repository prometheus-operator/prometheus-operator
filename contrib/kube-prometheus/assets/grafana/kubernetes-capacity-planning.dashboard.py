from grafanalib.core import *


dashboard = Dashboard(
    title='Kubernetes Capacity Planning',
    version=4,
    gnetId=22,
    graphTooltip=0,
    refresh=False,
    editable=False,
    schemaVersion=14,
    time=Time(start='now-1h'),
    timezone='browser',
    inputs=[
        {
            'name': 'DS_PROMETHEUS',
            'label': 'prometheus',
            'description': '',
            'type': 'datasource',
            'pluginId': 'prometheus',
            'pluginName': 'Prometheus',
        }
    ],
    rows=[
        Row(
            height=250, title='New Row', showTitle=False, editable=False,
            titleSize='h6', panels=[
                Graph(
                    title='Idle CPU',
                    id=3,
                    dataSource='${DS_PROMETHEUS}',
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=6,
                    tooltip=Tooltip(msResolution=False),
                    yAxes=YAxes(
                        YAxis(format='percent', label='cpu usage',),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum(rate(node_cpu{mode=\"idle\"}[2m])) '
                            '* 100',
                            'hide': False,
                            'intervalFactor': 10,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 50,
                        },
                    ],
                ),
                Graph(
                    title='System Load',
                    id=9,
                    dataSource='${DS_PROMETHEUS}',
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=6,
                    tooltip=Tooltip(msResolution=False),
                    yAxes=YAxes(
                        YAxis(format='percentunit', min=None),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum(node_load1)',
                            'intervalFactor': 4,
                            'legendFormat': 'load 1m',
                            'refId': 'A',
                            'step': 20,
                            'target': '',
                        },
                        {
                            'expr': 'sum(node_load5)',
                            'intervalFactor': 4,
                            'legendFormat': 'load 5m',
                            'refId': 'B',
                            'step': 20,
                            'target': ''
                        },
                        {
                            'expr': 'sum(node_load15)',
                            'intervalFactor': 4,
                            'legendFormat': 'load 15m',
                            'refId': 'C',
                            'step': 20,
                            'target': '',
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=250, title='New Row', showTitle=False, editable=False,
            titleSize='h6', panels=[
                Graph(
                    title='Memory Usage',
                    id=4,
                    dataSource='${DS_PROMETHEUS}',
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=9,
                    stack=True,
                    seriesOverrides=[
                        {
                            'alias': 'node_memory_SwapFree{instance='
                            '\"172.17.0.1:9100\",job=\"prometheus\"}',
                            'yaxis': 2,
                        }
                    ],
                    tooltip=Tooltip(
                        msResolution=False, valueType='individual'
                    ),
                    yAxes=YAxes(
                        YAxis(format='bytes', min='0'),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum(node_memory_MemTotal) - sum(node_'
                            'memory_MemFree) - sum(node_memory_Buffers) - '
                            'sum(node_memory_Cached)',
                            'intervalFactor': 2,
                            'legendFormat': 'memory usage',
                            'metric': 'memo',
                            'refId': 'A',
                            'step': 10,
                            'target': '',
                        },
                        {
                            'expr': 'sum(node_memory_Buffers)',
                            'interval': '',
                            'intervalFactor': 2,
                            'legendFormat': 'memory buffers',
                            'metric': 'memo',
                            'refId': 'B',
                            'step': 10,
                            'target': '',
                        },
                        {
                            'expr': 'sum(node_memory_Cached)',
                            'interval': '',
                            'intervalFactor': 2,
                            'legendFormat': 'memory cached',
                            'metric': 'memo',
                            'refId': 'C',
                            'step': 10,
                            'target': '',
                        },
                        {
                            'expr': 'sum(node_memory_MemFree)',
                            'interval': '',
                            'intervalFactor': 2,
                            'legendFormat': 'memory free',
                            'metric': 'memo',
                            'refId': 'D',
                            'step': 10,
                            'target': '',
                        },
                    ],
                ),
                SingleStat(
                    title='Memory Usage',
                    dataSource='${DS_PROMETHEUS}',
                    id=5,
                    format='percent',
                    span=3,
                    gauge=Gauge(show=True),
                    editable=False,
                    thresholds='80, 90',
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    targets=[
                        {
                            'expr': '((sum(node_memory_MemTotal) - '
                            'sum(node_memory_MemFree) - sum('
                            'node_memory_Buffers) - sum(node_memory_Cached)) '
                            '/ sum(node_memory_MemTotal)) * 100',
                            'intervalFactor': 2,
                            'metric': '',
                            'refId': 'A',
                            'step': 60,
                            'target': '',
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=246, title='New Row', showTitle=False, editable=False,
            titleSize='h6', panels=[
                Graph(
                    title='Disk I/O',
                    dataSource='${DS_PROMETHEUS}',
                    id=6,
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=9,
                    tooltip=Tooltip(msResolution=False),
                    seriesOverrides=[
                        {
                            'alias': 'read',
                            'yaxis': 1
                        },
                        {
                            'alias': '{instance=\"172.17.0.1:9100\"}',
                            'yaxis': 2,
                        },
                        {
                            'alias': 'io time',
                            'yaxis': 2,
                        },
                    ],
                    yAxes=YAxes(
                        YAxis(format='bytes', min=None),
                        YAxis(format='ms', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum(rate(node_disk_bytes_read[5m]))',
                            'hide': False,
                            'intervalFactor': 4,
                            'legendFormat': 'read',
                            'refId': 'A',
                            'step': 20,
                            'target': ''
                        },
                        {
                            'expr': 'sum(rate(node_disk_bytes_written[5m]))',
                            'intervalFactor': 4,
                            'legendFormat': 'written',
                            'refId': 'B',
                            'step': 20
                        },
                        {
                            'expr': 'sum(rate(node_disk_io_time_ms[5m]))',
                            'intervalFactor': 4,
                            'legendFormat': 'io time',
                            'refId': 'C',
                            'step': 20
                        },
                    ],
                ),
                SingleStat(
                    title='Disk Space Usage',
                    dataSource='${DS_PROMETHEUS}',
                    id=12,
                    span=3,
                    editable=False,
                    format='percentunit',
                    valueName='current',
                    gauge=Gauge(
                        maxValue=1,
                        show=True,
                    ),
                    thresholds='0.75, 0.9',
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    targets=[
                        {
                            'expr': '(sum(node_filesystem_size{device!='
                            '\"rootfs\"}) - sum(node_filesystem_free{'
                            'device!=\"rootfs\"})) / sum(node_filesystem_size'
                            '{device!=\"rootfs\"})',
                            'intervalFactor': 2,
                            'refId': 'A',
                            'step': 60,
                            'target': '',
                        },
                    ],
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                ),
            ]
        ),
        Row(
            height=250, title='New Row', showTitle=False, editable=False,
            titleSize='h6', panels=[
                Graph(
                    title='Network Received',
                    dataSource='${DS_PROMETHEUS}',
                    id=8,
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=6,
                    tooltip=Tooltip(msResolution=False),
                    seriesOverrides=[
                        {
                            'alias': 'transmitted',
                            'yaxis': 2,
                        },
                    ],
                    yAxes=YAxes(
                        YAxis(format='bytes', min=None),
                        YAxis(format='bytes', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum(rate(node_network_receive_bytes'
                            '{device!~\"lo\"}[5m]))',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 10,
                            'target': '',
                        },
                    ],
                ),
                Graph(
                    title='Network Transmitted',
                    dataSource='${DS_PROMETHEUS}',
                    id=10,
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=6,
                    tooltip=Tooltip(msResolution=False),
                    seriesOverrides=[
                        {
                            'alias': 'transmitted',
                            'yaxis': 2,
                        },
                    ],
                    yAxes=YAxes(
                        YAxis(format='bytes', min=None),
                        YAxis(format='bytes', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum(rate(node_network_transmit_bytes'
                            '{device!~\"lo\"}[5m]))',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'B',
                            'step': 10,
                            'target': '',
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=276, title='New Row', showTitle=False, editable=False,
            titleSize='h6',
            panels=[
                Graph(
                    title='Cluster Pod Utilization',
                    dataSource='${DS_PROMETHEUS}',
                    id=11,
                    span=9,
                    dashes=False,
                    editable=False,
                    spaceLength=11,
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
                            'expr': 'sum(kube_pod_info)',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': 'Current number of Pods',
                            'refId': 'A',
                            'step': 10,
                        },
                        {
                            'expr': 'sum(kube_node_status_capacity_pods)',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': 'Maximum capacity of pods',
                            'refId': 'B',
                            'step': 10,
                        }
                    ],
                ),
                SingleStat(
                    title='Pod Utilization',
                    dataSource='${DS_PROMETHEUS}',
                    id=7,
                    editable=False,
                    span=3,
                    format='percent',
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    gauge=Gauge(
                        show=True,
                    ),
                    thresholds='80, 90',
                    valueName='current',
                    targets=[
                        {
                            'expr': '100 - (sum(kube_node_status_capacity_'
                            'pods) - sum(kube_pod_info)) / sum(kube_node_'
                            'status_capacity_pods) * 100',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 60,
                            'target': '',
                        },
                    ],
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                ),
            ]
        ),
    ],
)
