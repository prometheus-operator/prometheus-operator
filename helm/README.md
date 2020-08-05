# Moved
The helm charts originally developed as part of this repository have been moved to [stable/prometheus-operator](https://github.com/helm/charts/tree/master/stable/prometheus-operator).

The multiple charts built in this repository have been combined into a single chart that installs prometheus operator, prometheus, alertmanager, grafana as well as the multitude of exporters necessary to monitor a cluster.

There is no direct migration path from this chart to the stable/prometheus-operator chart - there are numerous changes and capability enhancements. For more information, please check the [migrated chart readme](https://github.com/helm/charts/tree/master/stable/prometheus-operator#prometheus-operator) for the set of components used and for [migration suggestions from this chart](https://github.com/helm/charts/tree/master/stable/prometheus-operator#migrating-from-coreosprometheus-operator-chart)

It is still possible to run multiple prometheus instances on a single cluster - you will need to disable the parts of the chart you do not wish to deploy.

Issues and pull requests should be tracked using the [helm/charts](https://github.com/helm/charts) repository.

You can check out the tickets for this change [here](https://github.com/prometheus-operator/prometheus-operator/issues/592) and [here](https://github.com/helm/charts/pull/6765)
