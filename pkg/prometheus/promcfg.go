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
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
)

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}

func stringMapToMapSlice(m map[string]string) yaml.MapSlice {
	res := yaml.MapSlice{}
	ks := make([]string, 0)

	for k, _ := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	for _, k := range ks {
		res = append(res, yaml.MapItem{Key: k, Value: m[k]})
	}

	return res
}

func addTLStoYaml(cfg yaml.MapSlice, tls *v1.TLSConfig) yaml.MapSlice {
	if tls != nil {
		tlsConfig := yaml.MapSlice{
			{Key: "insecure_skip_verify", Value: tls.InsecureSkipVerify},
		}
		if tls.CAFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "ca_file", Value: tls.CAFile})
		}
		if tls.CertFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "cert_file", Value: tls.CertFile})
		}
		if tls.KeyFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "key_file", Value: tls.KeyFile})
		}
		if tls.ServerName != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "server_name", Value: tls.ServerName})
		}
		cfg = append(cfg, yaml.MapItem{Key: "tls_config", Value: tlsConfig})
	}
	return cfg
}

func buildExternalLabels(p *v1.Prometheus) yaml.MapSlice {
	m := map[string]string{}

	m["prometheus"] = fmt.Sprintf("%s/%s", p.Namespace, p.Name)
	m["prometheus_replica"] = "$(POD_NAME)"

	for n, v := range p.Spec.ExternalLabels {
		m[n] = v
	}
	return stringMapToMapSlice(m)
}

func generateConfig(p *v1.Prometheus, mons map[string]*v1.ServiceMonitor, basicAuthSecrets map[string]BasicAuthCredentials, additionalScrapeConfigs []byte, additionalAlertManagerConfigs []byte) ([]byte, error) {
	versionStr := p.Spec.Version
	if versionStr == "" {
		versionStr = DefaultPrometheusVersion
	}

	version, err := semver.Parse(strings.TrimLeft(versionStr, "v"))
	if err != nil {
		return nil, errors.Wrap(err, "parse version")
	}

	cfg := yaml.MapSlice{}

	scrapeInterval := "30s"
	if p.Spec.ScrapeInterval != "" {
		scrapeInterval = p.Spec.ScrapeInterval
	}

	evaluationInterval := "30s"
	if p.Spec.EvaluationInterval != "" {
		evaluationInterval = p.Spec.EvaluationInterval
	}

	cfg = append(cfg, yaml.MapItem{
		Key: "global",
		Value: yaml.MapSlice{
			{Key: "evaluation_interval", Value: evaluationInterval},
			{Key: "scrape_interval", Value: scrapeInterval},
			{Key: "external_labels", Value: buildExternalLabels(p)},
		},
	})

	cfg = append(cfg, yaml.MapItem{
		Key:   "rule_files",
		Value: []string{"/etc/prometheus/rules/*.yaml"},
	})

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
	if p.Spec.Alerting != nil {
		for _, am := range p.Spec.Alerting.Alertmanagers {
			alertmanagerConfigs = append(alertmanagerConfigs, generateAlertmanagerConfig(version, am))
		}
	}

	var additionalScrapeConfigsYaml []yaml.MapSlice
	err = yaml.Unmarshal([]byte(additionalScrapeConfigs), &additionalScrapeConfigsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling additional scrape configs failed")
	}

	cfg = append(cfg, yaml.MapItem{
		Key:   "scrape_configs",
		Value: append(scrapeConfigs, additionalScrapeConfigsYaml...),
	})

	var additionalAlertManagerConfigsYaml []yaml.MapSlice
	err = yaml.Unmarshal([]byte(additionalAlertManagerConfigs), &additionalAlertManagerConfigsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling additional alert manager configs failed")
	}

	alertmanagerConfigs = append(alertmanagerConfigs, additionalAlertManagerConfigsYaml...)

	var alertRelabelConfigs []yaml.MapSlice

	// action 'labeldrop' is not supported <= v1.4.1
	if version.GT(semver.MustParse("1.4.1")) {
		// Drop 'prometheus_replica' label, to make alerts from two Prometheus replicas alike
		alertRelabelConfigs = append(alertRelabelConfigs, yaml.MapSlice{
			{Key: "action", Value: "labeldrop"},
			{Key: "regex", Value: "prometheus_replica"},
		})
	}

	cfg = append(cfg, yaml.MapItem{
		Key: "alerting",
		Value: yaml.MapSlice{
			{
				Key:   "alert_relabel_configs",
				Value: alertRelabelConfigs,
			},
			{
				Key:   "alertmanagers",
				Value: alertmanagerConfigs,
			},
		},
	})

	if len(p.Spec.RemoteWrite) > 0 && version.Major >= 2 {
		cfg = append(cfg, generateRemoteWriteConfig(version, p.Spec.RemoteWrite, basicAuthSecrets))
	}

	if len(p.Spec.RemoteRead) > 0 && version.Major >= 2 {
		cfg = append(cfg, generateRemoteReadConfig(version, p.Spec.RemoteRead, basicAuthSecrets))
	}

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
	if ep.ProxyURL != nil {
		cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: ep.ProxyURL})
	}
	if ep.Params != nil {
		cfg = append(cfg, yaml.MapItem{Key: "params", Value: ep.Params})
	}
	if ep.Scheme != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: ep.Scheme})
	}

	cfg = addTLStoYaml(cfg, ep.TLSConfig)

	if ep.BearerTokenFile != "" {
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: ep.BearerTokenFile})
	}

	if ep.BasicAuth != nil {
		if s, ok := basicAuthSecrets[fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
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
	var labelKeys []string
	for k := range m.Spec.Selector.MatchLabels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)

	for _, k := range labelKeys {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(k)}},
			{Key: "regex", Value: m.Spec.Selector.MatchLabels[k]},
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

	// Relabel targetLabels from Service onto target.
	for _, l := range m.Spec.TargetLabels {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(l)}},
			{Key: "target_label", Value: sanitizeLabelName(l)},
			{Key: "regex", Value: "(.+)"},
			{Key: "replacement", Value: "${1}"},
		})
	}

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

	if ep.MetricRelabelConfigs != nil {
		var metricRelabelings []yaml.MapSlice
		for _, c := range ep.MetricRelabelConfigs {
			relabeling := yaml.MapSlice{}

			if len(c.SourceLabels) > 0 {
				relabeling = append(relabeling, yaml.MapItem{Key: "source_labels", Value: c.SourceLabels})
			}

			if c.Separator != "" {
				relabeling = append(relabeling, yaml.MapItem{Key: "separator", Value: c.Separator})
			}

			if c.TargetLabel != "" {
				relabeling = append(relabeling, yaml.MapItem{Key: "target_label", Value: c.TargetLabel})
			}

			if c.Regex != "" {
				relabeling = append(relabeling, yaml.MapItem{Key: "regex", Value: c.Regex})
			}

			if c.Modulus != uint64(0) {
				relabeling = append(relabeling, yaml.MapItem{Key: "modulus", Value: c.Modulus})
			}

			if c.Replacement != "" {
				relabeling = append(relabeling, yaml.MapItem{Key: "replacement", Value: c.Replacement})
			}

			if c.Action != "" {
				relabeling = append(relabeling, yaml.MapItem{Key: "action", Value: c.Action})
			}

			metricRelabelings = append(metricRelabelings, relabeling)
		}
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: metricRelabelings})
	}

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

	cfg = addTLStoYaml(cfg, am.TLSConfig)

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

	if am.BearerTokenFile != "" {
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: am.BearerTokenFile})
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

func generateRemoteReadConfig(version semver.Version, specs []v1.RemoteReadSpec, basicAuthSecrets map[string]BasicAuthCredentials) yaml.MapItem {

	cfgs := []yaml.MapSlice{}

	for i, spec := range specs {
		//defaults
		if spec.RemoteTimeout == "" {
			spec.RemoteTimeout = "30s"
		}

		cfg := yaml.MapSlice{
			{Key: "url", Value: spec.URL},
			{Key: "remote_timeout", Value: spec.RemoteTimeout},
		}

		if len(spec.RequiredMatchers) > 0 {
			cfg = append(cfg, yaml.MapItem{Key: "required_matchers", Value: stringMapToMapSlice(spec.RequiredMatchers)})
		}

		if spec.ReadRecent {
			cfg = append(cfg, yaml.MapItem{Key: "read_recent", Value: spec.ReadRecent})
		}

		if spec.BasicAuth != nil {
			if s, ok := basicAuthSecrets[fmt.Sprintf("remoteRead/%d", i)]; ok {
				cfg = append(cfg, yaml.MapItem{
					Key: "basic_auth", Value: yaml.MapSlice{
						{Key: "username", Value: s.username},
						{Key: "password", Value: s.password},
					},
				})
			}
		}

		if spec.BearerTokenFile != "" {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: spec.BearerTokenFile})
		}

		cfg = addTLStoYaml(cfg, spec.TLSConfig)

		if spec.ProxyURL != "" {
			cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: spec.ProxyURL})
		}

		cfgs = append(cfgs, cfg)

	}

	return yaml.MapItem{
		Key:   "remote_read",
		Value: cfgs,
	}
}

func generateRemoteWriteConfig(version semver.Version, specs []v1.RemoteWriteSpec, basicAuthSecrets map[string]BasicAuthCredentials) yaml.MapItem {

	cfgs := []yaml.MapSlice{}

	for i, spec := range specs {
		//defaults
		if spec.RemoteTimeout == "" {
			spec.RemoteTimeout = "30s"
		}

		cfg := yaml.MapSlice{
			{Key: "url", Value: spec.URL},
			{Key: "remote_timeout", Value: spec.RemoteTimeout},
		}

		if spec.WriteRelabelConfigs != nil {
			relabelings := []yaml.MapSlice{}
			for _, c := range spec.WriteRelabelConfigs {
				relabeling := yaml.MapSlice{
					{Key: "source_labels", Value: c.SourceLabels},
				}

				if c.Separator != "" {
					relabeling = append(relabeling, yaml.MapItem{Key: "separator", Value: c.Separator})
				}

				if c.TargetLabel != "" {
					relabeling = append(relabeling, yaml.MapItem{Key: "target_label", Value: c.TargetLabel})
				}

				if c.Regex != "" {
					relabeling = append(relabeling, yaml.MapItem{Key: "regex", Value: c.Regex})
				}

				if c.Modulus != uint64(0) {
					relabeling = append(relabeling, yaml.MapItem{Key: "modulus", Value: c.Modulus})
				}

				if c.Replacement != "" {
					relabeling = append(relabeling, yaml.MapItem{Key: "replacement", Value: c.Replacement})
				}

				if c.Action != "" {
					relabeling = append(relabeling, yaml.MapItem{Key: "action", Value: c.Action})
				}
				relabelings = append(relabelings, relabeling)
			}

			cfg = append(cfg, yaml.MapItem{Key: "write_relabel_configs", Value: relabelings})

		}

		if spec.BasicAuth != nil {
			if s, ok := basicAuthSecrets[fmt.Sprintf("remoteWrite/%d", i)]; ok {
				cfg = append(cfg, yaml.MapItem{
					Key: "basic_auth", Value: yaml.MapSlice{
						{Key: "username", Value: s.username},
						{Key: "password", Value: s.password},
					},
				})
			}
		}

		if spec.BearerToken != "" {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: spec.BearerToken})
		}

		if spec.BearerTokenFile != "" {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: spec.BearerTokenFile})
		}

		cfg = addTLStoYaml(cfg, spec.TLSConfig)

		if spec.ProxyURL != "" {
			cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: spec.ProxyURL})
		}

		if spec.QueueConfig != nil {
			queueConfig := yaml.MapSlice{}

			if spec.QueueConfig.Capacity != int(0) {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "capacity", Value: spec.QueueConfig.Capacity})
			}

			if spec.QueueConfig.MaxShards != int(0) {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "max_shards", Value: spec.QueueConfig.MaxShards})
			}

			if spec.QueueConfig.MaxSamplesPerSend != int(0) {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "max_samples_per_send", Value: spec.QueueConfig.MaxSamplesPerSend})
			}

			if spec.QueueConfig.BatchSendDeadline != "" {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "batch_send_deadline", Value: spec.QueueConfig.BatchSendDeadline})
			}

			if spec.QueueConfig.MaxRetries != int(0) {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "max_retries", Value: spec.QueueConfig.MaxRetries})
			}

			if spec.QueueConfig.MinBackoff != "" {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "min_backoff", Value: spec.QueueConfig.MinBackoff})
			}

			if spec.QueueConfig.MaxBackoff != "" {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "max_backoff", Value: spec.QueueConfig.MaxBackoff})
			}

			cfg = append(cfg, yaml.MapItem{Key: "queue_config", Value: queueConfig})
		}

		cfgs = append(cfgs, cfg)
	}

	return yaml.MapItem{
		Key:   "remote_write",
		Value: cfgs,
	}
}
