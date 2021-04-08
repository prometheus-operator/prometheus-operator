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
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	kubernetesSDRoleEndpoint = "endpoints"
	kubernetesSDRolePod      = "pod"
	kubernetesSDRoleIngress  = "ingress"
)

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

type configGenerator struct {
	logger log.Logger
}

func newConfigGenerator(logger log.Logger) *configGenerator {
	cg := &configGenerator{
		logger: logger,
	}
	return cg
}

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}

func stringMapToMapSlice(m map[string]string) yaml.MapSlice {
	res := yaml.MapSlice{}
	ks := make([]string, 0)

	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	for _, k := range ks {
		res = append(res, yaml.MapItem{Key: k, Value: m[k]})
	}

	return res
}

func addSafeTLStoYaml(cfg yaml.MapSlice, namespace string, tls v1.SafeTLSConfig) yaml.MapSlice {
	pathForSelector := func(sel v1.SecretOrConfigMap) string {
		return path.Join(tlsAssetsDir, assets.TLSAssetKeyFromSelector(namespace, sel).String())
	}
	tlsConfig := yaml.MapSlice{
		{Key: "insecure_skip_verify", Value: tls.InsecureSkipVerify},
	}
	if tls.CA.Secret != nil || tls.CA.ConfigMap != nil {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "ca_file", Value: pathForSelector(tls.CA)})
	}
	if tls.Cert.Secret != nil || tls.Cert.ConfigMap != nil {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "cert_file", Value: pathForSelector(tls.Cert)})
	}
	if tls.KeySecret != nil {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "key_file", Value: pathForSelector(v1.SecretOrConfigMap{Secret: tls.KeySecret})})
	}
	if tls.ServerName != "" {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "server_name", Value: tls.ServerName})
	}
	cfg = append(cfg, yaml.MapItem{Key: "tls_config", Value: tlsConfig})
	return cfg
}

func addTLStoYaml(cfg yaml.MapSlice, namespace string, tls *v1.TLSConfig) yaml.MapSlice {
	if tls != nil {
		tlsConfig := addSafeTLStoYaml(yaml.MapSlice{}, namespace, tls.SafeTLSConfig)[0].Value.(yaml.MapSlice)
		if tls.CAFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "ca_file", Value: tls.CAFile})
		}
		if tls.CertFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "cert_file", Value: tls.CertFile})
		}
		if tls.KeyFile != "" {
			tlsConfig = append(tlsConfig, yaml.MapItem{Key: "key_file", Value: tls.KeyFile})
		}
		cfg = append(cfg, yaml.MapItem{Key: "tls_config", Value: tlsConfig})
	}
	return cfg
}

func buildExternalLabels(p *v1.Prometheus) yaml.MapSlice {
	m := map[string]string{}

	// Use "prometheus" external label name by default if field is missing.
	// Do not add external label if field is set to empty string.
	prometheusExternalLabelName := "prometheus"
	if p.Spec.PrometheusExternalLabelName != nil {
		if *p.Spec.PrometheusExternalLabelName != "" {
			prometheusExternalLabelName = *p.Spec.PrometheusExternalLabelName
		} else {
			prometheusExternalLabelName = ""
		}
	}

	// Use defaultReplicaExternalLabelName constant by default if field is missing.
	// Do not add external label if field is set to empty string.
	replicaExternalLabelName := defaultReplicaExternalLabelName
	if p.Spec.ReplicaExternalLabelName != nil {
		if *p.Spec.ReplicaExternalLabelName != "" {
			replicaExternalLabelName = *p.Spec.ReplicaExternalLabelName
		} else {
			replicaExternalLabelName = ""
		}
	}

	if prometheusExternalLabelName != "" {
		m[prometheusExternalLabelName] = fmt.Sprintf("%s/%s", p.Namespace, p.Name)
	}

	if replicaExternalLabelName != "" {
		m[replicaExternalLabelName] = "$(POD_NAME)"
	}

	for n, v := range p.Spec.ExternalLabels {
		m[n] = v
	}
	return stringMapToMapSlice(m)
}

func (cg *configGenerator) generateConfig(
	p *v1.Prometheus,
	sMons map[string]*v1.ServiceMonitor,
	pMons map[string]*v1.PodMonitor,
	probes map[string]*v1.Probe,
	basicAuthSecrets map[string]assets.BasicAuthCredentials,
	bearerTokens map[string]assets.BearerToken,
	additionalScrapeConfigs []byte,
	additionalAlertRelabelConfigs []byte,
	additionalAlertManagerConfigs []byte,
	ruleConfigMapNames []string,
) ([]byte, error) {
	versionStr := p.Spec.Version
	if versionStr == "" {
		versionStr = operator.DefaultPrometheusVersion
	}

	version, err := semver.ParseTolerant(versionStr)
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

	globalItems := yaml.MapSlice{
		{Key: "evaluation_interval", Value: evaluationInterval},
		{Key: "scrape_interval", Value: scrapeInterval},
		{Key: "external_labels", Value: buildExternalLabels(p)},
	}

	if p.Spec.ScrapeTimeout != "" {
		globalItems = append(globalItems, yaml.MapItem{
			Key: "scrape_timeout", Value: p.Spec.ScrapeTimeout,
		})
	}

	if version.GTE(semver.MustParse("2.16.0")) && p.Spec.QueryLogFile != "" {
		globalItems = append(globalItems, yaml.MapItem{
			Key: "query_log_file", Value: p.Spec.QueryLogFile,
		})
	}

	cfg = append(cfg, yaml.MapItem{Key: "global", Value: globalItems})

	ruleFilePaths := []string{}
	for _, name := range ruleConfigMapNames {
		ruleFilePaths = append(ruleFilePaths, rulesDir+"/"+name+"/*.yaml")
	}
	cfg = append(cfg, yaml.MapItem{
		Key:   "rule_files",
		Value: ruleFilePaths,
	})

	sMonIdentifiers := make([]string, len(sMons))
	i := 0
	for k := range sMons {
		sMonIdentifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(sMonIdentifiers)

	pMonIdentifiers := make([]string, len(pMons))
	i = 0
	for k := range pMons {
		pMonIdentifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(pMonIdentifiers)

	probeIdentifiers := make([]string, len(probes))
	i = 0
	for k := range probes {
		probeIdentifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(probeIdentifiers)

	apiserverConfig := p.Spec.APIServerConfig
	shards := int32(1)
	if p.Spec.Shards != nil && *p.Spec.Shards > 1 {
		shards = *p.Spec.Shards
	}

	var scrapeConfigs []yaml.MapSlice
	for _, identifier := range sMonIdentifiers {
		for i, ep := range sMons[identifier].Spec.Endpoints {
			scrapeConfigs = append(scrapeConfigs,
				cg.generateServiceMonitorConfig(
					version,
					sMons[identifier],
					ep, i,
					apiserverConfig,
					basicAuthSecrets,
					bearerTokens,
					p.Spec.OverrideHonorLabels,
					p.Spec.OverrideHonorTimestamps,
					p.Spec.IgnoreNamespaceSelectors,
					p.Spec.EnforcedNamespaceLabel,
					p.Spec.EnforcedSampleLimit,
					p.Spec.EnforcedTargetLimit,
					shards,
				),
			)
		}
	}
	for _, identifier := range pMonIdentifiers {
		for i, ep := range pMons[identifier].Spec.PodMetricsEndpoints {
			scrapeConfigs = append(scrapeConfigs,
				cg.generatePodMonitorConfig(
					version,
					pMons[identifier], ep, i,
					apiserverConfig,
					basicAuthSecrets,
					bearerTokens,
					p.Spec.OverrideHonorLabels,
					p.Spec.OverrideHonorTimestamps,
					p.Spec.IgnoreNamespaceSelectors,
					p.Spec.EnforcedNamespaceLabel,
					p.Spec.EnforcedSampleLimit,
					p.Spec.EnforcedTargetLimit,
					shards,
				),
			)
		}
	}

	for _, identifier := range probeIdentifiers {
		scrapeConfigs = append(scrapeConfigs,
			cg.generateProbeConfig(
				version,
				probes[identifier],
				apiserverConfig,
				basicAuthSecrets,
				bearerTokens,
				p.Spec.OverrideHonorLabels,
				p.Spec.OverrideHonorTimestamps,
				p.Spec.IgnoreNamespaceSelectors,
				p.Spec.EnforcedNamespaceLabel,
			),
		)
	}

	var alertmanagerConfigs []yaml.MapSlice
	if p.Spec.Alerting != nil {
		for _, am := range p.Spec.Alerting.Alertmanagers {
			alertmanagerConfigs = append(alertmanagerConfigs, cg.generateAlertmanagerConfig(version, am, apiserverConfig, basicAuthSecrets))
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

	// Use defaultReplicaExternalLabelName constant by default if field is missing.
	// Do not add external label if field is set to empty string.
	replicaExternalLabelName := defaultReplicaExternalLabelName
	if p.Spec.ReplicaExternalLabelName != nil {
		if *p.Spec.ReplicaExternalLabelName != "" {
			replicaExternalLabelName = *p.Spec.ReplicaExternalLabelName
		} else {
			replicaExternalLabelName = ""
		}
	}

	// action 'labeldrop' is not supported <= v1.4.1
	if replicaExternalLabelName != "" && version.GT(semver.MustParse("1.4.1")) {
		// Drop replica label, to make alerts from multiple Prometheus replicas alike
		alertRelabelConfigs = append(alertRelabelConfigs, yaml.MapSlice{
			{Key: "action", Value: "labeldrop"},
			{Key: "regex", Value: regexp.QuoteMeta(replicaExternalLabelName)},
		})
	}

	var additionalAlertRelabelConfigsYaml []yaml.MapSlice
	err = yaml.Unmarshal([]byte(additionalAlertRelabelConfigs), &additionalAlertRelabelConfigsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling additional alerting relabel configs failed")
	}

	cfg = append(cfg, yaml.MapItem{
		Key: "alerting",
		Value: yaml.MapSlice{
			{
				Key:   "alert_relabel_configs",
				Value: append(alertRelabelConfigs, additionalAlertRelabelConfigsYaml...),
			},
			{
				Key:   "alertmanagers",
				Value: alertmanagerConfigs,
			},
		},
	})

	if len(p.Spec.RemoteWrite) > 0 {
		cfg = append(cfg, cg.generateRemoteWriteConfig(version, p, basicAuthSecrets))
	}

	if len(p.Spec.RemoteRead) > 0 {
		cfg = append(cfg, cg.generateRemoteReadConfig(version, p, basicAuthSecrets))
	}

	return yaml.Marshal(cfg)
}

// honorLabels determines the value of honor_labels.
// if overrideHonorLabels is true and user tries to set the
// value to true, we want to set honor_labels to false.
func honorLabels(userHonorLabels, overrideHonorLabels bool) bool {
	if userHonorLabels && overrideHonorLabels {
		return false
	}
	return userHonorLabels
}

// honorTimestamps adds option to enforce honor_timestamps option in scrape_config.
// We want to disable honoring timestamps when user specified it or when global
// override is set. For backwards compatibility with prometheus <2.9.0 we don't
// set honor_timestamps when that option wasn't specified anywhere
func honorTimestamps(cfg yaml.MapSlice, userHonorTimestamps *bool, overrideHonorTimestamps bool) yaml.MapSlice {
	// Ensuring backwards compatibility by checking if user set any option
	if userHonorTimestamps == nil && !overrideHonorTimestamps {
		return cfg
	}

	honor := false
	if userHonorTimestamps != nil {
		honor = *userHonorTimestamps
	}

	return append(cfg, yaml.MapItem{Key: "honor_timestamps", Value: honor && !overrideHonorTimestamps})
}

func initRelabelings() []yaml.MapSlice {
	// Relabel prometheus job name into a meta label
	return []yaml.MapSlice{
		{
			{Key: "source_labels", Value: []string{"job"}},
			{Key: "target_label", Value: "__tmp_prometheus_job_name"},
		},
	}
}

func (cg *configGenerator) generatePodMonitorConfig(
	version semver.Version,
	m *v1.PodMonitor,
	ep v1.PodMetricsEndpoint,
	i int, apiserverConfig *v1.APIServerConfig,
	basicAuthSecrets map[string]assets.BasicAuthCredentials,
	bearerTokens map[string]assets.BearerToken,
	ignoreHonorLabels bool,
	overrideHonorTimestamps bool,
	ignoreNamespaceSelectors bool,
	enforcedNamespaceLabel string,
	enforcedSampleLimit *uint64,
	enforcedTargetLimit *uint64,
	shards int32,
) yaml.MapSlice {
	hl := honorLabels(ep.HonorLabels, ignoreHonorLabels)
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i),
		},
		{
			Key:   "honor_labels",
			Value: hl,
		},
	}
	if version.Major == 2 && version.Minor >= 9 {
		cfg = honorTimestamps(cfg, ep.HonorTimestamps, overrideHonorTimestamps)
	}

	selectedNamespaces := getNamespacesFromNamespaceSelector(&m.Spec.NamespaceSelector, m.Namespace, ignoreNamespaceSelectors)
	cfg = append(cfg, cg.generateK8SSDConfig(selectedNamespaces, apiserverConfig, basicAuthSecrets, kubernetesSDRolePod))

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

	if ep.TLSConfig != nil {
		cfg = addSafeTLStoYaml(cfg, m.Namespace, ep.TLSConfig.SafeTLSConfig)
	}

	if ep.BearerTokenSecret.Name != "" {
		if s, ok := bearerTokens[fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	if ep.BasicAuth != nil {
		if s, ok := basicAuthSecrets[fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{
				Key: "basic_auth", Value: yaml.MapSlice{
					{Key: "username", Value: s.Username},
					{Key: "password", Value: s.Password},
				},
			})
		}
	}

	relabelings := initRelabelings()

	var labelKeys []string
	// Filter targets by pods selected by the monitor.
	// Exact label matches.
	for k := range m.Spec.Selector.MatchLabels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)

	for _, k := range labelKeys {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(k)}},
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
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: strings.Join(exp.Values, "|")},
			})
		case metav1.LabelSelectorOpNotIn:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "drop"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: strings.Join(exp.Values, "|")},
			})
		case metav1.LabelSelectorOpExists:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: "true"},
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "drop"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: "true"},
			})
		}
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_name"}},
			{Key: "regex", Value: ep.Port},
		})
	} else if ep.TargetPort != nil { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		level.Warn(cg.logger).Log("msg", "PodMonitor 'targetPort' is deprecated, use 'port' instead.",
			"podMonitor", m.Name)
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if ep.TargetPort.StrVal != "" {
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_name"}},
				{Key: "regex", Value: ep.TargetPort.String()},
			})
		} else if ep.TargetPort.IntVal != 0 { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_number"}},
				{Key: "regex", Value: ep.TargetPort.String()},
			})
		}
	}

	// Relabel namespace and pod and service labels into proper labels.
	relabelings = append(relabelings, []yaml.MapSlice{
		{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
			{Key: "target_label", Value: "namespace"},
		},
		{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_name"}},
			{Key: "target_label", Value: "container"},
		},
		{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_name"}},
			{Key: "target_label", Value: "pod"},
		},
	}...)

	// Relabel targetLabels from Pod onto target.
	for _, l := range m.Spec.PodTargetLabels {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(l)}},
			{Key: "target_label", Value: sanitizeLabelName(l)},
			{Key: "regex", Value: "(.+)"},
			{Key: "replacement", Value: "${1}"},
		})
	}

	// By default, generate a safe job name from the PodMonitor. We also keep
	// this around if a jobLabel is set in case the targets don't actually have a
	// value for it. A single pod may potentially have multiple metrics
	// endpoints, therefore the endpoints labels is filled with the ports name or
	// as a fallback the port number.

	relabelings = append(relabelings, yaml.MapSlice{
		{Key: "target_label", Value: "job"},
		{Key: "replacement", Value: fmt.Sprintf("%s/%s", m.GetNamespace(), m.GetName())},
	})
	if m.Spec.JobLabel != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(m.Spec.JobLabel)}},
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
	} else if ep.TargetPort != nil && ep.TargetPort.String() != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "endpoint"},
			{Key: "replacement", Value: ep.TargetPort.String()}, //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		})
	}

	if ep.RelabelConfigs != nil {
		for _, c := range ep.RelabelConfigs {
			relabelings = append(relabelings, generateRelabelConfig(c))
		}
	}
	// Because of security risks, whenever enforcedNamespaceLabel is set, we want to append it to the
	// relabel_configs as the last relabeling, to ensure it overrides any other relabelings.
	relabelings = enforceNamespaceLabel(relabelings, m.Namespace, enforcedNamespaceLabel)

	relabelings = generateAddressShardingRelabelingRules(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	if m.Spec.SampleLimit > 0 || enforcedSampleLimit != nil {
		cfg = append(cfg, yaml.MapItem{Key: "sample_limit", Value: getLimit(m.Spec.SampleLimit, enforcedSampleLimit)})
	}

	if version.Major == 2 && version.Minor >= 21 &&
		(m.Spec.TargetLimit > 0 || enforcedTargetLimit != nil) {
		cfg = append(cfg, yaml.MapItem{Key: "target_limit", Value: getLimit(m.Spec.TargetLimit, enforcedTargetLimit)})
	}

	if ep.MetricRelabelConfigs != nil {
		var metricRelabelings []yaml.MapSlice
		for _, c := range ep.MetricRelabelConfigs {
			if c.TargetLabel != "" && enforcedNamespaceLabel != "" && c.TargetLabel == enforcedNamespaceLabel {
				continue
			}
			relabeling := generateRelabelConfig(c)

			metricRelabelings = append(metricRelabelings, relabeling)
		}
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: metricRelabelings})
	}

	return cfg
}

func (cg *configGenerator) generateProbeConfig(
	version semver.Version,
	m *v1.Probe,
	apiserverConfig *v1.APIServerConfig,
	basicAuthSecrets map[string]assets.BasicAuthCredentials,
	bearerTokens map[string]assets.BearerToken,
	ignoreHonorLabels bool,
	overrideHonorTimestamps bool,
	ignoreNamespaceSelectors bool,
	enforcedNamespaceLabel string) yaml.MapSlice {

	jobName := fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: jobName,
		},
	}

	hTs := true
	cfg = honorTimestamps(cfg, &hTs, overrideHonorTimestamps)

	path := "/probe"
	if m.Spec.ProberSpec.Path != "" {
		path = m.Spec.ProberSpec.Path
	}
	cfg = append(cfg, yaml.MapItem{Key: "metrics_path", Value: path})

	if m.Spec.Interval != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_interval", Value: m.Spec.Interval})
	}
	if m.Spec.ScrapeTimeout != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_timeout", Value: m.Spec.ScrapeTimeout})
	}
	if m.Spec.ProberSpec.Scheme != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: m.Spec.ProberSpec.Scheme})
	}

	cfg = append(cfg, yaml.MapItem{Key: "params", Value: yaml.MapSlice{
		{Key: "module", Value: []string{m.Spec.Module}},
	}})

	relabelings := initRelabelings()

	if m.Spec.JobName != "" {
		relabelings = append(relabelings, []yaml.MapSlice{
			{
				{Key: "target_label", Value: "job"},
				{Key: "replacement", Value: m.Spec.JobName},
			},
		}...)
	}
	// Generate static_config section.
	if m.Spec.Targets.StaticConfig != nil {
		staticConfig := yaml.MapSlice{
			{Key: "targets", Value: m.Spec.Targets.StaticConfig.Targets},
		}

		if m.Spec.Targets.StaticConfig.Labels != nil {
			if _, ok := m.Spec.Targets.StaticConfig.Labels["namespace"]; !ok {
				m.Spec.Targets.StaticConfig.Labels["namespace"] = m.Namespace
			}
		} else {
			m.Spec.Targets.StaticConfig.Labels = map[string]string{"namespace": m.Namespace}
		}

		staticConfig = append(staticConfig, yaml.MapSlice{
			{Key: "labels", Value: m.Spec.Targets.StaticConfig.Labels},
		}...)

		cfg = append(cfg, yaml.MapItem{
			Key:   "static_configs",
			Value: []yaml.MapSlice{staticConfig},
		})

		// Relabelings for prober.
		relabelings = append(relabelings, []yaml.MapSlice{
			{
				{Key: "source_labels", Value: []string{"__address__"}},
				{Key: "target_label", Value: "__param_target"},
			},
			{
				{Key: "source_labels", Value: []string{"__param_target"}},
				{Key: "target_label", Value: "instance"},
			},
			{
				{Key: "target_label", Value: "__address__"},
				{Key: "replacement", Value: m.Spec.ProberSpec.URL},
			},
		}...)

		// Add configured relabelings.
		if m.Spec.Targets.StaticConfig.RelabelConfigs != nil {
			for _, r := range m.Spec.Targets.StaticConfig.RelabelConfigs {
				relabelings = append(relabelings, generateRelabelConfig(r))
			}
		}

		cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: enforceNamespaceLabel(relabelings, m.Namespace, enforcedNamespaceLabel)})
	}

	// Generate kubernetes_sd_config section for ingress resources.
	if m.Spec.Targets.StaticConfig == nil {
		labelKeys := make([]string, 0, len(m.Spec.Targets.Ingress.Selector.MatchLabels))

		// Filter targets by ingresses selected by the monitor.
		// Exact label matches.
		for k := range m.Spec.Targets.Ingress.Selector.MatchLabels {
			labelKeys = append(labelKeys, k)
		}
		sort.Strings(labelKeys)

		for _, k := range labelKeys {
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(k)}},
				{Key: "regex", Value: m.Spec.Targets.Ingress.Selector.MatchLabels[k]},
			})
		}

		// Set based label matching. We have to map the valid relations
		// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
		for _, exp := range m.Spec.Targets.Ingress.Selector.MatchExpressions {
			switch exp.Operator {
			case metav1.LabelSelectorOpIn:
				relabelings = append(relabelings, yaml.MapSlice{
					{Key: "action", Value: "keep"},
					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(exp.Key)}},
					{Key: "regex", Value: strings.Join(exp.Values, "|")},
				})
			case metav1.LabelSelectorOpNotIn:
				relabelings = append(relabelings, yaml.MapSlice{
					{Key: "action", Value: "drop"},
					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(exp.Key)}},
					{Key: "regex", Value: strings.Join(exp.Values, "|")},
				})
			case metav1.LabelSelectorOpExists:
				relabelings = append(relabelings, yaml.MapSlice{
					{Key: "action", Value: "keep"},
					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
					{Key: "regex", Value: "true"},
				})
			case metav1.LabelSelectorOpDoesNotExist:
				relabelings = append(relabelings, yaml.MapSlice{
					{Key: "action", Value: "drop"},
					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
					{Key: "regex", Value: "true"},
				})
			}
		}

		selectedNamespaces := getNamespacesFromNamespaceSelector(&m.Spec.Targets.Ingress.NamespaceSelector, m.Namespace, ignoreNamespaceSelectors)
		cfg = append(cfg, cg.generateK8SSDConfig(selectedNamespaces, apiserverConfig, basicAuthSecrets, kubernetesSDRoleIngress))

		// Relabelings for ingress SD.
		relabelings = append(relabelings, []yaml.MapSlice{
			{
				{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_scheme", "__address__", "__meta_kubernetes_ingress_path"}},
				{Key: "separator", Value: ";"},
				{Key: "regex", Value: "(.+);(.+);(.+)"},
				{Key: "target_label", Value: "__param_target"},
				{Key: "replacement", Value: "${1}://${2}${3}"},
				{Key: "action", Value: "replace"},
			},
			{
				{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
				{Key: "target_label", Value: "namespace"},
			},
			{
				{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_name"}},
				{Key: "target_label", Value: "ingress"},
			},
		}...)

		// Relabelings for prober.
		relabelings = append(relabelings, []yaml.MapSlice{
			{
				{Key: "source_labels", Value: []string{"__param_target"}},
				{Key: "target_label", Value: "instance"},
			},
			{
				{Key: "target_label", Value: "__address__"},
				{Key: "replacement", Value: m.Spec.ProberSpec.URL},
			},
		}...)

		// Add configured relabelings.
		if m.Spec.Targets.Ingress.RelabelConfigs != nil {
			for _, r := range m.Spec.Targets.Ingress.RelabelConfigs {
				relabelings = append(relabelings, generateRelabelConfig(r))
			}
		}

		relabelings = enforceNamespaceLabel(relabelings, m.Namespace, enforcedNamespaceLabel)
		cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	}

	if m.Spec.TLSConfig != nil {
		cfg = addSafeTLStoYaml(cfg, m.Namespace, m.Spec.TLSConfig.SafeTLSConfig)
	}

	if m.Spec.BearerTokenSecret.Name != "" {
		pnKey := fmt.Sprintf("probe/%s/%s", m.GetNamespace(), m.GetName())
		if s, ok := bearerTokens[pnKey]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	if m.Spec.BasicAuth != nil {
		if s, ok := basicAuthSecrets[fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)]; ok {
			cfg = append(cfg, yaml.MapItem{
				Key: "basic_auth", Value: yaml.MapSlice{
					{Key: "username", Value: s.Username},
					{Key: "password", Value: s.Password},
				},
			})
		}
	}

	return cfg
}

func (cg *configGenerator) generateServiceMonitorConfig(
	version semver.Version,
	m *v1.ServiceMonitor,
	ep v1.Endpoint,
	i int,
	apiserverConfig *v1.APIServerConfig,
	basicAuthSecrets map[string]assets.BasicAuthCredentials,
	bearerTokens map[string]assets.BearerToken,
	overrideHonorLabels bool,
	overrideHonorTimestamps bool,
	ignoreNamespaceSelectors bool,
	enforcedNamespaceLabel string,
	enforcedSampleLimit *uint64,
	enforcedTargetLimit *uint64,
	shards int32,
) yaml.MapSlice {
	hl := honorLabels(ep.HonorLabels, overrideHonorLabels)
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i),
		},
		{
			Key:   "honor_labels",
			Value: hl,
		},
	}
	if version.Major == 2 && version.Minor >= 9 {
		cfg = honorTimestamps(cfg, ep.HonorTimestamps, overrideHonorTimestamps)
	}

	selectedNamespaces := getNamespacesFromNamespaceSelector(&m.Spec.NamespaceSelector, m.Namespace, ignoreNamespaceSelectors)
	cfg = append(cfg, cg.generateK8SSDConfig(selectedNamespaces, apiserverConfig, basicAuthSecrets, kubernetesSDRoleEndpoint))

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

	cfg = addTLStoYaml(cfg, m.Namespace, ep.TLSConfig)

	if ep.BearerTokenFile != "" {
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: ep.BearerTokenFile})
	}

	if ep.BearerTokenSecret.Name != "" {
		if s, ok := bearerTokens[fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	if ep.BasicAuth != nil {
		if s, ok := basicAuthSecrets[fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{
				Key: "basic_auth", Value: yaml.MapSlice{
					{Key: "username", Value: s.Username},
					{Key: "password", Value: s.Password},
				},
			})
		}
	}

	relabelings := initRelabelings()

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
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: "true"},
			})
		case metav1.LabelSelectorOpDoesNotExist:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "drop"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: "true"},
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
	} else if ep.TargetPort != nil {
		if ep.TargetPort.StrVal != "" {
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
	}

	// Relabel namespace and pod and service labels into proper labels.
	relabelings = append(relabelings, []yaml.MapSlice{
		{ // Relabel node labels for pre v2.3 meta labels
			{Key: "source_labels", Value: []string{"__meta_kubernetes_endpoint_address_target_kind", "__meta_kubernetes_endpoint_address_target_name"}},
			{Key: "separator", Value: ";"},
			{Key: "regex", Value: "Node;(.*)"},
			{Key: "replacement", Value: "${1}"},
			{Key: "target_label", Value: "node"},
		},
		{ // Relabel pod labels for >=v2.3 meta labels
			{Key: "source_labels", Value: []string{"__meta_kubernetes_endpoint_address_target_kind", "__meta_kubernetes_endpoint_address_target_name"}},
			{Key: "separator", Value: ";"},
			{Key: "regex", Value: "Pod;(.*)"},
			{Key: "replacement", Value: "${1}"},
			{Key: "target_label", Value: "pod"},
		},
		{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_namespace"}},
			{Key: "target_label", Value: "namespace"},
		},
		{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_name"}},
			{Key: "target_label", Value: "service"},
		},
		{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_name"}},
			{Key: "target_label", Value: "pod"},
		},
		{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_name"}},
			{Key: "target_label", Value: "container"},
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

	for _, l := range m.Spec.PodTargetLabels {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(l)}},
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
	} else if ep.TargetPort != nil && ep.TargetPort.String() != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "endpoint"},
			{Key: "replacement", Value: ep.TargetPort.String()},
		})
	}

	if ep.RelabelConfigs != nil {
		for _, c := range ep.RelabelConfigs {
			relabelings = append(relabelings, generateRelabelConfig(c))
		}
	}
	// Because of security risks, whenever enforcedNamespaceLabel is set, we want to append it to the
	// relabel_configs as the last relabeling, to ensure it overrides any other relabelings.
	relabelings = enforceNamespaceLabel(relabelings, m.Namespace, enforcedNamespaceLabel)

	relabelings = generateAddressShardingRelabelingRules(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	if m.Spec.SampleLimit > 0 || enforcedSampleLimit != nil {
		cfg = append(cfg, yaml.MapItem{Key: "sample_limit", Value: getLimit(m.Spec.SampleLimit, enforcedSampleLimit)})
	}

	if version.Major == 2 && version.Minor >= 21 &&
		(m.Spec.TargetLimit > 0 || enforcedTargetLimit != nil) {
		cfg = append(cfg, yaml.MapItem{Key: "target_limit", Value: getLimit(m.Spec.TargetLimit, enforcedTargetLimit)})
	}

	if ep.MetricRelabelConfigs != nil {
		var metricRelabelings []yaml.MapSlice
		for _, c := range ep.MetricRelabelConfigs {
			if c.TargetLabel != "" && enforcedNamespaceLabel != "" && c.TargetLabel == enforcedNamespaceLabel {
				continue
			}
			relabeling := generateRelabelConfig(c)

			metricRelabelings = append(metricRelabelings, relabeling)
		}
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: metricRelabelings})
	}

	return cfg
}

func getLimit(user uint64, enforced *uint64) uint64 {
	if enforced != nil {
		if user < *enforced && user != 0 || *enforced == 0 {
			return user
		}
		return *enforced
	}
	return user
}

func generateAddressShardingRelabelingRules(relabelings []yaml.MapSlice, shards int32) []yaml.MapSlice {
	return append(relabelings, yaml.MapSlice{
		{Key: "source_labels", Value: []string{"__address__"}},
		{Key: "target_label", Value: "__tmp_hash"},
		{Key: "modulus", Value: shards},
		{Key: "action", Value: "hashmod"},
	}, yaml.MapSlice{
		{Key: "source_labels", Value: []string{"__tmp_hash"}},
		{Key: "regex", Value: "$(SHARD)"},
		{Key: "action", Value: "keep"},
	})
}

func enforceNamespaceLabel(relabelings []yaml.MapSlice, namespace, enforcedNamespaceLabel string) []yaml.MapSlice {
	if enforcedNamespaceLabel == "" {
		return relabelings
	}
	return append(relabelings, yaml.MapSlice{
		{Key: "target_label", Value: enforcedNamespaceLabel},
		{Key: "replacement", Value: namespace}})
}

func generateRelabelConfig(c *v1.RelabelConfig) yaml.MapSlice {
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

	return relabeling
}

// getNamespacesFromNamespaceSelector gets a list of namespaces to select based on
// the given namespace selector, the given default namespace, and whether to ignore namespace selectors
func getNamespacesFromNamespaceSelector(nsel *v1.NamespaceSelector, namespace string, ignoreNamespaceSelectors bool) []string {
	if ignoreNamespaceSelectors {
		return []string{namespace}
	} else if nsel.Any {
		return []string{}
	} else if len(nsel.MatchNames) == 0 {
		return []string{namespace}
	}
	return nsel.MatchNames
}

func (cg *configGenerator) generateK8SSDConfig(namespaces []string, apiserverConfig *v1.APIServerConfig, basicAuthSecrets map[string]assets.BasicAuthCredentials, role string) yaml.MapItem {
	k8sSDConfig := yaml.MapSlice{
		{
			Key:   "role",
			Value: role,
		},
	}

	if len(namespaces) != 0 {
		k8sSDConfig = append(k8sSDConfig, yaml.MapItem{
			Key: "namespaces",
			Value: yaml.MapSlice{
				{
					Key:   "names",
					Value: namespaces,
				},
			},
		})
	}

	if apiserverConfig != nil {
		k8sSDConfig = append(k8sSDConfig, yaml.MapItem{
			Key: "api_server", Value: apiserverConfig.Host,
		})

		if apiserverConfig.BasicAuth != nil && basicAuthSecrets != nil {
			if s, ok := basicAuthSecrets["apiserver"]; ok {
				k8sSDConfig = append(k8sSDConfig, yaml.MapItem{
					Key: "basic_auth", Value: yaml.MapSlice{
						{Key: "username", Value: s.Username},
						{Key: "password", Value: s.Password},
					},
				})
			}
		}

		if apiserverConfig.BearerToken != "" {
			k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token", Value: apiserverConfig.BearerToken})
		}

		if apiserverConfig.BearerTokenFile != "" {
			k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token_file", Value: apiserverConfig.BearerTokenFile})
		}

		// TODO: If we want to support secret refs for k8s service discovery tls
		// config as well, make sure to path the right namespace here.
		k8sSDConfig = addTLStoYaml(k8sSDConfig, "", apiserverConfig.TLSConfig)
	}

	return yaml.MapItem{
		Key: "kubernetes_sd_configs",
		Value: []yaml.MapSlice{
			k8sSDConfig,
		},
	}
}

func (cg *configGenerator) generateAlertmanagerConfig(version semver.Version, am v1.AlertmanagerEndpoints, apiserverConfig *v1.APIServerConfig, basicAuthSecrets map[string]assets.BasicAuthCredentials) yaml.MapSlice {
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

	if am.Timeout != nil {
		cfg = append(cfg, yaml.MapItem{Key: "timeout", Value: am.Timeout})
	}

	// TODO: If we want to support secret refs for alertmanager config tls
	// config as well, make sure to path the right namespace here.
	cfg = addTLStoYaml(cfg, "", am.TLSConfig)

	cfg = append(cfg, cg.generateK8SSDConfig([]string{am.Namespace}, apiserverConfig, basicAuthSecrets, kubernetesSDRoleEndpoint))

	if am.BearerTokenFile != "" {
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: am.BearerTokenFile})
	}

	if version.Major > 2 || (version.Major == 2 && version.Minor >= 11) {
		if am.APIVersion == "v1" || am.APIVersion == "v2" {
			cfg = append(cfg, yaml.MapItem{Key: "api_version", Value: am.APIVersion})
		}
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
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_number"}},
			{Key: "regex", Value: am.Port.String()},
		})
	}

	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	return cfg
}

func (cg *configGenerator) generateRemoteReadConfig(version semver.Version, p *v1.Prometheus, basicAuthSecrets map[string]assets.BasicAuthCredentials) yaml.MapItem {

	cfgs := []yaml.MapSlice{}

	for i, spec := range p.Spec.RemoteRead {
		//defaults
		if spec.RemoteTimeout == "" {
			spec.RemoteTimeout = "30s"
		}

		cfg := yaml.MapSlice{
			{Key: "url", Value: spec.URL},
			{Key: "remote_timeout", Value: spec.RemoteTimeout},
		}

		if spec.Name != "" && version.GTE(semver.MustParse("2.15.0")) {
			cfg = append(cfg, yaml.MapItem{Key: "name", Value: spec.Name})
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
						{Key: "username", Value: s.Username},
						{Key: "password", Value: s.Password},
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

		cfg = addTLStoYaml(cfg, p.ObjectMeta.Namespace, spec.TLSConfig)

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

func (cg *configGenerator) generateRemoteWriteConfig(version semver.Version, p *v1.Prometheus, basicAuthSecrets map[string]assets.BasicAuthCredentials) yaml.MapItem {

	cfgs := []yaml.MapSlice{}

	for i, spec := range p.Spec.RemoteWrite {
		//defaults
		if spec.RemoteTimeout == "" {
			spec.RemoteTimeout = "30s"
		}

		cfg := yaml.MapSlice{
			{Key: "url", Value: spec.URL},
			{Key: "remote_timeout", Value: spec.RemoteTimeout},
		}

		if len(spec.Headers) > 0 && version.GTE(semver.MustParse("2.25.0")) {
			cfg = append(cfg, yaml.MapItem{Key: "headers", Value: stringMapToMapSlice(spec.Headers)})
		}

		if spec.Name != "" && version.GTE(semver.MustParse("2.15.0")) {
			cfg = append(cfg, yaml.MapItem{Key: "name", Value: spec.Name})
		}

		if spec.WriteRelabelConfigs != nil {
			relabelings := []yaml.MapSlice{}
			for _, c := range spec.WriteRelabelConfigs {
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
				relabelings = append(relabelings, relabeling)
			}

			cfg = append(cfg, yaml.MapItem{Key: "write_relabel_configs", Value: relabelings})

		}

		if spec.BasicAuth != nil {
			if s, ok := basicAuthSecrets[fmt.Sprintf("remoteWrite/%d", i)]; ok {
				cfg = append(cfg, yaml.MapItem{
					Key: "basic_auth", Value: yaml.MapSlice{
						{Key: "username", Value: s.Username},
						{Key: "password", Value: s.Password},
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

		cfg = addTLStoYaml(cfg, p.ObjectMeta.Namespace, spec.TLSConfig)

		if spec.ProxyURL != "" {
			cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: spec.ProxyURL})
		}

		if spec.QueueConfig != nil {
			queueConfig := yaml.MapSlice{}

			if spec.QueueConfig.Capacity != int(0) {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "capacity", Value: spec.QueueConfig.Capacity})
			}

			if version.GTE(semver.MustParse("2.6.0")) {
				if spec.QueueConfig.MinShards != int(0) {
					queueConfig = append(queueConfig, yaml.MapItem{Key: "min_shards", Value: spec.QueueConfig.MinShards})
				}
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

		if spec.MetadataConfig != nil && version.GTE(semver.MustParse("2.23.0")) {
			metadataConfig := yaml.MapSlice{}
			metadataConfig = append(metadataConfig, yaml.MapItem{Key: "send", Value: spec.MetadataConfig.Send})
			if spec.MetadataConfig.SendInterval != "" {
				metadataConfig = append(metadataConfig, yaml.MapItem{Key: "send_interval", Value: spec.MetadataConfig.SendInterval})
			}
			cfg = append(cfg, yaml.MapItem{Key: "metadata_config", Value: metadataConfig})
		}

		cfgs = append(cfgs, cfg)
	}

	return yaml.MapItem{
		Key:   "remote_write",
		Value: cfgs,
	}
}
