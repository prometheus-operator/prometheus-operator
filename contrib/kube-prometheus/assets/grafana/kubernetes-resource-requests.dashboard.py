from grafanalib.core import *


dashboard = Dashboard(
    title='Kubernetes Resource Requests',
    version=2,
    graphTooltip=0,
    refresh=False,
    editable=False,
    schemaVersion=14,
    time=Time(start='now-3h'),
    timezone='browser',
    inputs=[
        {
            'name': 'prometheus',
            'label': 'prometheus',
            'description': '',
            'type': 'datasource',
            'pluginId': 'prometheus',
            'pluginName': 'Prometheus'
        },
    ],
    rows=[
        Row(
            height=300, title='CPU Cores', showTitle=False, editable=False,
            titleSize='h6', panels=[
                Graph(
                    title='CPU Cores',
                    description='This represents the total [CPU resource '
                    'requests](https://kubernetes.io/docs/concepts/configu'
                    'ration/manage-compute-resources-container/#meaning-of-'
                    'cpu) in the cluster.\nFor comparison the total '
                    '[allocatable CPU cores](https://github.com/kubernetes/'
                    'community/blob/master/contributors/design-proposals/'
                    'node-allocatable.md) is also shown.',
                    id=1,
                    dataSource='prometheus',
                    dashLength=10,
                    dashes=False,
                    isNew=False,
                    editable=False,
                    lineWidth=1,
                    spaceLength=10,
                    nullPointMode='null',
                    span=9,
                    tooltip=Tooltip(
                        msResolution=False, valueType='individual'
                    ),
                    yAxes=YAxes(
                        YAxis(format='short', label='CPU Cores', min=None,),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'min(sum(kube_node_status_allocatable_'
                            'cpu_cores) by (instance))',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': 'Allocatable CPU Cores',
                            'refId': 'A',
                            'step': 20,
                        },
                        {
                            'expr': 'max(sum(kube_pod_container_resource_'
                            'requests_cpu_cores) by (instance))',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': 'Requested CPU Cores',
                            'refId': 'B',
                            'step': 20,
                        },
                    ],
                ),
                SingleStat(
                    title='CPU Cores',
                    dataSource='prometheus',
                    id=2,
                    format='percent',
                    editable=False,
                    span=3,
                    gauge=Gauge(show=True),
                    sparkline=SparkLine(show=True),
                    valueFontSize='110%',
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
                            'expr': 'max(sum(kube_pod_container_resource_'
                            'requests_cpu_cores) by (instance)) / min(sum'
                            '(kube_node_status_allocatable_cpu_cores) by '
                            '(instance)) * 100',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 240,
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=300, title='Memory', showTitle=False, editable=False,
            titleSize='h6', panels=[
                Graph(
                    title='Memory',
                    id=3,
                    dataSource='prometheus',
                    description='This represents the total [memory resource '
                    'requests](https://kubernetes.io/docs/concepts/'
                    'configuration/manage-compute-resources-container/'
                    '#meaning-of-memory) in the cluster.\nFor comparison '
                    'the total [allocatable memory](https://github.com/'
                    'kubernetes/community/blob/master/contributors/'
                    'design-proposals/node-allocatable.md) is also shown.',
                    dashLength=10,
                    dashes=False,
                    lineWidth=1,
                    isNew=False,
                    editable=False,
                    spaceLength=10,
                    span=9,
                    nullPointMode='null',
                    tooltip=Tooltip(
                        msResolution=False, valueType='individual'
                    ),
                    yAxes=YAxes(
                        YAxis(format='bytes', label='Memory', min=None),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'min(sum(kube_node_status_allocatable_'
                            'memory_bytes) by (instance))',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': 'Allocatable Memory',
                            'refId': 'A',
                            'step': 20,
                        },
                        {
                            'expr': 'max(sum(kube_pod_container_resource_'
                            'requests_memory_bytes) by (instance))',
                            'hide': False,
                            'intervalFactor': 2,
                            'legendFormat': 'Requested Memory',
                            'refId': 'B',
                            'step': 20,
                        },
                    ],
                ),
                SingleStat(
                    title='Memory',
                    dataSource='prometheus',
                    id=4,
                    format='percent',
                    span=3,
                    gauge=Gauge(show=True),
                    sparkline=SparkLine(show=True),
                    editable=False,
                    valueFontSize='110%',
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
                            'expr': 'max(sum(kube_pod_container_resource_'
                            'requests_memory_bytes) by (instance)) / '
                            'min(sum(kube_node_status_allocatable_memory_'
                            'bytes) by (instance)) * 100',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 240,
                        },
                    ],
                ),
            ],
        ),
    ],
)
