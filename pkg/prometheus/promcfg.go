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
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespace-labeler"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	kubernetesSDRoleEndpoint      = "endpoints"
	kubernetesSDRoleEndpointSlice = "endpointslice"
	kubernetesSDRolePod           = "pod"
	kubernetesSDRoleIngress       = "ingress"
)

var (
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}

// ConfigGenerator knows how to generate a Prometheus configuration which is
// compatible with a given Prometheus version.
type ConfigGenerator struct {
	logger                 log.Logger
	version                semver.Version
	notCompatible          bool
	spec                   *v1.PrometheusSpec
	endpointSliceSupported bool
}

// NewConfigGenerator creates a ConfigGenerator for the provided Prometheus resource.
func NewConfigGenerator(logger log.Logger, p *v1.Prometheus, endpointSliceSupported bool) (*ConfigGenerator, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	promVersion := operator.StringValOrDefault(p.Spec.Version, operator.DefaultPrometheusVersion)
	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Prometheus version")
	}

	logger = log.WithSuffix(logger, "version", promVersion)

	return &ConfigGenerator{
		logger:                 logger,
		version:                version,
		spec:                   p.Spec.DeepCopy(),
		endpointSliceSupported: endpointSliceSupported,
	}, nil
}

// WithKeyVals returns a new ConfigGenerator with the same characteristics as
// the current object, expect that the keyvals are appended to the existing
// logger.
func (cg *ConfigGenerator) WithKeyVals(keyvals ...interface{}) *ConfigGenerator {
	return &ConfigGenerator{
		logger:                 log.WithSuffix(cg.logger, keyvals),
		version:                cg.version,
		notCompatible:          cg.notCompatible,
		spec:                   cg.spec,
		endpointSliceSupported: cg.endpointSliceSupported,
	}
}

// WithMinimumVersion returns a new ConfigGenerator that does nothing (except
// logging a warning message) if the Prometheus version is lesser than the
// given version.
// The method panics if version isn't a valid SemVer value.
func (cg *ConfigGenerator) WithMinimumVersion(version string) *ConfigGenerator {
	minVersion := semver.MustParse(version)

	if cg.version.LT(minVersion) {
		return &ConfigGenerator{
			logger:                 log.WithSuffix(cg.logger, "minimum_version", version),
			version:                cg.version,
			notCompatible:          true,
			spec:                   cg.spec,
			endpointSliceSupported: cg.endpointSliceSupported,
		}
	}

	return cg
}

// WithMaximumVersion returns a new ConfigGenerator that does nothing (except
// logging a warning message) if the Prometheus version is greater than or
// equal to the given version.
// The method panics if version isn't a valid SemVer value.
func (cg *ConfigGenerator) WithMaximumVersion(version string) *ConfigGenerator {
	minVersion := semver.MustParse(version)

	if cg.version.GTE(minVersion) {
		return &ConfigGenerator{
			logger:                 log.WithSuffix(cg.logger, "maximum_version", version),
			version:                cg.version,
			notCompatible:          true,
			spec:                   cg.spec,
			endpointSliceSupported: cg.endpointSliceSupported,
		}
	}

	return cg
}

// AppendMapItem appends the k/v item to the given yaml.MapSlice and returns
// the updated slice.
func (cg *ConfigGenerator) AppendMapItem(m yaml.MapSlice, k string, v interface{}) yaml.MapSlice {
	if cg.notCompatible {
		level.Warn(cg.logger).Log("msg", fmt.Sprintf("ignoring %q not supported by Prometheus", k))
		return m
	}

	return append(m, yaml.MapItem{Key: k, Value: v})
}

type limitKey struct {
	specField       string
	prometheusField string
	minVersion      string
}

var (
	sampleLimitKey = limitKey{
		specField:       "sampleLimit",
		prometheusField: "sample_limit",
	}
	targetLimitKey = limitKey{
		specField:       "targetLimit",
		prometheusField: "target_limit",
		minVersion:      "2.21.0",
	}
	labelLimitKey = limitKey{
		specField:       "labelLimit",
		prometheusField: "label_limit",
		minVersion:      "2.27.0",
	}
	labelNameLengthLimitKey = limitKey{
		specField:       "labelNameLengthLimit",
		prometheusField: "label_name_length_limit",
		minVersion:      "2.27.0",
	}
	labelValueLengthLimitKey = limitKey{
		specField:       "labelValueLengthLimit",
		prometheusField: "label_value_length_limit",
		minVersion:      "2.27.0",
	}
)

// AddLimitsToYAML appends the given limit key to the configuration if
// supported by the Prometheus version.
func (cg *ConfigGenerator) AddLimitsToYAML(cfg yaml.MapSlice, k limitKey, limit uint64, enforcedLimit *uint64) yaml.MapSlice {
	if limit == 0 && enforcedLimit == nil {
		return cfg
	}

	if k.minVersion == "" {
		return cg.AppendMapItem(cfg, k.prometheusField, getLimit(limit, enforcedLimit))
	}

	return cg.WithMinimumVersion(k.minVersion).AppendMapItem(cfg, k.prometheusField, getLimit(limit, enforcedLimit))
}

// AddHonorTimestamps adds the honor_timestamps field into scrape configurations.
// honor_timestamps is false only when the user specified it or when the global
// override applies.
// For backwards compatibility with Prometheus <2.9.0 we don't set
// honor_timestamps.
func (cg *ConfigGenerator) AddHonorTimestamps(cfg yaml.MapSlice, userHonorTimestamps *bool) yaml.MapSlice {
	// Fast path.
	if userHonorTimestamps == nil && !cg.spec.OverrideHonorTimestamps {
		return cfg
	}

	honor := false
	if userHonorTimestamps != nil {
		honor = *userHonorTimestamps
	}

	return cg.WithMinimumVersion("2.9.0").AppendMapItem(cfg, "honor_timestamps", honor && !cg.spec.OverrideHonorTimestamps)
}

// AddHonorLabels adds the honor_labels field into scrape configurations.
// if OverrideHonorLabels is true then honor_labels is always false.
func (cg *ConfigGenerator) AddHonorLabels(cfg yaml.MapSlice, honorLabels bool) yaml.MapSlice {
	if cg.spec.OverrideHonorLabels {
		honorLabels = false
	}

	return cg.AppendMapItem(cfg, "honor_labels", honorLabels)
}

func (cg *ConfigGenerator) EndpointSliceSupported() bool {
	return cg.version.GTE(semver.MustParse("2.21.0")) && cg.endpointSliceSupported
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
	if tls == nil {
		return cfg
	}

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

	return cfg
}

func (cg *ConfigGenerator) addSafeAuthorizationToYaml(
	cfg yaml.MapSlice,
	assetStoreKey string,
	store *assets.Store,
	auth *v1.SafeAuthorization,
) yaml.MapSlice {
	if auth == nil {
		return cfg
	}

	authCfg := yaml.MapSlice{}
	if auth.Type == "" {
		auth.Type = "Bearer"
	}

	authCfg = append(authCfg, yaml.MapItem{Key: "type", Value: strings.TrimSpace(auth.Type)})
	if auth.Credentials != nil {
		if s, ok := store.TokenAssets[assetStoreKey]; ok {
			authCfg = append(authCfg, yaml.MapItem{Key: "credentials", Value: s})
		}
	}

	// extract current cfg section from assetStoreKey, assuming
	// "<component>/something..."
	return cg.WithMinimumVersion("2.26.0").WithKeyVals("component", strings.Split(assetStoreKey, "/")[0]).AppendMapItem(cfg, "authorization", authCfg)
}

func (cg *ConfigGenerator) addAuthorizationToYaml(
	cfg yaml.MapSlice,
	assetStoreKey string,
	store *assets.Store,
	auth *v1.Authorization,
) yaml.MapSlice {
	if auth == nil {
		return cfg
	}

	// reuse addSafeAuthorizationToYaml and unpack the part we're interested
	// in, namely the value under the "authorization" key
	authCfg := cg.addSafeAuthorizationToYaml(yaml.MapSlice{}, assetStoreKey, store, &auth.SafeAuthorization)[0].Value.(yaml.MapSlice)

	if auth.CredentialsFile != "" {
		authCfg = append(authCfg, yaml.MapItem{Key: "credentials_file", Value: auth.CredentialsFile})
	}

	return cg.WithMinimumVersion("2.26.0").WithKeyVals("component", strings.Split(assetStoreKey, "/")[0]).AppendMapItem(cfg, "authorization", authCfg)
}

func buildExternalLabels(p *v1.Prometheus) yaml.MapSlice {
	m := map[string]string{}

	prometheusExternalLabelName := "prometheus"
	if p.Spec.PrometheusExternalLabelName != nil {
		prometheusExternalLabelName = *p.Spec.PrometheusExternalLabelName
	}

	// Do not add the external label if the resulting value is empty.
	if prometheusExternalLabelName != "" {
		m[prometheusExternalLabelName] = fmt.Sprintf("%s/%s", p.Namespace, p.Name)
	}

	replicaExternalLabelName := defaultReplicaExternalLabelName
	if p.Spec.ReplicaExternalLabelName != nil {
		replicaExternalLabelName = *p.Spec.ReplicaExternalLabelName
	}

	// Do not add the external label if the resulting value is empty.
	if replicaExternalLabelName != "" {
		m[replicaExternalLabelName] = "$(POD_NAME)"
	}

	for n, v := range p.Spec.ExternalLabels {
		m[n] = v
	}
	return stringMapToMapSlice(m)
}

// validateConfigInputs runs extra validation on the Prometheus fields which can't be done at the CRD schema validation level.
func validateConfigInputs(p *v1.Prometheus) error {
	// TODO(slashpai): Remove this validation after v0.57 since this is handled at CRD level
	if p.Spec.EnforcedBodySizeLimit != "" {
		if err := operator.ValidateSizeField(string(p.Spec.EnforcedBodySizeLimit)); err != nil {
			return errors.Wrap(err, "invalid enforcedBodySizeLimit value specified")
		}
	}

	// TODO(slashpai): Remove this validation after v0.57 since this is handled at CRD level
	if p.Spec.RetentionSize != "" {
		if err := operator.ValidateSizeField(string(p.Spec.RetentionSize)); err != nil {
			return errors.Wrap(err, "invalid retentionSize value specified")
		}
	}

	if p.Spec.Retention != "" {
		if err := operator.ValidateDurationField(p.Spec.Retention); err != nil {
			return errors.Wrap(err, "invalid retention value specified")
		}
	}

	if p.Spec.ScrapeInterval != "" {
		if err := operator.ValidateDurationField(p.Spec.ScrapeInterval); err != nil {
			return errors.Wrap(err, "invalid scrapeInterval value specified")
		}
	}

	if p.Spec.ScrapeTimeout != "" {
		if err := operator.ValidateDurationField(p.Spec.ScrapeTimeout); err != nil {
			return errors.Wrap(err, "invalid scrapeTimeout value specified")
		}
		if err := operator.CompareScrapeTimeoutToScrapeInterval(p.Spec.ScrapeTimeout, p.Spec.ScrapeInterval); err != nil {
			return err
		}
	}

	if p.Spec.EvaluationInterval != "" {
		if err := operator.ValidateDurationField(p.Spec.EvaluationInterval); err != nil {
			return errors.Wrap(err, "invalid evaluationInterval value specified")
		}
	}

	if p.Spec.Thanos != nil && p.Spec.Thanos.ReadyTimeout != "" {
		if err := operator.ValidateDurationField(p.Spec.Thanos.ReadyTimeout); err != nil {
			return errors.Wrap(err, "invalid thanos.readyTimeout value specified")
		}
	}

	if p.Spec.Query != nil && p.Spec.Query.Timeout != nil && *p.Spec.Query.Timeout != "" {
		if err := operator.ValidateDurationField(*p.Spec.Query.Timeout); err != nil {
			return errors.Wrap(err, "invalid query.timeout value specified")
		}
	}

	for i, rr := range p.Spec.RemoteRead {
		if rr.RemoteTimeout != "" {
			if err := operator.ValidateDurationField(rr.RemoteTimeout); err != nil {
				return errors.Wrapf(err, "invalid remoteRead[%d].remoteTimeout value specified", i)
			}
		}
	}

	for i, rw := range p.Spec.RemoteWrite {
		if rw.RemoteTimeout != "" {
			if err := operator.ValidateDurationField(rw.RemoteTimeout); err != nil {
				return errors.Wrapf(err, "invalid remoteWrite[%d].remoteTimeout value specified", i)
			}
		}

		if rw.MetadataConfig != nil && rw.MetadataConfig.SendInterval != "" {
			if err := operator.ValidateDurationField(rw.MetadataConfig.SendInterval); err != nil {
				return errors.Wrapf(err, "invalid remoteWrite[%d].metadataConfig.sendInterval value specified", i)
			}
		}
	}

	if p.Spec.Alerting != nil {
		for i, ap := range p.Spec.Alerting.Alertmanagers {
			if ap.Timeout != nil && *ap.Timeout != "" {
				if err := operator.ValidateDurationField(*ap.Timeout); err != nil {
					return errors.Wrapf(err, "invalid alertmanagers[%d].timeout value specified", i)
				}
			}
		}
	}

	return nil
}

// Generate creates a serialized YAML representation of a Prometheus configuration using the provided resources.
func (cg *ConfigGenerator) Generate(
	p *v1.Prometheus,
	sMons map[string]*v1.ServiceMonitor,
	pMons map[string]*v1.PodMonitor,
	probes map[string]*v1.Probe,
	store *assets.Store,
	additionalScrapeConfigs []byte,
	additionalAlertRelabelConfigs []byte,
	additionalAlertManagerConfigs []byte,
	ruleConfigMapNames []string,
) ([]byte, error) {
	// Validate Prometheus Config Inputs at Prometheus CRD level
	if err := validateConfigInputs(p); err != nil {
		return nil, err
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

	if p.Spec.QueryLogFile != "" {
		globalItems = cg.WithMinimumVersion("2.16.0").AppendMapItem(globalItems, "query_log_file", queryLogFilePath(p))
	}

	cfg = append(cfg, yaml.MapItem{Key: "global", Value: globalItems})

	if p.Spec.RuleSelector != nil {
		ruleFilePaths := []string{}
		for _, name := range ruleConfigMapNames {
			ruleFilePaths = append(ruleFilePaths, rulesDir+"/"+name+"/*.yaml")
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "rule_files",
			Value: ruleFilePaths,
		})
	}

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
				cg.WithKeyVals("service_monitor", identifier).generateServiceMonitorConfig(
					sMons[identifier],
					ep, i,
					apiserverConfig,
					store,
					shards,
				),
			)
		}
	}
	for _, identifier := range pMonIdentifiers {
		for i, ep := range pMons[identifier].Spec.PodMetricsEndpoints {
			scrapeConfigs = append(scrapeConfigs,
				cg.WithKeyVals("pod_monitor", identifier).generatePodMonitorConfig(
					pMons[identifier], ep, i,
					apiserverConfig,
					store,
					shards,
				),
			)
		}
	}

	for _, identifier := range probeIdentifiers {
		scrapeConfigs = append(scrapeConfigs,
			cg.WithKeyVals("probe", identifier).generateProbeConfig(
				probes[identifier],
				apiserverConfig,
				store,
				shards,
			),
		)
	}

	var addlScrapeConfigs []yaml.MapSlice
	addlScrapeConfigs, err := cg.generateAdditionalScrapeConfigs(additionalScrapeConfigs, shards)
	if err != nil {
		return nil, errors.Wrap(err, "generate additional scrape configs")
	}

	cfg = append(cfg, yaml.MapItem{
		Key:   "scrape_configs",
		Value: append(scrapeConfigs, addlScrapeConfigs...),
	})

	cfg, err = cg.appendAlertingConfig(cfg, p, additionalAlertRelabelConfigs, additionalAlertManagerConfigs, store)
	if err != nil {
		return nil, errors.Wrap(err, "generating alerting configuration failed")
	}

	if len(p.Spec.RemoteWrite) > 0 {
		cfg = append(cfg, cg.generateRemoteWriteConfig(p, store))
	}

	if len(p.Spec.RemoteRead) > 0 {
		cfg = append(cfg, cg.generateRemoteReadConfig(p, store))
	}

	return yaml.Marshal(cfg)
}

func (cg *ConfigGenerator) appendAlertingConfig(
	cfg yaml.MapSlice,
	p *v1.Prometheus,
	additionalAlertRelabelConfigs []byte,
	additionalAlertmanagerConfigs []byte,
	store *assets.Store,
) (yaml.MapSlice, error) {
	if p.Spec.Alerting == nil && additionalAlertRelabelConfigs == nil && additionalAlertmanagerConfigs == nil {
		return cfg, nil
	}

	alertmanagerConfigs := cg.generateAlertmanagerConfig(p.Spec.Alerting, p.Spec.APIServerConfig, store)

	var additionalAlertmanagerConfigsYaml []yaml.MapSlice
	if err := yaml.Unmarshal([]byte(additionalAlertmanagerConfigs), &additionalAlertmanagerConfigsYaml); err != nil {
		return nil, errors.Wrap(err, "unmarshalling additional alertmanager configs failed")
	}
	alertmanagerConfigs = append(alertmanagerConfigs, additionalAlertmanagerConfigsYaml...)

	var alertRelabelConfigs []yaml.MapSlice

	replicaExternalLabelName := defaultReplicaExternalLabelName
	if p.Spec.ReplicaExternalLabelName != nil {
		replicaExternalLabelName = *p.Spec.ReplicaExternalLabelName
	}

	if replicaExternalLabelName != "" {
		// Drop the replica label to enable proper deduplication on the Alertmanager side.
		alertRelabelConfigs = append(alertRelabelConfigs, yaml.MapSlice{
			{Key: "action", Value: "labeldrop"},
			{Key: "regex", Value: regexp.QuoteMeta(replicaExternalLabelName)},
		})
	}

	var additionalAlertRelabelConfigsYaml []yaml.MapSlice
	if err := yaml.Unmarshal([]byte(additionalAlertRelabelConfigs), &additionalAlertRelabelConfigsYaml); err != nil {
		return nil, errors.Wrap(err, "unmarshalling additional alerting relabel configs failed")
	}
	alertRelabelConfigs = append(alertRelabelConfigs, additionalAlertRelabelConfigsYaml...)

	return append(cfg, yaml.MapItem{
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
	}), nil
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

func (cg *ConfigGenerator) generatePodMonitorConfig(
	m *v1.PodMonitor,
	ep v1.PodMetricsEndpoint,
	i int, apiserverConfig *v1.APIServerConfig,
	store *assets.Store,
	shards int32,
) yaml.MapSlice {
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i),
		},
	}
	cfg = cg.AddHonorLabels(cfg, ep.HonorLabels)
	cfg = cg.AddHonorTimestamps(cfg, ep.HonorTimestamps)

	cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.NamespaceSelector, m.Namespace, apiserverConfig, store, kubernetesSDRolePod))

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
	if ep.FollowRedirects != nil {
		cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "follow_redirects", *ep.FollowRedirects)
	}

	if ep.TLSConfig != nil {
		cfg = addSafeTLStoYaml(cfg, m.Namespace, ep.TLSConfig.SafeTLSConfig)
	}

	if ep.BearerTokenSecret.Name != "" {
		if s, ok := store.TokenAssets[fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	if ep.BasicAuth != nil {
		if s, ok := store.BasicAuthAssets[fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{
				Key: "basic_auth", Value: yaml.MapSlice{
					{Key: "username", Value: s.Username},
					{Key: "password", Value: s.Password},
				},
			})
		}
	}

	assetKey := fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	cfg = cg.addOAuth2ToYaml(cfg, ep.OAuth2, store.OAuth2Assets, assetKey)

	cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("podMonitor/auth/%s/%s/%d", m.Namespace, m.Name, i), store, ep.Authorization)

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
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(k), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(k)}},
			{Key: "regex", Value: fmt.Sprintf("(%s);true", m.Spec.Selector.MatchLabels[k])},
		})
	}
	// Set based label matching. We have to map the valid relations
	// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
	for _, exp := range m.Spec.Selector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
			})
		case metav1.LabelSelectorOpNotIn:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "drop"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_pod_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
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
		level.Warn(cg.logger).Log("msg", "'targetPort' is deprecated, use 'port' instead.")
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

	labeler := namespacelabeler.New(cg.spec.EnforcedNamespaceLabel, cg.spec.ExcludedFromEnforcement, false)
	relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs))...)

	relabelings = generateAddressShardingRelabelingRules(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cg.spec.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cg.spec.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cg.spec.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cg.spec.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cg.spec.EnforcedLabelValueLengthLimit)

	if cg.spec.EnforcedBodySizeLimit != "" {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", cg.spec.EnforcedBodySizeLimit)
	}

	cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs))})

	return cfg
}

func (cg *ConfigGenerator) generateProbeConfig(
	m *v1.Probe,
	apiserverConfig *v1.APIServerConfig,
	store *assets.Store,
	shards int32,
) yaml.MapSlice {

	jobName := fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: jobName,
		},
	}

	hTs := true
	cfg = cg.AddHonorTimestamps(cfg, &hTs)

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
	if m.Spec.ProberSpec.ProxyURL != "" {
		cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: m.Spec.ProberSpec.ProxyURL})
	}

	if m.Spec.Module != "" {
		cfg = append(cfg, yaml.MapItem{Key: "params", Value: yaml.MapSlice{
			{Key: "module", Value: []string{m.Spec.Module}},
		}})
	}

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cg.spec.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cg.spec.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cg.spec.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cg.spec.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cg.spec.EnforcedLabelValueLengthLimit)

	if cg.spec.EnforcedBodySizeLimit != "" {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", cg.spec.EnforcedBodySizeLimit)
	}

	relabelings := initRelabelings()

	if m.Spec.JobName != "" {
		relabelings = append(relabelings, []yaml.MapSlice{
			{
				{Key: "target_label", Value: "job"},
				{Key: "replacement", Value: m.Spec.JobName},
			},
		}...)
	}
	labeler := namespacelabeler.New(cg.spec.EnforcedNamespaceLabel, cg.spec.ExcludedFromEnforcement, false)

	if m.Spec.Targets.StaticConfig != nil {
		// Generate static_config section.
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
		xc := labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.StaticConfig.RelabelConfigs)
		relabelings = append(relabelings, generateRelabelConfig(xc)...)
		cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})
	} else {
		// Generate kubernetes_sd_config section for the ingress resources.

		// Filter targets by ingresses selected by the monitor.
		// Exact label matches.
		labelKeys := make([]string, 0, len(m.Spec.Targets.Ingress.Selector.MatchLabels))
		for k := range m.Spec.Targets.Ingress.Selector.MatchLabels {
			labelKeys = append(labelKeys, k)
		}
		sort.Strings(labelKeys)

		for _, k := range labelKeys {
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(k), "__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(k)}},
				{Key: "regex", Value: fmt.Sprintf("(%s);true", m.Spec.Targets.Ingress.Selector.MatchLabels[k])},
			})
		}

		// Set based label matching. We have to map the valid relations
		// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
		for _, exp := range m.Spec.Targets.Ingress.Selector.MatchExpressions {
			switch exp.Operator {
			case metav1.LabelSelectorOpIn:
				relabelings = append(relabelings, yaml.MapSlice{
					{Key: "action", Value: "keep"},
					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
					{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
				})
			case metav1.LabelSelectorOpNotIn:
				relabelings = append(relabelings, yaml.MapSlice{
					{Key: "action", Value: "drop"},
					{Key: "source_labels", Value: []string{"__meta_kubernetes_ingress_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_ingress_labelpresent_" + sanitizeLabelName(exp.Key)}},
					{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
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

		cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.Targets.Ingress.NamespaceSelector, m.Namespace, apiserverConfig, store, kubernetesSDRoleIngress))

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
				{Key: "source_labels", Value: []string{"__address__"}},
				{Key: "separator", Value: ";"},
				{Key: "regex", Value: "(.*)"},
				{Key: "target_label", Value: "__tmp_ingress_address"},
				{Key: "replacement", Value: "$1"},
				{Key: "action", Value: "replace"},
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
		relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.Ingress.RelabelConfigs))...)
		relabelings = generateAddressShardingRelabelingRulesForProbes(relabelings, shards)

		cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	}

	if m.Spec.TLSConfig != nil {
		cfg = addSafeTLStoYaml(cfg, m.Namespace, m.Spec.TLSConfig.SafeTLSConfig)
	}

	if m.Spec.BearerTokenSecret.Name != "" {
		pnKey := fmt.Sprintf("probe/%s/%s", m.GetNamespace(), m.GetName())
		if s, ok := store.TokenAssets[pnKey]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	if m.Spec.BasicAuth != nil {
		if s, ok := store.BasicAuthAssets[fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)]; ok {
			cfg = append(cfg, yaml.MapItem{
				Key: "basic_auth", Value: yaml.MapSlice{
					{Key: "username", Value: s.Username},
					{Key: "password", Value: s.Password},
				},
			})
		}
	}

	assetKey := fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)
	cfg = cg.addOAuth2ToYaml(cfg, m.Spec.OAuth2, store.OAuth2Assets, assetKey)

	cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("probe/auth/%s/%s", m.Namespace, m.Name), store, m.Spec.Authorization)

	cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.MetricRelabelConfigs))})

	return cfg
}

func (cg *ConfigGenerator) generateServiceMonitorConfig(
	m *v1.ServiceMonitor,
	ep v1.Endpoint,
	i int,
	apiserverConfig *v1.APIServerConfig,
	store *assets.Store,
	shards int32,
) yaml.MapSlice {
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i),
		},
	}
	cfg = cg.AddHonorLabels(cfg, ep.HonorLabels)
	cfg = cg.AddHonorTimestamps(cfg, ep.HonorTimestamps)

	role := kubernetesSDRoleEndpoint
	if cg.EndpointSliceSupported() {
		role = kubernetesSDRoleEndpointSlice
	}

	cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.NamespaceSelector, m.Namespace, apiserverConfig, store, role))

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
	if ep.FollowRedirects != nil {
		cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "follow_redirects", *ep.FollowRedirects)
	}

	assetKey := fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	cfg = cg.addOAuth2ToYaml(cfg, ep.OAuth2, store.OAuth2Assets, assetKey)

	cfg = addTLStoYaml(cfg, m.Namespace, ep.TLSConfig)

	if ep.BearerTokenFile != "" {
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: ep.BearerTokenFile})
	}

	if ep.BearerTokenSecret.Name != "" {
		if s, ok := store.TokenAssets[fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	if ep.BasicAuth != nil {
		if s, ok := store.BasicAuthAssets[fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{
				Key: "basic_auth", Value: yaml.MapSlice{
					{Key: "username", Value: s.Username},
					{Key: "password", Value: s.Password},
				},
			})
		}
	}

	cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("serviceMonitor/auth/%s/%s/%d", m.Namespace, m.Name, i), store, ep.Authorization)

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
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(k), "__meta_kubernetes_service_labelpresent_" + sanitizeLabelName(k)}},
			{Key: "regex", Value: fmt.Sprintf("(%s);true", m.Spec.Selector.MatchLabels[k])},
		})
	}
	// Set based label matching. We have to map the valid relations
	// `In`, `NotIn`, `Exists`, and `DoesNotExist`, into relabeling rules.
	for _, exp := range m.Spec.Selector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_service_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
			})
		case metav1.LabelSelectorOpNotIn:
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "drop"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(exp.Key), "__meta_kubernetes_service_labelpresent_" + sanitizeLabelName(exp.Key)}},
				{Key: "regex", Value: fmt.Sprintf("(%s);true", strings.Join(exp.Values, "|"))},
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
		sourceLabels := []string{"__meta_kubernetes_endpoint_port_name"}
		if cg.EndpointSliceSupported() {
			sourceLabels = []string{"__meta_kubernetes_endpointslice_port_name"}
		}
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			yaml.MapItem{Key: "source_labels", Value: sourceLabels},
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

	sourceLabels := []string{"__meta_kubernetes_endpoint_address_target_kind", "__meta_kubernetes_endpoint_address_target_name"}
	if cg.EndpointSliceSupported() {
		sourceLabels = []string{"__meta_kubernetes_endpointslice_address_target_kind", "__meta_kubernetes_endpointslice_address_target_name"}
	}

	// Relabel namespace and pod and service labels into proper labels.
	relabelings = append(relabelings, []yaml.MapSlice{
		{ // Relabel node labels with meta labels available with Prometheus >= v2.3.
			yaml.MapItem{Key: "source_labels", Value: sourceLabels},
			{Key: "separator", Value: ";"},
			{Key: "regex", Value: "Node;(.*)"},
			{Key: "replacement", Value: "${1}"},
			{Key: "target_label", Value: "node"},
		},
		{ // Relabel pod labels for >=v2.3 meta labels
			yaml.MapItem{Key: "source_labels", Value: sourceLabels},
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
	// value for it.
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

	// A single service may potentially have multiple metrics
	//	endpoints, therefore the endpoints labels is filled with the ports name or
	//	as a fallback the port number.
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

	labeler := namespacelabeler.New(cg.spec.EnforcedNamespaceLabel, cg.spec.ExcludedFromEnforcement, false)
	relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs))...)

	relabelings = generateAddressShardingRelabelingRules(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cg.spec.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cg.spec.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cg.spec.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cg.spec.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cg.spec.EnforcedLabelValueLengthLimit)

	if cg.spec.EnforcedBodySizeLimit != "" {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", cg.spec.EnforcedBodySizeLimit)
	}

	cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs))})

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
	return generateAddressShardingRelabelingRulesWithSourceLabel(relabelings, shards, "__address__")
}

func generateAddressShardingRelabelingRulesForProbes(relabelings []yaml.MapSlice, shards int32) []yaml.MapSlice {
	return generateAddressShardingRelabelingRulesWithSourceLabel(relabelings, shards, "__param_target")
}

func generateAddressShardingRelabelingRulesWithSourceLabel(relabelings []yaml.MapSlice, shards int32, shardLabel string) []yaml.MapSlice {
	return append(relabelings, yaml.MapSlice{
		{Key: "source_labels", Value: []string{shardLabel}},
		{Key: "target_label", Value: "__tmp_hash"},
		{Key: "modulus", Value: shards},
		{Key: "action", Value: "hashmod"},
	}, yaml.MapSlice{
		{Key: "source_labels", Value: []string{"__tmp_hash"}},
		{Key: "regex", Value: "$(SHARD)"},
		{Key: "action", Value: "keep"},
	})
}

func generateRelabelConfig(rc []*v1.RelabelConfig) []yaml.MapSlice {
	var cfg []yaml.MapSlice

	for _, c := range rc {
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

		cfg = append(cfg, relabeling)
	}
	return cfg
}

// GetNamespacesFromNamespaceSelector gets a list of namespaces to select based on
// the given namespace selector, the given default namespace, and whether to ignore namespace selectors
func (cg *ConfigGenerator) getNamespacesFromNamespaceSelector(nsel v1.NamespaceSelector, namespace string) []string {
	if cg.spec.IgnoreNamespaceSelectors {
		return []string{namespace}
	} else if nsel.Any {
		return []string{}
	} else if len(nsel.MatchNames) == 0 {
		return []string{namespace}
	}
	return nsel.MatchNames
}

// generateK8SSDConfig generates a kubernetes_sd_configs entry.
func (cg *ConfigGenerator) generateK8SSDConfig(
	namespaceSelector v1.NamespaceSelector,
	namespace string,
	apiserverConfig *v1.APIServerConfig,
	store *assets.Store,
	role string,
) yaml.MapItem {
	k8sSDConfig := yaml.MapSlice{
		{
			Key:   "role",
			Value: role,
		},
	}

	namespaces := cg.getNamespacesFromNamespaceSelector(namespaceSelector, namespace)
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

		if apiserverConfig.BasicAuth != nil && store.BasicAuthAssets != nil {
			if s, ok := store.BasicAuthAssets["apiserver"]; ok {
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

		k8sSDConfig = cg.addAuthorizationToYaml(k8sSDConfig, "apiserver/auth", store, apiserverConfig.Authorization)

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

func (cg *ConfigGenerator) generateAlertmanagerConfig(alerting *v1.AlertingSpec, apiserverConfig *v1.APIServerConfig, store *assets.Store) []yaml.MapSlice {
	if alerting == nil || len(alerting.Alertmanagers) == 0 {
		return nil
	}

	alertmanagerConfigs := make([]yaml.MapSlice, 0, len(alerting.Alertmanagers))
	for i, am := range alerting.Alertmanagers {
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

		cfg = append(cfg, cg.generateK8SSDConfig(v1.NamespaceSelector{}, am.Namespace, apiserverConfig, store, kubernetesSDRoleEndpoint))

		if am.BearerTokenFile != "" {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: am.BearerTokenFile})
		}

		cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("alertmanager/auth/%d", i), store, am.Authorization)

		if am.APIVersion == "v1" || am.APIVersion == "v2" {
			cfg = cg.WithMinimumVersion("2.11.0").AppendMapItem(cfg, "api_version", am.APIVersion)
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
		alertmanagerConfigs = append(alertmanagerConfigs, cfg)
	}

	return alertmanagerConfigs
}

func (cg *ConfigGenerator) generateAdditionalScrapeConfigs(
	additionalScrapeConfigs []byte,
	shards int32,
) ([]yaml.MapSlice, error) {
	var additionalScrapeConfigsYaml []yaml.MapSlice
	err := yaml.Unmarshal([]byte(additionalScrapeConfigs), &additionalScrapeConfigsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling additional scrape configs failed")
	}
	if shards == 1 {
		return additionalScrapeConfigsYaml, nil
	}

	var addlScrapeConfigs []yaml.MapSlice
	for _, mapSlice := range additionalScrapeConfigsYaml {
		var addlScrapeConfig yaml.MapSlice
		var relabelings []yaml.MapSlice
		var otherConfigItems []yaml.MapItem
		for _, mapItem := range mapSlice {
			if mapItem.Key != "relabel_configs" {
				otherConfigItems = append(otherConfigItems, mapItem)
				continue
			}
			values, ok := mapItem.Value.([]interface{})
			if !ok {
				return nil, errors.Wrap(err, "error parsing relabel configs")
			}
			for _, value := range values {
				relabeling, ok := value.(yaml.MapSlice)
				if !ok {
					return nil, errors.Wrap(err, "error parsing relabel config")
				}
				relabelings = append(relabelings, relabeling)
			}
		}
		relabelings = generateAddressShardingRelabelingRules(relabelings, shards)
		addlScrapeConfig = append(addlScrapeConfig, otherConfigItems...)
		addlScrapeConfig = append(addlScrapeConfig, yaml.MapItem{Key: "relabel_configs", Value: relabelings})
		addlScrapeConfigs = append(addlScrapeConfigs, addlScrapeConfig)
	}
	return addlScrapeConfigs, nil
}

func (cg *ConfigGenerator) generateRemoteReadConfig(
	p *v1.Prometheus,
	store *assets.Store,
) yaml.MapItem {
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

		if len(spec.Headers) > 0 {
			cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "headers", stringMapToMapSlice(spec.Headers))
		}

		if spec.Name != "" {
			cfg = cg.WithMinimumVersion("2.15.0").AppendMapItem(cfg, "name", spec.Name)
		}

		if len(spec.RequiredMatchers) > 0 {
			cfg = append(cfg, yaml.MapItem{Key: "required_matchers", Value: stringMapToMapSlice(spec.RequiredMatchers)})
		}

		if spec.ReadRecent {
			cfg = append(cfg, yaml.MapItem{Key: "read_recent", Value: spec.ReadRecent})
		}

		if spec.BasicAuth != nil {
			if s, ok := store.BasicAuthAssets[fmt.Sprintf("remoteRead/%d", i)]; ok {
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

		cfg = cg.addOAuth2ToYaml(cfg, spec.OAuth2, store.OAuth2Assets, fmt.Sprintf("remoteRead/%d", i))

		cfg = addTLStoYaml(cfg, p.ObjectMeta.Namespace, spec.TLSConfig)

		cfg = cg.addAuthorizationToYaml(cfg, fmt.Sprintf("remoteRead/auth/%d", i), store, spec.Authorization)

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

func (cg *ConfigGenerator) addOAuth2ToYaml(
	cfg yaml.MapSlice,
	oauth2 *v1.OAuth2,
	tlsAssets map[string]assets.OAuth2Credentials,
	assetKey string,
) yaml.MapSlice {
	if oauth2 == nil {
		return cfg
	}

	tlsAsset, ok := tlsAssets[assetKey]
	if !ok {
		return cfg
	}

	oauth2Cfg := yaml.MapSlice{}
	oauth2Cfg = append(oauth2Cfg,
		yaml.MapItem{Key: "client_id", Value: tlsAsset.ClientID},
		yaml.MapItem{Key: "client_secret", Value: tlsAsset.ClientSecret},
		yaml.MapItem{Key: "token_url", Value: oauth2.TokenURL},
	)

	if len(oauth2.Scopes) > 0 {
		oauth2Cfg = append(oauth2Cfg, yaml.MapItem{Key: "scopes", Value: oauth2.Scopes})
	}

	if len(oauth2.EndpointParams) > 0 {
		oauth2Cfg = append(oauth2Cfg, yaml.MapItem{Key: "endpoint_params", Value: oauth2.EndpointParams})
	}

	return cg.WithMinimumVersion("2.27.0").AppendMapItem(cfg, "oauth2", oauth2Cfg)
}

func (cg *ConfigGenerator) generateRemoteWriteConfig(
	p *v1.Prometheus,
	store *assets.Store,
) yaml.MapItem {

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
		if len(spec.Headers) > 0 {
			cfg = cg.WithMinimumVersion("2.15.0").AppendMapItem(cfg, "headers", stringMapToMapSlice(spec.Headers))
		}

		if spec.Name != "" {
			cfg = cg.WithMinimumVersion("2.15.0").AppendMapItem(cfg, "name", spec.Name)
		}

		if spec.SendExemplars != nil {
			cfg = cg.WithMinimumVersion("2.27.0").AppendMapItem(cfg, "send_exemplars", spec.SendExemplars)
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
			if s, ok := store.BasicAuthAssets[fmt.Sprintf("remoteWrite/%d", i)]; ok {
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

		cfg = cg.addOAuth2ToYaml(cfg, spec.OAuth2, store.OAuth2Assets, fmt.Sprintf("remoteWrite/%d", i))

		cfg = addTLStoYaml(cfg, p.ObjectMeta.Namespace, spec.TLSConfig)

		cfg = cg.addAuthorizationToYaml(cfg, fmt.Sprintf("remoteWrite/auth/%d", i), store, spec.Authorization)

		if spec.ProxyURL != "" {
			cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: spec.ProxyURL})
		}

		if spec.Sigv4 != nil {
			sigV4 := yaml.MapSlice{}
			if spec.Sigv4.Region != "" {
				sigV4 = append(sigV4, yaml.MapItem{Key: "region", Value: spec.Sigv4.Region})
			}
			key := fmt.Sprintf("remoteWrite/%d", i)
			if store.SigV4Assets[key].AccessKeyID != "" {
				sigV4 = append(sigV4, yaml.MapItem{Key: "access_key", Value: store.SigV4Assets[key].AccessKeyID})
			}
			if store.SigV4Assets[key].SecretKeyID != "" {
				sigV4 = append(sigV4, yaml.MapItem{Key: "secret_key", Value: store.SigV4Assets[key].SecretKeyID})
			}
			if spec.Sigv4.Profile != "" {
				sigV4 = append(sigV4, yaml.MapItem{Key: "profile", Value: spec.Sigv4.Profile})
			}
			if spec.Sigv4.RoleArn != "" {
				sigV4 = append(sigV4, yaml.MapItem{Key: "role_arn", Value: spec.Sigv4.RoleArn})
			}

			cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "sigv4", sigV4)
		}

		if spec.QueueConfig != nil {
			queueConfig := yaml.MapSlice{}

			if spec.QueueConfig.Capacity != int(0) {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "capacity", Value: spec.QueueConfig.Capacity})
			}

			if spec.QueueConfig.MinShards != int(0) {
				queueConfig = cg.WithMinimumVersion("2.6.0").AppendMapItem(queueConfig, "min_shards", spec.QueueConfig.MinShards)
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
				queueConfig = cg.WithMaximumVersion("2.11.0").AppendMapItem(queueConfig, "max_retries", spec.QueueConfig.MaxRetries)
			}

			if spec.QueueConfig.MinBackoff != "" {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "min_backoff", Value: spec.QueueConfig.MinBackoff})
			}

			if spec.QueueConfig.MaxBackoff != "" {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "max_backoff", Value: spec.QueueConfig.MaxBackoff})
			}

			if spec.QueueConfig.RetryOnRateLimit {
				queueConfig = cg.WithMinimumVersion("2.26.0").AppendMapItem(queueConfig, "retry_on_http_429", spec.QueueConfig.RetryOnRateLimit)
			}

			cfg = append(cfg, yaml.MapItem{Key: "queue_config", Value: queueConfig})
		}

		if spec.MetadataConfig != nil {
			metadataConfig := append(yaml.MapSlice{}, yaml.MapItem{Key: "send", Value: spec.MetadataConfig.Send})
			if spec.MetadataConfig.SendInterval != "" {
				metadataConfig = append(metadataConfig, yaml.MapItem{Key: "send_interval", Value: spec.MetadataConfig.SendInterval})
			}

			cfg = cg.WithMinimumVersion("2.23.0").AppendMapItem(cfg, "metadata_config", metadataConfig)
		}

		cfgs = append(cfgs, cfg)
	}

	return yaml.MapItem{
		Key:   "remote_write",
		Value: cfgs,
	}
}
