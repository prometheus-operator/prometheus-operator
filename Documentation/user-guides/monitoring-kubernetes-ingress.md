<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.39.0, Prometheus Operator requires use of Kubernetes v1.16.x and up.<br><br>
This documentation is for an alpha feature.
</div>

# Monitoring Kubernetes Ingress with Ambassador

[Ambassador](https://getambassador.io/) is a popular open-source API gateway for Kubernetes. Built on [Envoy Proxy](https://envoyproxy.io), Ambassador natively exposes statistics that give you better insight to what is happening at the edge of your Kubernetes cluster. In this guide we will:

* Create a simple Kubernetes application
* Deploy Ambassador as your Kubernetes ingress controller
* Use the Prometheus Operator to deploy and manage our Prometheus instance
* Examine some metrics related to the performance of Kubernetes application at the edge

## Prerequisites
* A Kubernetes Cluster
* The [Kubernetes command line tool](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Deploy and Expose an Application on Kubernetes

First, we need an application running on Kubernetes for our users to access. You can deploy any application you would like but, for simplicity, we will use a [sample application](https://github.com/datawire/tour) provided by the Ambassador team.

We can quickly deploy this application using `kubectl`:

```sh
kubectl apply -f https://raw.githubusercontent.com/datawire/tour/refs/heads/master/k8s/tour.yaml
```

Check the application's status and wait for it to start running:

```console
$ kubectl get po --selector=app=tour

NAME                    READY   STATUS    RESTARTS   AGE
tour-6df995489d-lc9hs   2/2     Running   0          1m
```

Now that we have an application running in Kubernetes, we need to expose it to the outside world. We will do this with Ambassador to take advantage of Envoy's robust metrics generation.

1. Deploy Ambassador to your cluster with `kubectl`:

   ```yaml
    ---
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        service: ambassador-admin
      name: ambassador-admin
    spec:
      type: NodePort
      ports:
      - name: ambassador-admin
        port: 8877
        targetPort: 8877
      selector:
        service: ambassador
    ---
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRole
    metadata:
      name: ambassador
    rules:
    - apiGroups: [""]
      resources: [ "endpoints", "namespaces", "secrets", "services" ]
      verbs: ["get", "list", "watch"]
    - apiGroups: [ "getambassador.io" ]
      resources: [ "*" ]
      verbs: ["get", "list", "watch"]
    - apiGroups: [ "getambassador.io" ]
      resources: [ "mappings/status" ]
      verbs: ["update"]
    - apiGroups: [ "apiextensions.k8s.io" ]
      resources: [ "customresourcedefinitions" ]
      verbs: ["get", "list", "watch"]
    - apiGroups: [ "networking.internal.knative.dev" ]
      resources: [ "clusteringresses", "ingresses" ]
      verbs: ["get", "list", "watch"]
    - apiGroups: [ "networking.internal.knative.dev" ]
      resources: [ "ingresses/status", "clusteringresses/status" ]
      verbs: ["update"]
    - apiGroups: [ "extensions", "networking.k8s.io" ]
      resources: [ "ingresses" ]
      verbs: ["get", "list", "watch"]
    - apiGroups: [ "extensions", "networking.k8s.io" ]
      resources: [ "ingresses/status" ]
      verbs: ["update"]
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: ambassador
    ---
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRoleBinding
    metadata:
      name: ambassador
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: ambassador
    subjects:
    - kind: ServiceAccount
      name: ambassador
      namespace: default
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: ambassador
    spec:
      replicas: 3
      selector:
        matchLabels:
          service: ambassador
      template:
        metadata:
          annotations:
            consul.hashicorp.com/connect-inject: 'false'
            sidecar.istio.io/inject: 'false'
          labels:
            service: ambassador
        spec:
          affinity:
            podAntiAffinity:
              preferredDuringSchedulingIgnoredDuringExecution:
              - podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      service: ambassador
                  topologyKey: kubernetes.io/hostname
                weight: 100
          containers:
          - env:
            - name: AMBASSADOR_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            image: quay.io/datawire/ambassador:1.1.1
            livenessProbe:
              httpGet:
                path: /ambassador/v0/check_alive
                port: 8877
              initialDelaySeconds: 30
              periodSeconds: 3
            name: ambassador
            ports:
            - containerPort: 8080
              name: http
            - containerPort: 8443
              name: https
            - containerPort: 8877
              name: admin
            readinessProbe:
              httpGet:
                path: /ambassador/v0/check_ready
                port: 8877
              initialDelaySeconds: 30
              periodSeconds: 3
            resources:
              limits:
                cpu: 1
                memory: 400Mi
              requests:
                cpu: 200m
                memory: 100Mi
            volumeMounts:
            - mountPath: /tmp/ambassador-pod-info
              name: ambassador-pod-info
          restartPolicy: Always
          securityContext:
            runAsUser: 8888
          serviceAccountName: ambassador
          volumes:
          - downwardAPI:
              items:
              - fieldRef:
                  fieldPath: metadata.labels
                path: labels
            name: ambassador-pod-info

   ```

   ```sh
   kubectl apply -f ambassador-rbac.yaml
   ```

   Ambassador is now running in your cluster and is ready to start routing traffic to your application.

2. Expose Ambassador to the internet.

   ```yaml
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: ambassador
     labels:
       app.kubernetes.io/component: ambassador-service
   spec:
     type: LoadBalancer
     externalTrafficPolicy: Local
     ports:
      - port: 80
        targetPort: 8080
     selector:
       service: ambassador
   ```

   ```sh
   kubectl apply -f ambassador-service.yaml
   ```

   This will create a `LoadBalancer` service in Kubernetes which will automatically create a cloud load balancer if you are running in cloud-managed Kubernetes.

   Kubernetes will automatically assign the load balancer's IP address as the `EXTERNAL_IP` of the service. You can view this with `kubectl`:

   ```console
   $ kubectl get svc ambassador

   NAME         TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)       AGE
   ambassador   LoadBalancer   10.96.110.170   34.233.165.XXX   80:33241/TCP  87m
   ```

   **Note:** If you are running in a different Kubernetes environment that does not automatically create a load balancer (like minikube), you can still access Ambassador using the `NodePort` of the service (33241 in this example).

3. Route traffic to your application

   You configure Ambassador to expose your application using [annotations](https://github.com/datawire/ambassador-docs/blob/rel/1.1.1/reference/core/annotations.md#annotations) on the Kubernetes service of the application like the one below.

   ```yaml
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: tour
     annotations:
       getambassador.io/config: |
         ---
         apiVersion: ambassador/v1
         kind: Mapping
         name: tour-ui_mapping
         prefix: /
         service: tour:5000
   spec:
     ports:
     - name: ui
       port: 5000
       targetPort: 5000
     selector:
       app: tour
   ```

   The example application above was deployed with a pre-configured annotation to expose it. You can can now access this application using the `EXTERNAL_IP` above and going to http://{AMBASSADOR_EXTERNAL_IP}/ from a web-browser.

You now have an application running in Kubernetes and exposed to the internet.

## Deploy Prometheus

Now that we have an application running and exposed by Ambassador, we need to configure Prometheus to scrape the metrics from Ambassador. The Prometheus Operator gives us a way to deploy and manage Prometheus deployments using Kubernetes-style resources

The Prometheus Operator creates Kubernetes Custom Resource Definitions (CRDs) so we can manage our Prometheus deployment using Kubernetes-style declarative YAML manifests. To deploy the Prometheus Operator, you can clone the [repository](https://github.com/prometheus-operator/prometheus-operator) and follow the instructions in the README. You can also just create it with `kubectl`:

```sh
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/master/bundle.yaml
```

Once the Prometheus operator is running, we need to create a Prometheus instance. The Prometheus Operator manages Prometheus deployments with the `Prometheus` CRD. To create a Prometheus instance and Kubernetes service, copy the following YAML to a file called `prometheus.yaml` and deploy it with `kubectl`:

```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  annotations:
    getambassador.io/config: |
       ---
       apiVersion: ambassador/v1
       kind: Mapping
       name: prometheus_mapping
       prefix: /prometheus/
       service: prometheus:9090
spec:
  type: ClusterIP
  ports:
  - name: web
    port: 9090
    protocol: TCP
    targetPort: 9090
  selector:
    prometheus: prometheus
---
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  externalUrl: http://{AMBASSADOR_EXTERNAL_IP}/prometheus
  routePrefix: /prometheus
  ruleSelector:
    matchLabels:
      app.kubernetes.io/name: prometheus-operator
  serviceAccountName: prometheus-operator
  serviceMonitorSelector:
    matchLabels:
      app: ambassador
  resources:
    requests:
      memory: 400Mi
```

**Note:** The `externalUrl` and `routePrefix` fields allows for you to route requests to Prometheus through Ambassador on the `/prometheus` path. Replace `{AMBASSADOR_EXTERNAL_IP}` with the value from above.

```sh
kubectl apply -f prometheus.yaml
```

We now have Prometheus running in the cluster and exposed through Ambassador. View the Prometheus UI by going to http://{AMBASSADOR_EXTERNAL_IP}/prometheus/graph from a web browser.

Finally, we need tell Prometheus where to scrape metrics from. The Prometheus Operator easily manages this using a `ServiceMonitor` CRD. To tell Prometheus to scrape metrics from Ambassador's `/metrics` endpoint, we will use the Ambassador admin service and port `ambassador-admin`(8877). Copy the following YAML to a file called `ambassador-monitor.yaml` and apply it with `kubectl`:

```yaml
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ambassador-monitor
  namespace: monitoring
  labels:
    app: ambassador
spec:
  namespaceSelector:
    matchNames:
    - default
  selector:
    matchLabels:
      service: ambassador-admin
  endpoints:
  - port: ambassador-admin
```

```sh
kubectl apply -f ambassador-monitor.yaml
```

Prometheus will now be configured to collect metrics from the `ambassador-admin` Kubernetes service with the internal address: `http://ambassador-admin:8877/metrics`.

## Examining Ingress Metrics

If you go to `http://{AMBASSADOR_EXTERNAL_IP}/prometheus/targets` from a web browser you will now see `ambassador-monitor` as a target for Prometheus to scrape metrics from Ambassador. Clicking on the drop down menu at `http://{AMBASSADOR_EXTERNAL_IP}/prometheus/graph`, you can see the various ingress-related metrics output by Envoy.

Envoy's metrics data model is remarkably similar to that of Prometheus and uses the same three kinds of statistics (`Counters`, `Gauges`, and `Histograms`). This allows for Envoy to export dynamic and data-rich statistics that are immediately useable by Prometheus's analytical functions.

**Notable Metrics:**

| Metric Category           | Notable Metrics                                                                                                                         | Description                                                                                                                                                                                                                                                       |
|---------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| envoy_http_downstream_rq  | envoy_http_downstream_rq_http1_total, envoy_http_downstream_rq_http1_total, envoy_http_downstream_rq_total, envoy_http_downstream_rq_xx | Statistics regarding traffic from the internet, to each Ambassador instance. Tracking this will give you insight into how each pod is performing for various requests.                                                                                            |
| envoy_cluster_upstream_rq | envoy_cluster_upstream_rq, envoy_cluster_upstream_rq_xx, envoy_cluster_upstream_rq_total, envoy_cluster_upstream_rq_retry               | Statistics regarding traffic from Envoy to each upstream service. Tracking this will give you insight to how the request is performing after reaching Ambassador. It will help you pinpoint whether failures are happening in Ambassador or the upstream service. |

Envoy collects many more statistics including some regarding rate limiting, circuit breaking, and distributed tracing. See the [Envoy's documentation](https://envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats) for more information on the metrics envoy collects.
