# Monitoring external etcd
This guide will help you monitor an external etcd cluster. When the etcd cluster is not hosted inside Kubernetes.
This is often the case with Kubernetes setups. This approach has been tested with kube-aws but the same principals apply to other tools.

Note that [etcd.jsonnet](../examples/etcd.jsonnet) & [kube-prometheus-static-etcd.libsonnet](../jsonnet/kube-prometheus/kube-prometheus-static-etcd.libsonnet) (which are described by a section of the [Readme](../README.md#static-etcd-configuration)) do the following:
 * Put the three etcd TLS client files (CA & cert & key) into a secret in the namespace, and have Prometheus Operator load the secret.
 * Create the following (to expose etcd metrics - port 2379): a Service, Endpoint, & ServiceMonitor.

# Step 1: Open the port

You now need to allow the nodes Prometheus are running on to talk to the etcd on the port 2379 (if 2379 is the port used by etcd to expose the metrics)

If using kube-aws, you will need to edit the etcd security group inbound, specifying the security group of your Kubernetes node (worker) as the source.

## kube-aws and EIP or ENI inconsistency
With kube-aws, each etcd node has two IP addresses:

* EC2 instance IP
* EIP or ENI (depending on the chosen method in yuour cluster.yaml)

For some reason, some etcd node answer to :2379/metrics on the intance IP (eth0), some others on the EIP|ENI address (eth1). See issue https://github.com/kubernetes-incubator/kube-aws/issues/923
It would be of course much better if we could hit the EPI/ENI all the time as they don't change even if the underlying EC2 intance goes down.
If specifying the Instance IP (eth0) in the Prometheus Operator ServiceMonitor, and the EC2 intance goes down, one would have to update the ServiceMonitor.

Another idea woud be to use the DNS entries of etcd, but those are not currently supported for EndPoints objects in Kubernetes.

# Step 2: verify

Go to the Prometheus UI on :9090/config and check that you have an etcd job entry:
```
- job_name: monitoring/etcd-k8s/0
  scrape_interval: 30s
  scrape_timeout: 10s
  ...
```

On the :9090/targets page:
 * You should see "etcd" with the UP state. If not, check the Error column for more information.
 * If no "etcd" targets are even shown on this page, prometheus isn't attempting to scrape it.

# Step 3: Grafana dashboard

## Find a dashboard you like

Try to load this dashboard:
https://grafana.com/dashboards/3070

## Save the dashboard in the configmap

As documented here, [Developing Alerts and Dashboards](developing-prometheus-rules-and-grafana-dashboards.md), the Grafana instances are stateless. The dashboards are automatically re-loaded from the ConfigMap.
So if you load a dashboard through the Grafana UI, it won't be kept unless saved in ConfigMap

Read [the document](developing-prometheus-rules-and-grafana-dashboards.md), but in summary:

### Copy your dashboard:
Once you are happy with the dashboard, export it and move it to `prometheus-operator/contrib/kube-prometheus/assets/grafana/` (ending in "-dashboard.json")

### Regenetate the grafana dashboard manifest:
`hack/scripts/generate-dashboards-configmap.sh > manifests/grafana/grafana-dashboards.yaml`

### Reload the manifest in Kubernetes:
` kubectl -n monitoring replace -f manifests/grafana/grafana-dashboards.yaml`

After a few minutes your dasboard will be available permanently to all Grafana instances
