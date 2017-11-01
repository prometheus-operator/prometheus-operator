import sys
import os.path
sys.path.insert(0, os.path.dirname(__file__))
from _grafanalib import *


dashboard = Dashboard(
    title='Kubernetes Cluster Status',
    version=3,
    time=Time(start='now-6h'),
    rows=[
        Row(
            height=129, title='Cluster Health', showTitle=True,
            panels=[
                SingleStat(
                    title='Control Plane UP',
                    id=5,
                    gauge=Gauge(show=False),
                    colorValue=True,
                    mappingType=1,
                    thresholds='1, 3',
                    valueName='total',
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
                            'text': 'UP',
                            'value': 'null',
                        },
                    ],
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
                    targets=[
                        {
                            'expr': 'sum(up{job=~"apiserver|kube-scheduler|'
                            'kube-controller-manager"} == 0)',
                            'format': 'time_series',
                        },
                    ]
                ),
                SingleStat(
                    title='Alerts Firing',
                    id=6,
                    gauge=Gauge(show=False),
                    colorValue=True,
                    mappingType=1,
                    thresholds='3, 5',
                    valueName='current',
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
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
                            'text': '0',
                            'value': 'null',
                        },
                    ],
                    targets=[
                        {
                            'expr': 'sum(ALERTS{alertstate="firing",'
                            'alertname!="DeadMansSwitch"})',
                            'format': 'time_series',
                        },
                    ]
                ),
            ],
        ),
        Row(
            height=168, title='Control Plane Status', showTitle=True,
            panels=[
                SingleStat(
                    title='API Servers UP',
                    id=1,
                    mappingType=1,
                    format='percent',
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    thresholds='50, 80',
                    span=3,
                    valueName='current',
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
                        },
                    ],
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
                    targets=[
                        {
                            'expr': '(sum(up{job="apiserver"} == 1) / '
                            'count(up{job="apiserver"})) * 100',
                            'format': 'time_series',
                        },
                    ]
                ),
                SingleStat(
                    title='Controller Managers UP',
                    id=2,
                    span=3,
                    mappingType=1,
                    thresholds='50, 80',
                    format='percent',
                    valueName='current',
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                    targets=[
                        {
                            'expr': '(sum(up{job="kube-controller-manager"} =='
                            ' 1) / count(up{job="kube-controller-manager"})) '
                            '* 100',
                            'format': 'time_series',
                        },
                    ]
                ),
                SingleStat(
                    title='Schedulers UP',
                    id=3,
                    span=3,
                    mappingType=1,
                    format='percent',
                    thresholds='50, 80',
                    valueName='current',
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                    targets=[
                        {
                            'expr': '(sum(up{job="kube-scheduler"} == 1) / '
                            'count(up{job="kube-scheduler"})) * 100',
                            'format': 'time_series',
                        },
                    ]
                ),
                SingleStat(
                    title='Crashlooping Control Plane Pods',
                    id=4,
                    colorValue=True,
                    gauge=Gauge(show=False),
                    span=3,
                    mappingType=1,
                    thresholds='1, 3',
                    valueName='current',
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
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
                            'text': '0',
                            'value': 'null',
                        },
                    ],
                    targets=[
                        {
                            'expr': 'count(increase(kube_pod_container_'
                            'status_restarts{namespace=~"kube-system|'
                            'tectonic-system"}[1h]) > 5)',
                            'format': 'time_series',
                        },
                    ]
                ),
            ],
        ),
        Row(
            height=158, title='Capacity Planning', showTitle=True,
            panels=[
                SingleStat(
                    title='CPU Utilization',
                    id=8,
                    format='percent',
                    mappingType=1,
                    span=3,
                    thresholds='80, 90',
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
                        },
                    ],
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
                    targets=[
                        {
                            'expr': 'sum(100 - (avg by (instance) (rate('
                            'node_cpu{job="node-exporter",mode="idle"}[5m])) '
                            '* 100)) / count(node_cpu{job="node-exporter",'
                            'mode="idle"})',
                            'format': 'time_series',
                        },
                    ]
                ),
                SingleStat(
                    title='Memory Utilization',
                    id=7,
                    format='percent',
                    span=3,
                    mappingType=1,
                    thresholds='80, 90',
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
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
                            'expr': '((sum(node_memory_MemTotal) - sum('
                            'node_memory_MemFree) - sum(node_memory_Buffers) '
                            '- sum(node_memory_Cached)) / sum('
                            'node_memory_MemTotal)) * 100',
                            'format': 'time_series',
                        },
                    ]
                ),
                SingleStat(
                    title='Filesystem Utilization',
                    id=9,
                    span=3,
                    format='percent',
                    mappingType=1,
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
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
                    targets=[
                        {
                            'expr': '(sum(node_filesystem_size{device!='
                            '"rootfs"}) - sum(node_filesystem_free{device!='
                            '"rootfs"})) / sum(node_filesystem_size{device!='
                            '"rootfs"})',
                            'format': 'time_series',
                        },
                    ]
                ),
                SingleStat(
                    title='Pod Utilization',
                    id=10,
                    gauge=Gauge(show=True),
                    span=3,
                    mappingType=1,
                    format='percent',
                    thresholds='80, 90',
                    mappingTypes=[
                        {
                            'name': 'value to text',
                            'value': 1,
                        },
                        {
                            'name': 'range to text',
                            'value': 2,
                        },
                    ],
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
                            'expr': '100 - (sum(kube_node_status_capacity_pods'
                            ') - sum(kube_pod_info)) / sum(kube_node_status_'
                            'capacity_pods) * 100',
                            'format': 'time_series',
                        },
                    ]
                ),
            ],
        ),
    ],
)
