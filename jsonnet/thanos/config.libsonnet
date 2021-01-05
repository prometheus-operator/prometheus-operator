local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

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
                'prometheus': 'self',
            },
            commonLabels:{
             'prometheus': 'self',
             'app': 'prometheus'
            },
            queryLabels: {'app':'thanos-query'},
            sidecarLabels: {'app':'thanos-query'},
            rulerLabels: {'app':'thanos-ruler'},
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
			"apiVersion": "monitoring.coreos.com/v1",
        	"kind": "Prometheus",
        	"metadata": {
        		"labels": po.prometheusLabels,
        		"name": "self",
        		"namespace": "%s" % $._config.namespace
        	},
        	"spec": {
        		"replicas": 2,
        		"serviceMonitorSelector": {
        			"matchLabels": {
        				"app": "prometheus"
        			}
        		},
        		"ruleSelector": {
        			"matchLabels": {
        				"role": "prometheus-rulefiles",
        				"prometheus": "k8s"
        			}
        		},
        		"thanos": {
        			"version": "%s" % $._config.versions.thanos
        		}
        	}
		},
        clusterRole:
          local clusterRole = k.rbac.v1.clusterRole;
          local policyRule = clusterRole.rulesType;
          local monitoringRule = policyRule.new() +
                                 policyRule.withApiGroups(['']) +
                                 policyRule.withResources([
                                       'nodes',
                                       'nodes/metrics',
                                       'services',
                                       'endpoints',
                                       'pods',
                                       ]) +
                                       policyRule.withVerbs(['get','list','watch']);
           local configmapRule =     policyRule.new()+
                                policyRule.withApiGroups(['']) +
                                policyRule.withResources([
                                     'configmaps',
                                     ]) +
                                     policyRule.withVerbs(['get'])
                                 ;
          local noResourceRule= policyRule.new()+policyRule.withNonResourceUrls(['/metrics'])  ;
          local rules = [monitoringRule,configmapRule,noResourceRule];

          clusterRole.new() +
          clusterRole.mixin.metadata.withName(po.prometheusName) +
          clusterRole.withRules(rules),
        clusterRoleBinding:
          local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;
          clusterRoleBinding.new() +
          clusterRoleBinding.mixin.metadata.withName(po.prometheusName) +
          clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
          clusterRoleBinding.mixin.roleRef.withName(po.prometheusName) +
          clusterRoleBinding.mixin.roleRef.withKind('ClusterRole') +
          clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'default', namespace: po.namespace }]),
        service:
          local service = k.core.v1.service;
          local servicePort = k.core.v1.service.mixin.spec.portsType;

          local poServicePort = servicePort.newNamed('web', 9090, 'web');

          service.new(po.prometheusName, po.prometheusLabels, [poServicePort]) +
          service.mixin.metadata.withLabels(po.commonLabels) +
          service.mixin.metadata.withNamespace(po.namespace) ,
        serviceMonitor:
          {
            apiVersion: 'monitoring.coreos.com/v1',
            kind: 'ServiceMonitor',
            metadata: {
              name: po.prometheusName,
              namespace: po.namespace,
              labels: po.prometheusLabels,
          },
          spec: {
            endpoints: [
            {
              port: 'web',
              interval: '30s',
            },
            ],
            selector: {
              matchLabels:  {
                  "app":"prometheus"
              },
            },
           },
          },
        queryDeployment:
          local deployment = k.apps.v1.deployment;
          local container = k.apps.v1.deployment.mixin.spec.template.spec.containersType;
          local containerPort = container.portsType;
          local name = 'thanos-query';
          local queryContainer =
           container.new(name, po.image + ':' + po.version) +

           container.withPorts([containerPort.newNamed(10902, 'http'),containerPort.newNamed(10901, 'grpc')]) +
           container.withArgs([
                'query',
                '--log.level=debug',
                '--query.replica-label=prometheus_replica',
                '--query.replica-label=thanos_ruler_replica',
                '--store=dnssrv+_grpc._tcp.thanos-sidecar.default.svc.cluster.local',
                '--store=dnssrv+_grpc._tcp.thanos-ruler.default.svc.cluster.local'
           ]) +
          container.mixin.resources.withRequests({ cpu: '100m', memory: '100Mi' }) +
          container.mixin.resources.withLimits({ cpu: '200m', memory: '200Mi' });

          deployment.new(name, 1, queryContainer, {app:"thanos-query"}) +
          deployment.mixin.metadata.withNamespace(po.namespace) +
          deployment.mixin.metadata.withLabels({app:"thanos-query"}) +
          deployment.mixin.spec.selector.withMatchLabels({app:"thanos-query"}) ,
        queryService:
          local service = k.core.v1.service;
          local servicePort = k.core.v1.service.mixin.spec.portsType;

          local poServicePort = servicePort.newNamed('http', 9090, 'http');
          service.new(po.thanosQueryName, po.queryLabels, [poServicePort]) +
          service.mixin.metadata.withLabels(po.queryLabels) +
          service.mixin.metadata.withNamespace(po.namespace) ,
        sidecarService:
          local service = k.core.v1.service;
          local servicePort = k.core.v1.service.mixin.spec.portsType;

          local poServicePort = servicePort.newNamed('grpc', 10901, 'grpc');
          service.new(po.thanosSidecarName, po.commonLabels, [poServicePort]) +
          service.mixin.metadata.withLabels(po.sidecarLabels) +
          service.mixin.metadata.withNamespace(po.namespace) +
          service.mixin.spec.withSelector(po.prometheusLabels) +
          service.mixin.spec.withClusterIp('None'),
        thanosRuler:
            {
                "apiVersion": "monitoring.coreos.com/v1",
                "kind": "ThanosRuler",
                "metadata": {
                    "labels":po.rulerLabels,
                    "name": po.thanosRulerName,
                    "namespace": po.namespace
                },
                "spec": {
                    "image": po.image + ':' + po.version,
                    "ruleSelector":{
                        "matchLabels":{
                            "role":"thanos-example"
                        }
                    },
                    "queryEndpoints":["dnssrv+_http._tcp.thanos-query.default.svc.cluster.local"]
                }
            },
        prometheusRule:
            {
                "apiVersion": "monitoring.coreos.com/v1",
                "kind": "PrometheusRule",
                "metadata": {
                    "labels": {
                        "prometheus": "example-alert",
                        "role": "thanos-example"
                    },
                    "creationTimestamp": null,
                    "name": "prometheus-example-alerts",
                    "namespace": "default"
                },
                "spec": {
                    "groups": [{
                        "name": "./example-alert.rules",
                        "rules": [{
                            "alert": "ExampleAlert",
                            "expr": "vector(1)"

                        }]
                    }]
                }
            },
        thanosRulerService:
          local service = k.core.v1.service;
          local servicePort = k.core.v1.service.mixin.spec.portsType;
          local grpcServicePort = servicePort.newNamed('grpc', 10901, 'grpc');
          local httpServicePort = servicePort.newNamed('http', 10902, 'web');

          service.new(po.thanosRulerName, po.rulerLabels, [grpcServicePort,httpServicePort]) +
          service.mixin.metadata.withLabels(po.rulerLabels) +
          service.mixin.metadata.withNamespace(po.namespace) +
          service.mixin.spec.withSelector(po.rulerLabels) +
          service.mixin.spec.withClusterIp('None'),
	}
}