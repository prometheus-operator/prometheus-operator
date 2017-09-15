from grafanalib import core
from grafanalib.core import Graph, Time, SparkLine, \
    Gauge, Templating, XAxis, YAxes


def Dashboard(
    title, version, time, rows, graphTooltip=0, templating=None,
):
    optional_args = {}
    if templating is not None:
        optional_args['templating'] = templating
    return core.Dashboard(
        title=title, refresh=None, schemaVersion=14,
        version=version, time=time, timezone='browser', inputs=[
            {
                'name': 'DS_PROMETHEUS',
                'label': 'prometheus',
                'description': '',
                'type': 'datasource',
                'pluginId': 'prometheus',
                'pluginName': 'Prometheus'
            },
        ], rows=rows, graphTooltip=graphTooltip, **optional_args,
    )


def Row(
    panels, height=None, title='Dashboard Row', showTitle=False,
    editable=None
):
    assert isinstance(height, (type(None), int))
    return core.Row(
        panels=panels, height=height, title=title, showTitle=showTitle,
        titleSize='h6', editable=editable,
    )


def SingleStat(
    title, id, targets, colorValue=False, gauge=Gauge(show=True),
    valueFontSize='80%', thresholds=None, valueName='avg', valueMaps=None,
    rangeMaps=None, mappingTypes=None, mappingType=None, postfix=None,
    sparkline=SparkLine(), prefixFontSize='50%', colors=[
        (50, 172, 45, 0.97),
        (237, 129, 40, 0.89),
        (245, 54, 54, 0.9),
    ], span=None, format='none', transparent=None,
):
    def merge_target(target):
        return {**{
            'intervalFactor': 2,
            'refId': 'A',
            'step': 600,
        }, **target}
    targets = [merge_target(t) for t in targets]

    return core.SingleStat(
        title=title, id=id, colorValue=colorValue,
        dataSource='${DS_PROMETHEUS}', gauge=gauge,
        valueFontSize=valueFontSize, thresholds=thresholds,
        valueName=valueName, valueMaps=valueMaps, rangeMaps=rangeMaps,
        mappingTypes=mappingTypes, targets=targets,
        mappingType=mappingType, format=format, colors=colors, span=span,
        postfix=postfix, sparkline=sparkline, prefixFontSize=prefixFontSize,
        hideTimeOverride=None, transparent=transparent,
    )


def Graph(
    id, title, targets, dashLength=None, dashes=False, spaceLength=None,
    xAxis=None, yAxes=None, nullPointMode='connected',
):
    def merge_target(target):
        return {**{
            'intervalFactor': 2,
            'legendFormat': '',
            'refId': 'A',
            'step': 600,
        }, **target}

    targets = [merge_target(t) for t in targets]
    assert isinstance(yAxes, YAxes)
    return core.Graph(
        id=id, title=title, dashLength=dashLength, dashes=dashes,
        spaceLength=spaceLength, targets=targets, xAxis=xAxis, yAxes=yAxes,
        dataSource='${DS_PROMETHEUS}', nullPointMode=nullPointMode,
    )


def YAxis(format='none', label='', min=0, show=True):
    return core.YAxis(
        format=format, label=label, min=min, show=show
    )
