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
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
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

func generateConfig(p *v1alpha1.Prometheus, mons map[string]*v1alpha1.ServiceMonitor, ruleConfigMaps int, basicAuthSecrets map[string]BasicAuthCredentials) ([]byte, error) {
	cfg := yaml.MapSlice{}

	cfg = append(cfg, yaml.MapItem{
		Key: "global",
		Value: yaml.MapSlice{
			{Key: "evaluation_interval", Value: "30s"},
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
			scrapeConfigs = append(scrapeConfigs, generateServiceMonitorConfig(mons[identifier], ep, i, basicAuthSecrets))
		}
	}
	var alertmanagerConfigs []yaml.MapSlice
	for _, am := range p.Spec.Alerting.Alertmanagers {
		alertmanagerConfigs = append(alertmanagerConfigs, generateAlertmanagerConfig(am))
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

func generateServiceMonitorConfig(m *v1alpha1.ServiceMonitor, ep v1alpha1.Endpoint, i int, basicAuthSecrets map[string]BasicAuthCredentials) yaml.MapSlice {
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("%s/%s/%d", m.Namespace, m.Name, i),
		},
		{
			Key:   "honor_labels",
			Value: ep.HonorLabels,
		},
		{
			Key: "kubernetes_sd_configs",
			Value: []yaml.MapSlice{
				yaml.MapSlice{
					{Key: "role", Value: "endpoints"},
				},
			},
		},
	}

	if ep.Interval != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_interval", Value: ep.Interval})
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
			{Key: "source_labels", Value: []string{"__meta_kubernetes_container_port_name"}},
			{Key: "regex", Value: ep.TargetPort.String()},
		})
	} else if ep.TargetPort.IntVal != 0 {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_container_port_number"}},
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
	if ep.Address != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "__address__"},
			{Key: "replacement", Value: ep.Address},
		})
	}
	if ep.Path != "" {
		if u, err := url.Parse(ep.Path); err == nil {
			path, sourceLabels := parsePath(u.Path)
			if len(sourceLabels) == 0 {
				cfg = append(cfg, yaml.MapItem{Key: "metrics_path", Value: path})
			} else {
				relabelings = append(relabelings, yaml.MapSlice{
					{Key: "source_labels", Value: sourceLabels},
					{Key: "regex", Value: strings.Repeat(";(.+)", len(sourceLabels))[1:]},
					{Key: "target_label", Value: "__metrics_path__"},
					{Key: "replacement", Value: path},
				})
			}
			for k, v := range u.Query() {
				path, sourceLabels := parsePath(v[0])
				if len(sourceLabels) == 0 {
					relabelings = append(relabelings, yaml.MapSlice{
						{Key: "target_label", Value: "__param_" + k},
						{Key: "replacement", Value: path},
					})
				} else {
					relabelings = append(relabelings, yaml.MapSlice{
						{Key: "source_labels", Value: sourceLabels},
						{Key: "regex", Value: strings.Repeat(";(.+)", len(sourceLabels))[1:]},
						{Key: "target_label", Value: "__param_" + k},
						{Key: "replacement", Value: path},
					})
				}
			}
		}
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

	return cfg
}

func parsePath(path string) (string, []string) {
	if strings.HasSuffix(path, "__meta_kubernetes_") {
		return "$1", []string{path}
	}

	m := make(map[string]int) // meta_label => regex_pos
	var r bytes.Buffer        // resulting metrics_path

	meta := []rune("${__meta_kubernetes_")
	meta_len := len(meta)

	s := make([]rune, 0, len(path))
	for _, c := range path {
		s = append(s, c)
	}

	i := 0
	for i < len(s)-meta_len-1 {
		if eq(s[i:i+meta_len], meta) {
			start := i
			end := i + meta_len + 1
			for end < len(s) {
				if s[end] == '}' {
					break
				} else {
					end += 1
				}
			}
			key := string(s[start+2 : end])
			pos, ok := m[key]
			if !ok {
				pos = len(m)
				m[key] = pos
			}

			r.WriteString("${")
			r.WriteString(strconv.Itoa(pos + 1))
			r.WriteRune('}')

			i = end + 1
		} else {
			r.WriteRune(s[i])
			i += 1
		}
	}
	for i < len(s) {
		r.WriteRune(s[i])
		i += 1
	}
	sourceLabels := make([]string, len(m))
	for k, v := range m {
		sourceLabels[v] = k
	}
	return r.String(), sourceLabels
}

func eq(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if b[i] != x {
			return false
		}
	}
	return true
}

func generateAlertmanagerConfig(am v1alpha1.AlertmanagerEndpoints) yaml.MapSlice {
	if am.Scheme == "" {
		am.Scheme = "http"
	}

	cfg := yaml.MapSlice{
		{
			Key: "kubernetes_sd_configs",
			Value: []yaml.MapSlice{
				yaml.MapSlice{
					{Key: "role", Value: "endpoints"},
				},
			},
		},
		{Key: "scheme", Value: am.Scheme},
	}

	var relabelings []yaml.MapSlice

	relabelings = append(relabelings, yaml.MapSlice{
		{Key: "action", Value: "keep"},
		{Key: "source_labels", Value: []string{"__meta_kubernetes_service_name"}},
		{Key: "regex", Value: am.Name},
	})
	relabelings = append(relabelings, yaml.MapSlice{
		{Key: "action", Value: "keep"},
		{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
		{Key: "regex", Value: am.Namespace},
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

	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	return cfg
}
