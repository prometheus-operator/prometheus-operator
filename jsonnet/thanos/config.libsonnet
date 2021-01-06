local service(name, namespace, labels, selector, ports) = {
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: name,
    namespace: namespace,
    labels: labels,
  },
  spec: {
    ports+: ports,
    selector: selector,
  },
};

{
  _config+:: {
    namespace: 'default',
    prometheusName: 'prometheus-self',
    thanosQueryName: 'thanos-query',
    thanosRulerName: 'thanos-ruler',
    thanosSidecarName: 'thanos-sidecar',
    versions+:: {
      thanos: 'v0.11.2',
    },

    imageRepos+:: {
      thanos: 'quay.io/thanos/thanos',
    },
    labels+:: {
      prometheusLabels: {
        prometheus: 'self',
      },
      commonLabels: {
        prometheus: 'self',
        app: 'prometheus',
      },
      queryLabels: { app: 'thanos-query' },
      sidecarLabels: { app: 'thanos-sidecar' },
      rulerLabels: { app: 'thanos-ruler' },
    },

  },
  thanos+:: {
    local po = self,
    namespace:: $._config.namespace,
    image:: $._config.imageRepos.thanos,
    version:: $._config.versions.thanos,
    commonLabels:: $._config.labels.commonLabels,
    prometheusLabels:: $._config.labels.prometheusLabels,
    queryLabels:: $._config.labels.queryLabels,
    sidecarLabels:: $._config.labels.sidecarLabels,
    rulerLabels:: $._config.labels.rulerLabels,
    prometheusName:: $._config.prometheusName,
    thanosQueryName:: $._config.thanosQueryName,
    thanosRulerName:: $._config.thanosRulerName,
    thanosSidecarName:: $._config.thanosSidecarName,
    prometheus+:: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'Prometheus',
      metadata: {
        labels: po.prometheusLabels,
        name: 'self',
        namespace: '%s' % $._config.namespace,
      },
      spec: {
        replicas: 2,
        serviceMonitorSelector: {
          matchLabels: {
            app: 'prometheus',
          },
        },
        ruleSelector: {
          matchLabels: {
            role: 'prometheus-rulefiles',
            prometheus: 'k8s',
          },
        },
        thanos: {
          version: '%s' % $._config.versions.thanos,
        },
      },
    },
    clusterRole: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      metadata: {
        name: po.prometheusName,
      },
      rules: [
        {
          apiGroups: [''],
          resources: ['nodes', 'nodes/metrics', 'services', 'endpoints', 'pods'],
          verbs: ['get', 'list', 'watch'],
        },
        {
          apiGroups: [''],
          resources: ['configmaps'],
          verbs: ['get'],
        },
        {
          nonResourceURLs: ['/metrics'],
          verbs: ['get'],
        },
      ],
    },
    clusterRoleBinding: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRoleBinding',
      metadata: {
        name: po.prometheusName,
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: po.prometheusName,
      },
      subjects: [{
        kind: 'ServiceAccount',
        name: 'default',
        namespace: po.namespace,
      }],
    },
    service: service(
      po.prometheusName,
      po.namespace,
      po.commonLabels,
      po.prometheusLabels,
      [{ name: 'web', port: 9090, targetPort: 'web' }]
    ),
    serviceMonitor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: po.prometheusName,
          namespace: po.namespace,
          labels: po.commonLabels,
        },
        spec: {
          endpoints: [
            {
              port: 'web',
              interval: '30s',
            },
          ],
          selector: {
            matchLabels: {
              app: 'prometheus',
            },
          },
        },
      },
    queryDeployment: {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: po.thanosQueryName,
        namespace: po.namespace,
        labels: po.queryLabels,
      },
      spec: {
        replicas: 1,
        selector: { matchLabels: po.queryLabels },
        template: {
          metadata: { labels: po.queryLabels },
          spec: {
            containers: [
              {
                name: 'thanos-query',
                image: po.image + ':' + po.version,
                args: [
                  'query',
                  '--log.level=debug',
                  '--query.replica-label=prometheus_replica',
                  '--query.replica-label=thanos_ruler_replica',
                  '--store=dnssrv+_grpc._tcp.thanos-sidecar.default.svc.cluster.local',
                  '--store=dnssrv+_grpc._tcp.thanos-ruler.default.svc.cluster.local',
                ],
                ports: [
                  {
                    name: 'http',
                    containerPort: 10902,
                  },
                  {
                    name: 'grpc',
                    containerPort: 10901,
                  },
                ],
              },
            ],
          },
        },
      },
    },
    queryService: service(
      po.thanosQueryName,
      po.namespace,
      po.queryLabels,
      po.queryLabels,
      [{ name: 'http', port: 10902, targetPort: 'http' }]
    ),
    sidecarService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: po.thanosSidecarName,
        namespace: po.namespace,
        labels: po.sidecarLabels,
      },
      spec: {
        ports: [{
          name: 'grpc',
          port: 10901,
          targetPort: 'grpc',
        }],
        selector: po.prometheusLabels,
        clusterIP: 'None',
      },
    },
    thanosRuler:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ThanosRuler',
        metadata: {
          labels: po.rulerLabels,
          name: po.thanosRulerName,
          namespace: po.namespace,
        },
        spec: {
          image: po.image + ':' + po.version,
          ruleSelector: {
            matchLabels: {
              role: 'thanos-example',
            },
          },
          queryEndpoints: ['dnssrv+_http._tcp.thanos-query.default.svc.cluster.local'],
        },
      },
    prometheusRule:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'PrometheusRule',
        metadata: {
          labels: {
            prometheus: 'example-alert',
            role: 'thanos-example',
          },
          creationTimestamp: null,
          name: 'prometheus-example-alerts',
          namespace: 'default',
        },
        spec: {
          groups: [{
            name: './example-alert.rules',
            rules: [{
              alert: 'ExampleAlert',
              expr: 'vector(1)',

            }],
          }],
        },
      },
    thanosRulerService: service(
      po.thanosRulerName,
      po.namespace,
      po.rulerLabels,
      po.rulerLabels,
      [{
        name: 'grpc',
        port: 10901,
        targetPort: 'grpc',
      }, {
        name: 'http',
        port: 10902,
        targetPort: 'web',
      }]
    ),
  },
}
