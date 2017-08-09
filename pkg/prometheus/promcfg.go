// Copyright 2016 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
)

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}

func configMapRuleFileFolder(configMapNumber int) string {
	return fmt.Sprintf("/etc/prometheus/rules/rules-%d/", configMapNumber)
}

func stringMapToMapSlice(m map[string]string) yaml.MapSlice {
	res := yaml.MapSlice{}

	for k, v := range m {
		res = append(res, yaml.MapItem{Key: k, Value: v})
	}

	return res
}

func generateConfig(p *v1.Prometheus, mons map[string]*v1.ServiceMonitor, ruleConfigMaps int, basicAuthSecrets map[string]BasicAuthCredentials) ([]byte, error) {
	versionStr := p.Spec.Version
	if versionStr == "" {
		versionStr = DefaultVersion
	}

	version, err := semver.Parse(strings.TrimLeft(versionStr, "v"))
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	cfg := yaml.MapSlice{}

	evaluationInterval := "30s"
	if p.Spec.EvaluationInterval != "" {
		evaluationInterval = p.Spec.EvaluationInterval
	}

	cfg = append(cfg, yaml.MapItem{
		Key: "global",
		Value: yaml.MapSlice{
			{Key: "evaluation_interval", Value: evaluationInterval},
			{Key: "scrape_interval", Value: "30s"},
			{Key: "external_labels", Value: stringMapToMapSlice(p.Spec.ExternalLabels)},
		},
	})

	if ruleConfigMaps > 0 {
		configMaps := make([]string, ruleConfigMaps)
		for i := 0; i < ruleConfigMaps; i++ {
			configMaps[i] = configMapRuleFileFolder(i) + "*.rules"
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "rule_files",
			Value: configMaps,
		})
	}

	identifiers := make([]string, len(mons))
	i := 0
	for k, _ := range mons {
		identifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(identifiers)

	var scrapeConfigs []yaml.MapSlice
	for _, identifier := range identifiers {
		for i, ep := range mons[identifier].Spec.Endpoints {
			scrapeConfigs = append(scrapeConfigs, generateServiceMonitorConfig(version, mons[identifier], ep, i, basicAuthSecrets))
		}
	}
	var alertmanagerConfigs []yaml.MapSlice
	for _, am := range p.Spec.Alerting.Alertmanagers {
		alertmanagerConfigs = append(alertmanagerConfigs, generateAlertmanagerConfig(version, am))
	}

	cfg = append(cfg, yaml.MapItem{
		Key:   "scrape_configs",
		Value: scrapeConfigs,
	})

	cfg = append(cfg, yaml.MapItem{
		Key: "alerting",
		Value: yaml.MapSlice{
			{
				Key:   "alertmanagers",
				Value: alertmanagerConfigs,
			},
		},
	})

	return yaml.Marshal(cfg)
}

func generateServiceMonitorConfig(version semver.Version, m *v1.ServiceMonitor, ep v1.Endpoint, i int, basicAuthSecrets map[string]BasicAuthCredentials) yaml.MapSlice {
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("%s/%s/%d", m.Namespace, m.Name, i),
		},
		{
			Key:   "honor_labels",
			Value: ep.HonorLabels,
		},
	}

	switch version.Major {
	case 1:
		if version.Minor < 7 {
			cfg = append(cfg, k8sSDAllNamespaces())
		} else {
			cfg = append(cfg, k8sSDFromServiceMonitor(m))
		}
	case 2:
		cfg = append(cfg, k8sSDFromServiceMonitor(m))
	}

	if ep.Interval != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_interval", Value: ep.Interval})
	}
	if ep.ScrapeTimeout != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_timeout", Value: ep.ScrapeTimeout})
	}
	if ep.Path != "" {
		cfg = append(cfg, yaml.MapItem{Key: "metrics_path", Value: ep.Path})
	}
	if ep.Scheme != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: ep.Scheme})
	}
	if ep.TLSConfig != nil {
		tlsConfig := yaml.MapSlice{
			{Key: "insecure_skip_verify", Value: ep.TLSConfig.InsecureSkipVerify},
		}
		if ep.TLSConfig.CAFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "ca_file", Value: ep.TLSConfig.CAFile})
		}
		if ep.TLSConfig.CertFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "cert_file", Value: ep.TLSConfig.CertFile})
		}
		if ep.TLSConfig.KeyFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "key_file", Value: ep.TLSConfig.KeyFile})
		}
		if ep.TLSConfig.ServerName != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "server_name", Value: ep.TLSConfig.ServerName})
		}
		cfg = append(cfg, yaml.MapItem{Key: "tls_config", Value: tlsConfig})
	}
	if ep.BearerTokenFile != "" {
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: ep.BearerTokenFile})
	}

	if ep.BasicAuth != nil {
		if s, ok := basicAuthSecrets[fmt.Sprintf("%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{
				Key: "basic_auth", Value: yaml.MapSlice{
					{Key: "username", Value: s.username},
					{Key: "password", Value: s.password},
				},
			})
		}
	}

	var relabelings []yaml.MapSlice

	// Filter targets by services selected by the monitor.

	// Exact label matches.
	labelKeys := make([]string, len(m.Spec.Selector.MatchLabels))
	i = 0
	for k, _ := range m.Spec.Selector.MatchLabels {
		labelKeys[i] = k
		i++
	}
	sort.Strings(labelKeys)
	for i := range labelKeys {
		k := labelKeys[i]
		v := m.Spec.Selector.MatchLabels[k]
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(k)}},
			{Key: "regex", Value: v},
		})
	}
	// Set based label matching. We have to map the valid relations
	// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
	for _, exp := range m.Spec.Selector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: strings.Join(exp.Values, "|")},
			})
		case metav1.LabelSelectorOpNotIn:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "drop"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: strings.Join(exp.Values, "|")},
			})
		case metav1.LabelSelectorOpExists:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: ".+"},
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "drop"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: ".+"},
			})
		}
	}

	if version.Major == 1 && version.Minor < 7 {
		// Filter targets based on the namespace selection configuration.
		// By default we only discover services within the namespace of the
		// ServiceMonitor.
		// Selections allow extending this to all namespaces or to a subset
		// of them specified by label or name matching.
		//
		// Label selections are not supported yet as they require either supported
		// in the upstream SD integration or require out-of-band implementation
		// in the operator with configuration reload.
		//
		// There's no explicit nil for the selector, we decide for the default
		// case if it's all zero values.
		nsel := m.Spec.NamespaceSelector

		if !nsel.Any && len(nsel.MatchNames) == 0 {
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
				{Key: "regex", Value: m.Namespace},
			})
		} else if len(nsel.MatchNames) > 0 {
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
				{Key: "regex", Value: strings.Join(nsel.MatchNames, "|")},
			})
		}
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_endpoint_port_name"}},
			{Key: "regex", Value: ep.Port},
		})
	} else if ep.TargetPort.StrVal != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_name"}},
			{Key: "regex", Value: ep.TargetPort.String()},
		})
	} else if ep.TargetPort.IntVal != 0 {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_number"}},
			{Key: "regex", Value: ep.TargetPort.String()},
		})
	}

	// Relabel namespace and pod and service labels into proper labels.
	relabelings = append(relabelings, []yaml.MapSlice{
		yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
			{Key: "target_label", Value: "namespace"},
		},
		yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_name"}},
			{Key: "target_label", Value: "pod"},
		},
		yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_name"}},
			{Key: "target_label", Value: "service"},
		},
	}...)

	// By default, generate a safe job name from the service name.  We also keep
	// this around if a jobLabel is set in case the targets don't actually have a
	// value for it. A single service may potentially have multiple metrics
	// endpoints, therefore the endpoints labels is filled with the ports name or
	// as a fallback the port number.

	relabelings = append(relabelings, yaml.MapSlice{
		{Key: "source_labels", Value: []string{"__meta_kubernetes_service_name"}},
		{Key: "target_label", Value: "job"},
		{Key: "replacement", Value: "${1}"},
	})
	if m.Spec.JobLabel != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(m.Spec.JobLabel)}},
			{Key: "target_label", Value: "job"},
			{Key: "regex", Value: "(.+)"},
			{Key: "replacement", Value: "${1}"},
		})
	}

	if ep.Port != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "endpoint"},
			{Key: "replacement", Value: ep.Port},
		})
	} else if ep.TargetPort.String() != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "endpoint"},
			{Key: "replacement", Value: ep.TargetPort.String()},
		})
	}

	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	return cfg
}

func k8sSDFromServiceMonitor(m *v1.ServiceMonitor) yaml.MapItem {
	nsel := m.Spec.NamespaceSelector
	namespaces := []string{}
	if !nsel.Any && len(nsel.MatchNames) == 0 {
		namespaces = append(namespaces, m.Namespace)
	}
	if !nsel.Any && len(nsel.MatchNames) > 0 {
		for i := range nsel.MatchNames {
			namespaces = append(namespaces, nsel.MatchNames[i])
		}
	}

	return k8sSDWithNamespaces(namespaces)
}

func k8sSDWithNamespaces(namespaces []string) yaml.MapItem {
	return yaml.MapItem{
		Key: "kubernetes_sd_configs",
		Value: []yaml.MapSlice{
			yaml.MapSlice{
				{
					Key:   "role",
					Value: "endpoints",
				},
				{
					Key: "namespaces",
					Value: yaml.MapSlice{
						{
							Key:   "names",
							Value: namespaces,
						},
					},
				},
			},
		},
	}
}

func k8sSDAllNamespaces() yaml.MapItem {
	return yaml.MapItem{
		Key: "kubernetes_sd_configs",
		Value: []yaml.MapSlice{
			yaml.MapSlice{
				{
					Key:   "role",
					Value: "endpoints",
				},
			},
		},
	}
}

func generateAlertmanagerConfig(version semver.Version, am v1.AlertmanagerEndpoints) yaml.MapSlice {
	if am.Scheme == "" {
		am.Scheme = "http"
	}

	if am.PathPrefix == "" {
		am.PathPrefix = "/"
	}

	cfg := yaml.MapSlice{
		{Key: "path_prefix", Value: am.PathPrefix},
		{Key: "scheme", Value: am.Scheme},
	}

	switch version.Major {
	case 1:
		if version.Minor < 7 {
			cfg = append(cfg, k8sSDAllNamespaces())
		} else {
			cfg = append(cfg, k8sSDWithNamespaces([]string{am.Namespace}))
		}
	case 2:
		cfg = append(cfg, k8sSDWithNamespaces([]string{am.Namespace}))
	}

	var relabelings []yaml.MapSlice

	relabelings = append(relabelings, yaml.MapSlice{
		{Key: "action", Value: "keep"},
		{Key: "source_labels", Value: []string{"__meta_kubernetes_service_name"}},
		{Key: "regex", Value: am.Name},
	})

	if am.Port.StrVal != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_endpoint_port_name"}},
			{Key: "regex", Value: am.Port.String()},
		})
	} else if am.Port.IntVal != 0 {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_container_port_number"}},
			{Key: "regex", Value: am.Port.String()},
		})
	}

	if version.Major == 1 && version.Minor < 7 {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
			{Key: "regex", Value: am.Namespace},
		})
	}

	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	return cfg
}
