from grafanalib.core import *


dashboard = Dashboard(
    title='Nodes',
    version=2,
    description='Dashboard to get an overview of one server',
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
            'pluginName': 'Prometheus'
        },
    ],
    templating=Templating(list=[
        {
            'allValue': None,
            'current': {},
            'datasource': '${DS_PROMETHEUS}',
            'hide': 0,
            'includeAll': False,
            'label': None,
            'multi': False,
            'name': 'server',
            'options': [],
            'query': 'label_values(node_boot_time, instance)',
            'refresh': 1,
            'regex': '',
            'sort': 0,
            'tagValuesQuery': '',
            'tags': [],
            'tagsQuery': '',
            'type': 'query',
            'useTags': False,
        },
    ]),
    rows=[
        Row(
            height=250, title='New Row', showTitle=False, editable=False,
            titleSize='h6', panels=[
                Graph(
                    title='Idle CPU',
                    dataSource='${DS_PROMETHEUS}',
                    id=3,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=6,
                    dashLength=10,
                    dashes=False,
                    tooltip=Tooltip(msResolution=False),
                    yAxes=YAxes(
                        YAxis(
                            format='percent',
                            label='cpu usage',
                            max=100,
                        ),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': '100 - (avg by (cpu) (irate(node_cpu'
                            '{mode=\"idle\", instance=\"$server\"}[5m])) '
                            '* 100)',
                            'hide': False,
                            'intervalFactor': 10,
                            'legendFormat': '{{cpu}}',
                            'refId': 'A',
                            'step': 50,
                        }
                    ],
                ),
                Graph(
                    title='System Load',
                    dataSource='${DS_PROMETHEUS}',
                    id=9,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=6,
                    dashLength=10,
                    dashes=False,
                    tooltip=Tooltip(msResolution=False),
                    yAxes=YAxes(
                        YAxis(format='percentunit', min=None,),
                        YAxis(format='short', min=None,),
                    ),
                    targets=[
                        {
                            'expr': 'node_load1{instance=\"$server\"}',
                            'intervalFactor': 4,
                            'legendFormat': 'load 1m',
                            'refId': 'A',
                            'step': 20,
                            'target': '',
                        },
                        {
                            'expr': 'node_load5{instance=\"$server\"}',
                            'intervalFactor': 4,
                            'legendFormat': 'load 5m',
                            'refId': 'B',
                            'step': 20,
                            'target': '',
                        },
                        {
                            'expr': 'node_load15{instance=\"$server\"}',
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
                    dataSource='${DS_PROMETHEUS}',
                    id=4,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=9,
                    stack=True,
                    dashLength=10,
                    dashes=False,
                    tooltip=Tooltip(
                        msResolution=False, valueType='individual',
                    ),
                    seriesOverrides=[
                        {
                            'alias': 'node_memory_SwapFree{instance='
                            '\"172.17.0.1:9100\",job=\"prometheus\"}',
                            'yaxis': 2,
                        },
                    ],
                    yAxes=YAxes(
                        YAxis(format='bytes', min='0',),
                        YAxis(format='short', min=None,),
                    ),
                    targets=[
                        {
                            'expr': 'node_memory_MemTotal{instance='
                            '\"$server\"} - node_memory_MemFree{instance='
                            '\"$server\"} - node_memory_Buffers{instance='
                            '\"$server\"} - node_memory_Cached{instance='
                            '\"$server\"}',
                            'hide': False,
                            'interval': '',
                            'intervalFactor': 2,
                            'legendFormat': 'memory used',
                            'metric': '',
                            'refId': 'C',
                            'step': 10,
                        },
                        {
                            'expr': 'node_memory_Buffers{instance='
                            '\"$server\"}',
                            'interval': '',
                            'intervalFactor': 2,
                            'legendFormat': 'memory buffers',
                            'metric': '',
                            'refId': 'E',
                            'step': 10,
                        },
                        {
                            'expr': 'node_memory_Cached{instance=\"$server\"}',
                            'intervalFactor': 2,
                            'legendFormat': 'memory cached',
                            'metric': '',
                            'refId': 'F',
                            'step': 10,
                        },
                        {
                            'expr': 'node_memory_MemFree{instance='
                            '\"$server\"}',
                            'intervalFactor': 2,
                            'legendFormat': 'memory free',
                            'metric': '',
                            'refId': 'D',
                            'step': 10,
                        },
                    ],
                ),
                SingleStat(
                    title='Memory Usage',
                    dataSource='${DS_PROMETHEUS}',
                    id=5,
                    format='percent',
                    gauge=Gauge(show=True),
                    editable=False,
                    span=3,
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        }
                    ],
                    thresholds='80, 90',
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                    targets=[
                        {
                            'expr': '((node_memory_MemTotal{instance='
                            '\"$server\"} - node_memory_MemFree{instance='
                            '\"$server\"}  - node_memory_Buffers{instance='
                            '\"$server\"} - node_memory_Cached{instance='
                            '\"$server\"}) / node_memory_MemTotal{instance='
                            '\"$server\"}) * 100',
                            'intervalFactor': 2,
                            'refId': 'A',
                            'step': 60,
                            'target': '',
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=250, titleSize='h6', title='New Row', editable=False,
            showTitle=False, panels=[
                Graph(
                    title='Disk I/O',
                    dataSource='${DS_PROMETHEUS}',
                    id=6,
                    dashLength=10,
                    dashes=False,
                    editable=False,
                    spaceLength=10,
                    span=9,
                    tooltip=Tooltip(msResolution=False),
                    yAxes=YAxes(
                        YAxis(
                            format='bytes',
                            min=None,
                        ),
                        YAxis(
                            format='ms',
                            min=None,
                        ),
                    ),
                    seriesOverrides=[
                        {
                            'alias': 'read',
                            'yaxis': 1,
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
                    targets=[
                        {
                            'expr': 'sum by (instance) (rate(node_disk_'
                            'bytes_read{instance=\"$server\"}[2m]))',
                            'hide': False,
                            'intervalFactor': 4,
                            'legendFormat': 'read',
                            'refId': 'A',
                            'step': 20,
                            'target': '',
                        },
                        {
                            'expr': 'sum by (instance) (rate(node_disk_'
                            'bytes_written{instance=\"$server\"}[2m]))',
                            'intervalFactor': 4,
                            'legendFormat': 'written',
                            'refId': 'B',
                            'step': 20
                        },
                        {
                            'expr': 'sum by (instance) (rate(node_disk_io_'
                            'time_ms{instance=\"$server\"}[2m]))',
                            'intervalFactor': 4,
                            'legendFormat': 'io time',
                            'refId': 'C',
                            'step': 20,
                        },
                    ],
                ),
                SingleStat(
                    title='Disk Space Usage',
                    dataSource='${DS_PROMETHEUS}',
                    id=7,
                    thresholds='0.75, 0.9',
                    editable=False,
                    valueName='current',
                    format='percentunit',
                    span=3,
                    gauge=Gauge(
                        maxValue=1,
                        show=True,
                    ),
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
                    targets=[
                        {
                            'expr': '(sum(node_filesystem_size{device!='
                            '\"rootfs\",instance=\"$server\"}) - '
                            'sum(node_filesystem_free{device!=\"rootfs\",'
                            'instance=\"$server\"})) / sum(node_filesystem_'
                            'size{device!=\"rootfs\",instance=\"$server\"})',
                            'intervalFactor': 2,
                            'refId': 'A',
                            'step': 60,
                            'target': '',
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=250, title='New Row', titleSize='h6',
            showTitle=False, editable=False,
            panels=[
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
                    yAxes=YAxes(
                        YAxis(format='bytes', min=None),
                        YAxis(format='bytes', min=None),
                    ),
                    seriesOverrides=[
                        {
                            'alias': 'transmitted',
                            'yaxis': 2,
                        },
                    ],
                    targets=[
                        {
                            'expr': 'rate(node_network_receive_bytes{'
                            'instance=\"$server\",device!~\"lo\"}[5m])',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': '{{device}}',
                            'refId': 'A',
                            'step': 10,
                            'target': ''
                        }
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
                    yAxes=YAxes(
                        YAxis(format='bytes', min=None),
                        YAxis(format='bytes', min=None),
                    ),
                    seriesOverrides=[
                        {
                            'alias': 'transmitted',
                            'yaxis': 2,
                        },
                    ],
                    targets=[
                        {
                            'expr': 'rate(node_network_transmit_bytes'
                            '{instance=\"$server\",device!~\"lo\"}[5m])',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': '{{device}}',
                            'refId': 'B',
                            'step': 10,
                            'target': '',
                        },
                    ],
                ),
            ],
        ),
    ],
)
