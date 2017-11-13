import sys
import os.path
sys.path.insert(0, os.path.dirname(__file__))
from _grafanalib import *


dashboard = Dashboard(
    title='StatefulSet',
    version=1,
    graphTooltip=1,
    time=Time(start='now-6h'),
    templating=Templating(list=[
        {
            'allValue': '.*',
            'current': {},
            'datasource': '${DS_PROMETHEUS}',
            'hide': 0,
            'includeAll': False,
            'label': 'Namespace',
            'multi': False,
            'name': 'statefulset_namespace',
            'options': [],
            'query': 'label_values(kube_statefulset_metadata_generation, '
            'namespace)',
            'refresh': 1,
            'regex': '',
            'sort': 0,
            'tagValuesQuery': None,
            'tags': [],
            'tagsQuery': '',
            'type': 'query',
            'useTags': False,
        },
        {
            'allValue': None,
            'current': {},
            'datasource': '${DS_PROMETHEUS}',
            'hide': 0,
            'includeAll': False,
            'label': 'StatefulSet',
            'multi': False,
            'name': 'statefulset_name',
            'options': [],
            'query': 'label_values(kube_statefulset_metadata_generation'
            '{namespace="$statefulset_namespace"}, statefulset)',
            'refresh': 1,
            'regex': '',
            'sort': 0,
            'tagValuesQuery': '',
            'tags': [],
            'tagsQuery': 'statefulset',
            'type': 'query',
            'useTags': False,
        },
    ]),
    rows=[
        Row(panels=[
            SingleStat(
                title='CPU',
                id=8,
                gauge=Gauge(show=False),
                postfix='cores',
                span=4,
                valueFontSize='110%',
                mappingType=1,
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
                colors=[
                    (245, 54, 54, 0.9),
                    (237, 129, 40, 0.89),
                    (50, 172, 45, 0.97),
                ],
                sparkline=SparkLine(
                    fillColor=(31, 118, 189, 0.18),
                    lineColor=(31, 120, 193),
                    show=True,
                ),
                targets=[
                    {
                        'expr': 'sum(rate(container_cpu_usage_seconds_total'
                        '{namespace=\"$statefulset_namespace\",pod_name=~\"'
                        '$statefulset_name.*\"}[3m]))',
                    },
                ],
            ),
            SingleStat(
                title='Memory',
                id=9,
                postfix='GB',
                prefixFontSize='80%',
                gauge=Gauge(show=False),
                span=4,
                valueFontSize='110%',
                mappingType=1,
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
                sparkline=SparkLine(
                    fillColor=(31, 118, 189, 0.18),
                    lineColor=(31, 120, 193),
                    show=True,
                ),
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
                colors=[
                    (245, 54, 54, 0.9),
                    (237, 129, 40, 0.89),
                    (50, 172, 45, 0.97),
                ],
                targets=[
                    {
                        'expr': 'sum(container_memory_usage_bytes{namespace='
                        '\"$statefulset_namespace\",pod_name=~\"$'
                        'statefulset_name.*\"}) / 1024^3',
                        'intervalFactor': 2,
                        'refId': 'A',
                        'step': 600,
                    },
                ],
            ),
            SingleStat(
                title='Network',
                format='Bps',
                gauge=Gauge(thresholdMarkers=False),
                id=7,
                postfix='',
                span=4,
                mappingType=1,
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
                sparkline=SparkLine(
                    fillColor=(31, 118, 189, 0.18),
                    lineColor=(31, 120, 193),
                    show=True,
                ),
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
                colors=[
                    (245, 54, 54, 0.9),
                    (237, 129, 40, 0.89),
                    (50, 172, 45, 0.97),
                ],
                targets=[
                    {
                        'expr': 'sum(rate(container_network_transmit_'
                        'bytes_total'
                        '{namespace=\"$statefulset_namespace\",pod_name=~\"'
                        '$statefulset_name.*\"}[3m])) + '
                        'sum(rate(container_network_receive_bytes_total'
                        '{namespace=\"$statefulset_namespace\",pod_name=~'
                        '\"$statefulset_name.*\"}[3m]))',
                    },
                ],
            ),
        ],
            height=200,
        ),
        Row(
            height=100, panels=[
                SingleStat(
                    title='Desired Replicas',
                    id=5,
                    mappingType=1,
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
                    span=3,
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    targets=[
                        {
                            'metric': 'kube_statefulset_replicas',
                            'expr': 'max(kube_statefulset_replicas'
                            '{statefulset="$statefulset_name",namespace='
                            '"$statefulset_namespace"}) without '
                            '(instance, pod)',
                        },
                    ],
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                    gauge=Gauge(thresholdMarkers=False, show=False),
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                ),
                SingleStat(
                    title='Available Replicas',
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    gauge=Gauge(show=False),
                    id=6,
                    mappingType=1,
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
                            'expr': 'min(kube_statefulset_status_replicas'
                            '{statefulset=\"$statefulset_name\",'
                            'namespace=\"$statefulset_namespace\"}) without '
                            '(instance, pod)',
                        },
                    ],
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    span=3,
                    sparkline=SparkLine(),
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        }
                    ],
                ),
                SingleStat(
                    title='Observed Generation',
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    gauge=Gauge(),
                    id=3,
                    mappingType=1,
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
                            'expr': 'max(kube_statefulset_status_observed_'
                            'generation{statefulset=\"$statefulset_name\",'
                            'namespace=\"$statefulset_namespace\"}) without '
                            '(instance, pod)',
                        },
                    ],
                    rangeMaps=[
                        {
                            'from': "null",
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    span=3,
                    sparkline=SparkLine(),
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        }
                    ],
                ),
                SingleStat(
                    title='Metadata Generation',
                    colors=[
                        (245, 54, 54, 0.9),
                        (237, 129, 40, 0.89),
                        (50, 172, 45, 0.97),
                    ],
                    gauge=Gauge(show=False),
                    id=2,
                    mappingType=1,
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
                            'expr': 'max(kube_statefulset_metadata_generation'
                            '{statefulset=\"$statefulset_name\",namespace=\"'
                            '$statefulset_namespace\"}) without (instance, '
                            'pod)',
                        },
                    ],
                    rangeMaps=[
                        {
                            'from': 'null',
                            'text': 'N/A',
                            'to': 'null',
                        },
                    ],
                    span=3,
                    sparkline=SparkLine(),
                    valueMaps=[
                        {
                            'op': '=',
                            'text': 'N/A',
                            'value': 'null',
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=350, panels=[
                Graph(
                    title='Replicas',
                    dashLength=10,
                    dashes=False,
                    id=1,
                    spaceLength=10,
                    targets=[
                        {
                            'expr': 'min(kube_statefulset_status_replicas'
                            '{statefulset=\"$statefulset_name\",'
                            'namespace=\"$statefulset_namespace\"}) without '
                            '(instance, pod)',
                            'legendFormat': 'available',
                            'refId': 'B',
                            'step': 30,
                        },
                        {
                            'expr': 'max(kube_statefulset_replicas'
                            '{statefulset=\"$statefulset_name\",namespace=\"'
                            '$statefulset_namespace\"}) without '
                            '(instance, pod)',
                            'legendFormat': 'desired',
                            'refId': 'E',
                            'step': 30,
                        }
                    ],
                    xAxis=XAxis(mode='time'),
                    yAxes=YAxes(
                        YAxis(min=None),
                        YAxis(format='short', min=None, show=False),
                    ),
                ),
            ]
        ),
    ],
)
