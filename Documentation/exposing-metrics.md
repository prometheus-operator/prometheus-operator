<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

# Exposing Metrics

There are a number of [applications](https://prometheus.io/docs/instrumenting/exporters/#directly-instrumented-software) that are natively instrumented with Prometheus metrics. Those applications simply expose the metrics through an HTTP server.

The Prometheus developers and the community are maintaining [client libraries](https://prometheus.io/docs/instrumenting/clientlibs/#client-libraries) for various languages. If you want to monitor your own applications and instrument them natively, chances are there is already a client library for your language.

Not all software is natively instrumented with Prometheus metrics, but still record metrics in some other form. For these kinds of applications there are so called [exporters](https://prometheus.io/docs/instrumenting/exporters/#third-party-exporters).

Exporters can generally be divided into two categories:

* Instance exporters: These expose metrics about a single instance of an application. For example the HTTP requests that a single HTTP server has exporters served. These exporters are deployed as a [side-car](http://blog.kubernetes.io/2015/06/the-distributed-system-toolkit-patterns.html) container in the same pod as the actual instance of the respective application.  A real life example is the [`dnsmasq` metrics sidecar](https://github.com/kubernetes/dns/blob/master/docs/sidecar/README.md), which converts the proprietary metrics format communicated over the DNS protocol by `dnsmasq` to the Prometheus exposition format and exposes it on an HTTP server.

* Cluster-state exporters: These expose metrics about an entire system. For example these could be the number of 3D objects in a game, or metrics about a Kubernetes deployment. These exporters are typically deployed as a normal Kubernetes deployment, but can vary depending on the nature of the particular exporter. A real life example of this is the [`kube-state-metrics`](https://github.com/kubernetes/kube-state-metrics) exporter, which exposes metrics about the cluster state of a Kubernetes cluster.

Lastly in some cases it is not a viable option to expose metrics via an HTTP server. For example a `CronJob` may only run for a few seconds - not long enough for Prometheus to be able to scrape the HTTP endpoint. The Pushgateway was developed to be able to collect metrics for that kind of a scenario. If possible it is highly recommended not to use the Pushgateway. Read more about when to use the Pushgateway and alternative strategies here: https://prometheus.io/docs/practices/pushing/#should-i-be-using-the-pushgateway.
