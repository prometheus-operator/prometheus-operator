# Moved
The helm charts originally developed as part of this repository have been moved to [stable/prometheus-operator](https://github.com/helm/charts/tree/master/stable/prometheus-operator).

The multiple charts built in this repository have been combined into a single chart that installs prometheus operator, prometheus, alertmanager, grafana as well as the multitude of exporters necessary to monitor a cluster.

There is no direct migration path from this chart to the stable/prometheus-operator chart - there are numerous changes and capability enhancements.

It is still possible to run multiple prometheus instances on a single cluster - you will need to disable the parts of the chart you do not wish to deploy.

Issues and pull requests should be tracked using the [helm/charts](https://github.com/helm/charts) repository.

You can check out the tickets for this change [here](https://github.com/coreos/prometheus-operator/issues/592) and [here](https://github.com/helm/charts/pull/6765)

## Changes
The chart has 3 dependencies, that can be seen in the chart's requirements file:
https://github.com/helm/charts/blob/master/stable/prometheus-operator/requirements.yaml

### Node-Exporter, Kube-State-Metrics
These components are loaded as dependencies into the chart. The source for both charts is found in the same repository. They are relatively simple components.

### Grafana
The Grafana chart is more feature-rich than this chart - it contains a sidecard that is able to load data sources and dashboards from configmaps deployed into the same cluster. For more information check out the [documentatin for the chart](https://github.com/helm/charts/tree/master/stable/grafana)

### Coreos CRDs
The CRDs are provisioned using crd-install hooks, rather than relying on a separate chart installation. If you already have these CRDs provisioned and don't want to remove them, you can disable the CRD creation by these hooks by passing `prometheusOperator.createCustomResource=false`

### Kubelet Service
Because the kubelet service has a new name in the chart, make sure to clean up the old kubelet service in the `kube-system` namespace to prevent counting container metrics twice

### Persistent Volumes
If you would like to keep the data of the current persistent volumes, it should realistically be possible to appropriately rename and relabel the PVs and PVCs from their convention using the old chart to the new chart conventions. Alternatively you could copy the data from the PVCs using mechanisms available in your cluster