# Monitoring external etcd
This guide will help you monitor an external etcd cluster. When the etcd cluster is not hosted inside Kubernetes.
This is often the case with Kubernetes setups. This approach has been tested with kube-aws but the same principals apply to other tools.

# Step 1 - Make the etcd certificates available to Prometheus pod
Prometheus Operator (and Prometheus) allow us to specify a tlsConfig. This is required as most likely your etcd metrics end points is secure.

## a - Create the secrets in the namespace 
Prometheus Operator allows us to mount secrets in the pod. By loading the secrets as files, they can be made available inside the Prometheus pod.

`kubectl -n monitoring create secret generic etcd-certs --from-file=CREDENTIAL_PATH/etcd-client.pem --from-file=CREDENTIAL_PATH/etcd-client-key.pem --from-file=CREDENTIAL_PATH/ca.pem`

where CREDENTIAL_PATH is the path to your etcd client credentials on your work machine. 
(Kube-aws stores them inside the credential folder).

## b - Get Prometheus Operator to load the secret
In the previous step we have named the secret 'etcd-certs'.

Edit prometheus-operator/contrib/kube-prometheus/manifests/prometheus/prometheus-k8s.yaml and add the secret under the spec of the Prometheus object manifest:

```
  secrets: 
  - etcd-certs
```

The manifest will look like that:
```
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: k8s
  labels:
    prometheus: k8s
spec:
  replicas: 2
  secrets: 
  - etcd-certs
  version: v1.7.1
```

If your Prometheus Operator is already in place, update it:

`kubectl -n monitoring replace -f contrib/kube-prometheus/manifests/prometheus/prometheus-k8s.yaml

# Step 2 - Create the Service, endpoints and ServiceMonitor

The below manifest creates a Service to expose etcd metrics (port 2379)

* Replace `IP_OF_YOUR_ETCD_NODE_[0/1/2]` with the IP addresses of your etcd nodes. If you have more than one node, add them to the same list.
* Use `#insecureSkipVerify: true` or replace `ETCD_DNS_OR_ALTERNAME_NAME` with a valid name for the certificate. 

In case you have generated the etcd certificated with kube-aws, you will need to use insecureSkipVerify as the valid certificate domain will be different for each etcd node (etcd0, etcd1, etcd2). If you only have one etcd node, you can use the value from `etcd.internalDomainName` speficied in your kube-aws `cluster.yaml`

In this example we use insecureSkipVerify: true as kube-aws default certificates are not valid against the IP. They were created for the DNS. Depending on your use case, you might want to remove this flag or set it to false. (true required for kube-aws if using default certificate generators method)

```
apiVersion: v1
kind: Service
metadata:
  name: etcd-k8s
  labels:
    k8s-app: etcd
spec:
  type: ClusterIP
  clusterIP: None
  ports:
  - name: api
    port: 2379
    protocol: TCP
---
apiVersion: v1
kind: Endpoints
metadata:
  name: etcd-k8s
  labels:
    k8s-app: etcd
subsets:
- addresses:
  - ip: IP_OF_YOUR_ETCD_NODE_0
    nodeName: etcd0
  - ip: IP_OF_YOUR_ETCD_NODE_1
    nodeName: etcd1
  - ip: IP_OF_YOUR_ETCD_NODE_2
    nodeName: etcd2
  ports:
  - name: api
    port: 2379
    protocol: TCP
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: etcd-k8s
  labels:
    k8s-app: etcd-k8s
spec:
  jobLabel: k8s-app
  endpoints:
  - port: api
    interval: 30s
    scheme: https
    tlsConfig:
      caFile: /etc/prometheus/secrets/etcd-certs/ca.pem
      certFile: /etc/prometheus/secrets/etcd-certs/etcd-client.pem
      keyFile: /etc/prometheus/secrets/etcd-certs/etcd-client-key.pem
      #use insecureSkipVerify only if you cannot use a Subject Alternative Name
      #insecureSkipVerify: true 
      serverName: ETCD_DNS_OR_ALTERNAME_NAME
  selector:
    matchLabels:
      k8s-app: etcd
  namespaceSelector:
    matchNames:
    - monitoring
```

# Step 3: Open the port 

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

# Step 4: verify

Go to the Prometheus UI on :9090/config and check that you have an etcd job entry:
```
- job_name: monitoring/etcd-k8s/0
  scrape_interval: 30s
  scrape_timeout: 10s
  ...
```

On the :9090/targets page, you should see "etcd" with the UP state. If not, check the Error column for more information.

# Step 5: Grafana dashboard

## Find a dashboard you like

Try to load this dashboard:
https://grafana.com/dashboards/3070

## Save the dashboard in the configmap

As documented here, [Developing Alerts and Dashboards](developing-alerts-and-dashboards.md), the Grafana instances are stateless. The dashboards are automatically re-loaded from the ConfigMap.
So if you load a dashboard through the Grafana UI, it won't be kept unless saved in ConfigMap

Read [the document](developing-alerts-and-dashboards.md), but in summary:

### Copy your dashboard:
Once you are happy with the dashboard, export it and move it to `prometheus-operator/contrib/kube-prometheus/assets/grafana/` (ending in "-dashboard.json")

### Regenetate the grafana dashboard manifest:
`hack/scripts/generate-dashboards-configmap.sh > manifests/grafana/grafana-dashboards.yaml`

### Reload the manifest in Kubernetes:
` kubectl -n monitoring replace -f manifests/grafana/grafana-dashboards.yaml`

After a few minutes your dasboard will be available permanently to all Grafana instances
