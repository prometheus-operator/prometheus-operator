local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
	_config+:: {
		namespace: 'default',
		versions+:: {
			thanos: 'v0.11.2',
		},

		imageRepos+:: {
			thanos: 'quay.io/thanos/thanos',
		},
	},
	thanos+:: {
		local po = self,
		namespace:: $._config.namespace,
		image:: $._config.imageRepos.thanos,
		version:: $._config.versions.thanos,
		prometheus+:: {
			"apiVersion": "monitoring.coreos.com/v1",
        	"kind": "Prometheus",
        	"metadata": {
        		"labels": {
        			"prometheus": "self"
        		},
        		"name": "self",
        		"namespace": "%s" % $._config.namespace
        	},
        	"spec": {
        		"replicas": "2",
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
          policyRule.new() +
                                 policyRule.withApiGroups(['']) +
                                 policyRule.withResources([
                                   'nodes',
                                   'nodes/metrics',
                                   'services',
                                   'endpoints',
                                   'pods',
                                 ]) +
                                 policyRule.withVerbs(['get','list','watch']) +
                                 policyRule.withApiGroups(['']) +
                                 policyRule.withResources([
                                 'configmaps',
                                  ]) +
                                 policyRule.withVerbs(['get']) +
                                 policyRule.withNonResourceUrls(['/metrics']) +
                                 policyRule.withVerbs(['get']),
          clusterRole.new(),
	}
}