from grafanalib.core import *


dashboard = Dashboard(
    title='Pods',
    version=1,
    graphTooltip=1,
    refresh=False,
    schemaVersion=14,
    time=Time(start='now-6h'),
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
            'allValue': '.*',
            'current': {},
            'datasource': '${DS_PROMETHEUS}',
            'hide': 0,
            'includeAll': True,
            'label': 'Namespace',
            'multi': False,
            'name': 'namespace',
            'options': [],
            'query': 'label_values(kube_pod_info, namespace)',
            'refresh': 1,
            'regex': '',
            'sort': 0,
            'tagValuesQuery': '',
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
            'label': 'Pod',
            'multi': False,
            'name': 'pod',
            'options': [],
            'query': 'label_values(kube_pod_info{namespace=~"$namespace"}, '
            'pod)',
            'refresh': 1,
            'regex': '',
            'sort': 0,
            'tagValuesQuery': '',
            'tags': [],
            'tagsQuery': '',
            'type': 'query',
            'useTags': False,
        },
        {
            'allValue': '.*',
            'current': {},
            'datasource': '${DS_PROMETHEUS}',
            'hide': 0,
            'includeAll': True,
            'label': 'Container',
            'multi': False,
            'name': 'container',
            'options': [],
            'query': 'label_values(kube_pod_container_info{namespace='
            '"$namespace", pod="$pod"}, container)',
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
            height=250, title='Row', showTitle=False,
            titleSize='h6', panels=[
                Graph(
                    title='Memory Usage',
                    dataSource='${DS_PROMETHEUS}',
                    id=1,
                    isNew=False,
                    spaceLength=10,
                    span=12,
                    dashLength=10,
                    dashes=False,
                    tooltip=Tooltip(msResolution=True, valueType='cumulative'),
                    legend=Legend(
                        alignAsTable=True, avg=True, current=True,
                        rightSide=True, total=False, values=True,
                    ),
                    yAxes=YAxes(
                        YAxis(
                            format='bytes', min=None,
                        ),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum by(container_name) (container_'
                            'memory_usage_bytes{pod_name="$pod", '
                            'container_name=~"$container", '
                            'container_name!="POD"})',
                            'interval': '10s',
                            'intervalFactor': 1,
                            'legendFormat': 'Current: {{ container_name }}',
                            'metric': 'container_memory_usage_bytes',
                            'refId': 'A',
                            'step': 15,
                        },
                        {
                            'expr': 'kube_pod_container_resource_requests_'
                            'memory_bytes{pod="$pod", container=~'
                            '"$container"}',
                            'interval': '10s',
                            'intervalFactor': 2,
                            'legendFormat': 'Requested: {{ container }}',
                            'metric': 'kube_pod_container_resource_'
                            'requests_memory_bytes',
                            'refId': 'B',
                            'step': 20,
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=250, title='Row', showTitle=False,
            titleSize='h6', panels=[
                Graph(
                    title='CPU Usage',
                    dataSource='${DS_PROMETHEUS}',
                    id=2,
                    isNew=False,
                    spaceLength=10,
                    span=12,
                    dashLength=10,
                    dashes=False,
                    legend=Legend(
                        alignAsTable=True, avg=True, current=True,
                        rightSide=True, total=False, values=True,
                    ),
                    tooltip=Tooltip(msResolution=True, valueType='cumulative'),
                    yAxes=YAxes(
                        YAxis(
                            format='short', min=None,
                        ),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sum by (container_name)('
                            'rate(container_cpu_usage_seconds_total'
                            '{image!="",container_name!="POD",pod_name='
                            '"$pod"}[1m]))',
                            'intervalFactor': 2,
                            'legendFormat': '{{ container_name }}',
                            'refId': 'A',
                            'step': 30
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=250, title='New Row', showTitle=False,
            titleSize='h6', panels=[
                Graph(
                    title='Network I/O',
                    dataSource='${DS_PROMETHEUS}',
                    id=3,
                    isNew=False,
                    spaceLength=10,
                    span=12,
                    dashLength=10,
                    dashes=False,
                    legend=Legend(
                        alignAsTable=True, avg=True, current=True,
                        rightSide=True, total=False, values=True,
                    ),
                    tooltip=Tooltip(msResolution=True, valueType='cumulative'),
                    yAxes=YAxes(
                        YAxis(
                            format='bytes', min=None,
                        ),
                        YAxis(format='short', min=None),
                    ),
                    targets=[
                        {
                            'expr': 'sort_desc(sum by (pod_name) (rate'
                            '(container_network_receive_bytes_total{'
                            'pod_name="$pod"}[1m])))',
                            'intervalFactor': 2,
                            'legendFormat': '{{ pod_name }}',
                            'refId': 'A',
                            'step': 30
                        },
                    ],
                ),
            ],
        ),
    ],
)
