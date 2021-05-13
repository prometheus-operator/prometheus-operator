local defaults = {
  local defaults = self,
  name: 'prometheus-operator',
  namespace: error 'must provide namespace',
  version: error 'must provide version',
  image: error 'must provide image',
  configReloaderImage: error 'must provide configReloaderImage',
  port: 8080,
  resources: {
    limits: { cpu: '200m', memory: '200Mi' },
    requests: { cpu: '100m', memory: '100Mi' },
  },
  commonLabels:: {
    'app.kubernetes.io/name': 'prometheus-operator',
    'app.kubernetes.io/version': defaults.version,
    'app.kubernetes.io/component': 'controller',
  },
  selectorLabels:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if !std.setMember(labelName, ['app.kubernetes.io/version'])
  },
};

function(params) {
  local po = self,
  config:: defaults + params,

  // Prefixing with 0 to ensure these manifests are listed and therefore created first.
  '0alertmanagerCustomResourceDefinition': import 'alertmanager-crd.libsonnet',
  '0alertmanagerConfigCustomResourceDefinition': import 'alertmanagerconfig-crd.libsonnet',
  '0prometheusCustomResourceDefinition': import 'prometheus-crd.libsonnet',
  '0servicemonitorCustomResourceDefinition': import 'servicemonitor-crd.libsonnet',
  '0podmonitorCustomResourceDefinition': import 'podmonitor-crd.libsonnet',
  '0probeCustomResourceDefinition': import 'probe-crd.libsonnet',
  '0prometheusruleCustomResourceDefinition': import 'prometheusrule-crd.libsonnet',
  '0thanosrulerCustomResourceDefinition': import 'thanosruler-crd.libsonnet',

  clusterRoleBinding: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'ClusterRoleBinding',
    metadata: {
      name: po.config.name,
      labels: po.config.commonLabels,
    },
    roleRef: {
      apiGroup: 'rbac.authorization.k8s.io',
      kind: 'ClusterRole',
      name: po.config.name,
    },
    subjects: [{
      kind: 'ServiceAccount',
      name: po.config.name,
      namespace: po.config.namespace,
    }],
  },

  clusterRole: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'ClusterRole',
    metadata: {
      name: po.config.name,
      labels: po.config.commonLabels,
    },
    rules: [
      {
        apiGroups: ['monitoring.coreos.com'],
        resources: [
          'alertmanagers',
          'alertmanagers/finalizers',
          'alertmanagerconfigs',
          'prometheuses',
          'prometheuses/finalizers',
          'thanosrulers',
          'thanosrulers/finalizers',
          'servicemonitors',
          'podmonitors',
          'probes',
          'prometheusrules',
        ],
        verbs: ['*'],
      },
      {
        apiGroups: ['apps'],
        resources: ['statefulsets'],
        verbs: ['*'],
      },
      {
        apiGroups: [''],
        resources: ['configmaps', 'secrets'],
        verbs: ['*'],
      },
      {
        apiGroups: [''],
        resources: ['pods'],
        verbs: ['list', 'delete'],
      },
      {
        apiGroups: [''],
        resources: [
          'services',
          'services/finalizers',
          'endpoints',
        ],
        verbs: ['get', 'create', 'update', 'delete'],
      },
      {
        apiGroups: [''],
        resources: ['nodes'],
        verbs: ['list', 'watch'],
      },
      {
        apiGroups: [''],
        resources: ['namespaces'],
        verbs: ['get', 'list', 'watch'],
      },
      {
        apiGroups: ['networking.k8s.io'],
        resources: ['ingresses'],
        verbs: ['get', 'list', 'watch'],
      },
    ],
  },

  deployment:
    local container = {
      name: po.config.name,
      image: po.config.image,
      args: [
        '--kubelet-service=kube-system/kubelet',
        '--prometheus-config-reloader=' + po.config.configReloaderImage,
      ],
      ports: [{
        containerPort: po.config.port,
        name: 'http',
      }],
      resources: po.config.resources,
      securityContext: {
        allowPrivilegeEscalation: false,
      },
    };
    {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: po.config.name,
        namespace: po.config.namespace,
        labels: po.config.commonLabels,
      },
      spec: {
        replicas: 1,
        selector: { matchLabels: po.config.selectorLabels },
        template: {
          metadata: {
            labels: po.config.commonLabels,
            annotations: {
              "kubectl.kubernetes.io/default-container": container.name,
            }
          },
          spec: {
            containers: [container],
            nodeSelector: {
              'kubernetes.io/os': 'linux',
            },

            securityContext: {
              runAsNonRoot: true,
              runAsUser: 65534,
            },
            serviceAccountName: po.config.name,
          },
        },
      },
    },

  serviceAccount: {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: {
      name: po.config.name,
      namespace: po.config.namespace,
      labels: po.config.commonLabels,
    },
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: po.config.name,
      namespace: po.config.namespace,
      labels: po.config.commonLabels,
    },
    spec: {
      ports: [
        { name: 'http', targetPort: 'http', port: po.config.port },
      ],
      selector: po.config.selectorLabels,
      clusterIP: 'None',
    },
  },

  serviceMonitor: {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'ServiceMonitor',
    metadata: {
      name: 'prometheus-operator',
      namespace: po.config.namespace,
      labels: po.config.commonLabels,
    },
    spec: {
      endpoints: [
        {
          port: 'http',
          honorLabels: true,
        },
      ],
      selector: {
        matchLabels: po.config.commonLabels,
      },
    },
  },
}
