from grafanalib.core import *


dashboard = Dashboard(
    title='Kubernetes Cluster Health',
    version=9,
    graphTooltip=0,
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
    rows=[
        Row(
            height=254, title='Row', showTitle=False,
            titleSize='h6', panels=[
                SingleStat(
                    title='Control Plane Components Down',
                    id=1,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    span=3,
                    thresholds='1, 3',
                    colorValue=True,
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
                            'text': 'Everything UP and healthy',
                            'value': 'null',
                        },
                        {
                            'op': '=',
                            'text': '',
                            'value': '',
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
                        }
                    ],
                    targets=[
                        {
                            'expr': 'sum(up{job=~"apiserver|kube-scheduler|'
                            'kube-controller-manager"} == 0)',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
                SingleStat(
                    title='Alerts Firing',
                    id=2,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    colorValue=True,
                    span=3,
                    valueName='current',
                    thresholds='1, 3',
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
                            'expr': 'sum(ALERTS{alertstate="firing",'
                            'alertname!="DeadMansSwitch"})',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
                SingleStat(
                    title='Alerts Pending',
                    id=3,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    colorValue=True,
                    span=3,
                    valueName='current',
                    thresholds='3, 5',
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
                            'expr': 'sum(ALERTS{alertstate="pending",'
                            'alertname!="DeadMansSwitch"})',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
                SingleStat(
                    title='Crashlooping Pods',
                    id=4,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    colorValue=True,
                    span=3,
                    valueName='current',
                    thresholds='1, 3',
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
                            'expr': 'count(increase(kube_pod_container_'
                            'status_restarts[1h]) > 5)',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
            ],
        ),
        Row(
            height=250, title='Row', showTitle=False,
            titleSize='h6', panels=[
                SingleStat(
                    title='Node Not Ready',
                    id=5,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    colorValue=True,
                    span=3,
                    valueName='current',
                    thresholds='1, 3',
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
                        }
                    ],
                    targets=[
                        {
                            'expr': 'sum(kube_node_status_condition{'
                            'condition="Ready",status!="true"})',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
                SingleStat(
                    title='Node Disk Pressure',
                    id=6,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    colorValue=True,
                    span=3,
                    valueName='current',
                    thresholds='1, 3',
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
                        }
                    ],
                    targets=[
                        {
                            'expr': 'sum(kube_node_status_condition'
                            '{condition="DiskPressure",status="true"})',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
                SingleStat(
                    title='Node Memory Pressure',
                    id=7,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    colorValue=True,
                    span=3,
                    valueName='current',
                    thresholds='1, 3',
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
                        }
                    ],
                    targets=[
                        {
                            'expr': 'sum(kube_node_status_condition'
                            '{condition="MemoryPressure",status="true"})',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
                SingleStat(
                    title='Nodes Unschedulable',
                    id=8,
                    dataSource='${DS_PROMETHEUS}',
                    gauge=Gauge(),
                    colorValue=True,
                    span=3,
                    valueName='current',
                    thresholds='1, 3',
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
                        }
                    ],
                    targets=[
                        {
                            'expr': 'sum(kube_node_spec_unschedulable)',
                            'format': 'time_series',
                            'intervalFactor': 2,
                            'legendFormat': '',
                            'refId': 'A',
                            'step': 600,
                        },
                    ],
                ),
            ],
        ),
    ],
)
