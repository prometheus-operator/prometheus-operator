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
	"context"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/units"
	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	kubernetesSDRoleEndpoint      = "endpoints"
	kubernetesSDRoleEndpointSlice = "endpointslice"
	kubernetesSDRolePod           = "pod"
	kubernetesSDRoleIngress       = "ingress"

	defaultPrometheusExternalLabelName = "prometheus"
	defaultReplicaExternalLabelName    = "prometheus_replica"
)

var invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}

// ConfigGenerator knows how to generate a Prometheus configuration which is
// compatible with a given Prometheus version.
type ConfigGenerator struct {
	logger                 log.Logger
	version                semver.Version
	notCompatible          bool
	prom                   monitoringv1.PrometheusInterface
	endpointSliceSupported bool
	scrapeClasses          map[string]*monitoringv1.ScrapeClass
	defaultScrapeClassName string
}

// NewConfigGenerator creates a ConfigGenerator for the provided Prometheus resource.
func NewConfigGenerator(logger log.Logger, p monitoringv1.PrometheusInterface, endpointSliceSupported bool) (*ConfigGenerator, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	cpf := p.GetCommonPrometheusFields()

	promVersion := operator.StringValOrDefault(cpf.Version, operator.DefaultPrometheusVersion)
	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus version: %w", err)
	}

	if version.Major != 2 {
		return nil, fmt.Errorf("unsupported Prometheus major version %s: %w", version, err)
	}

	logger = log.WithSuffix(logger, "version", promVersion)

	scrapeClasses, defaultScrapeClassName, err := getScrapeClassConfig(p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scrape classes: %w", err)
	}

	return &ConfigGenerator{
		logger:                 logger,
		version:                version,
		prom:                   p,
		endpointSliceSupported: endpointSliceSupported,
		scrapeClasses:          scrapeClasses,
		defaultScrapeClassName: defaultScrapeClassName,
	}, nil
}

func getScrapeClassConfig(p monitoringv1.PrometheusInterface) (map[string]*monitoringv1.ScrapeClass, string, error) {
	cpf := p.GetCommonPrometheusFields()
	scrapeClasses := make(map[string]*monitoringv1.ScrapeClass, len(cpf.ScrapeClasses))
	defaultScrapeClass := ""
	for _, scrapeClass := range cpf.ScrapeClasses {
		scrapeClasses[scrapeClass.Name] = &scrapeClass
		// Validate all scrape class relabelings are correct.
		if err := validateRelabelConfigs(p, scrapeClass.Relabelings); err != nil {
			return nil, "", fmt.Errorf("invalid relabelings for scrapeClass %s: %w", scrapeClass.Name, err)
		}
		if ptr.Deref(scrapeClass.Default, false) {
			if defaultScrapeClass == "" {
				defaultScrapeClass = scrapeClass.Name
				continue
			}
			return nil, "", fmt.Errorf("multiple default scrape classes defined")
		}
	}
	return scrapeClasses, defaultScrapeClass, nil
}

// WithKeyVals returns a new ConfigGenerator with the same characteristics as
// the current object, expect that the keyvals are appended to the existing
// logger.
func (cg *ConfigGenerator) WithKeyVals(keyvals ...interface{}) *ConfigGenerator {
	return &ConfigGenerator{
		logger:                 log.WithSuffix(cg.logger, keyvals...),
		version:                cg.version,
		notCompatible:          cg.notCompatible,
		prom:                   cg.prom,
		endpointSliceSupported: cg.endpointSliceSupported,
		scrapeClasses:          cg.scrapeClasses,
		defaultScrapeClassName: cg.defaultScrapeClassName,
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
			prom:                   cg.prom,
			endpointSliceSupported: cg.endpointSliceSupported,
			scrapeClasses:          cg.scrapeClasses,
			defaultScrapeClassName: cg.defaultScrapeClassName,
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
			prom:                   cg.prom,
			endpointSliceSupported: cg.endpointSliceSupported,
			scrapeClasses:          cg.scrapeClasses,
			defaultScrapeClassName: cg.defaultScrapeClassName,
		}
	}

	return cg
}

// AppendMapItem appends the k/v item to the given yaml.MapSlice and returns
// the updated slice.
func (cg *ConfigGenerator) AppendMapItem(m yaml.MapSlice, k string, v interface{}) yaml.MapSlice {
	if cg.notCompatible {
		cg.Warn(k)
		return m
	}

	return append(m, yaml.MapItem{Key: k, Value: v})
}

// AppendCommandlineArgument appends the name/v argument to the given []monitoringv1.Argument and returns
// the updated slice.
func (cg *ConfigGenerator) AppendCommandlineArgument(m []monitoringv1.Argument, argument monitoringv1.Argument) []monitoringv1.Argument {
	if cg.notCompatible {
		level.Warn(cg.logger).Log("msg", fmt.Sprintf("ignoring command line argument %q not supported by Prometheus", argument.Name))
		return m
	}

	return append(m, argument)
}

// IsCompatible return true or false depending if the version being used is compatible.
func (cg *ConfigGenerator) IsCompatible() bool {
	return !cg.notCompatible
}

// Warn logs a warning.
func (cg *ConfigGenerator) Warn(field string) {
	level.Warn(cg.logger).Log("msg", fmt.Sprintf("ignoring %q not supported by Prometheus", field))
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
	keepDroppedTargetsKey = limitKey{
		specField:       "keepDroppedTargets",
		prometheusField: "keep_dropped_targets",
		minVersion:      "2.47.0",
	}
)

// AddLimitsToYAML appends the given limit key to the configuration if
// supported by the Prometheus version.
func (cg *ConfigGenerator) AddLimitsToYAML(cfg yaml.MapSlice, k limitKey, limit *uint64, enforcedLimit *uint64) yaml.MapSlice {
	finalLimit := getLimit(limit, enforcedLimit)
	if finalLimit == nil {
		return cfg
	}

	if k.minVersion == "" {
		return cg.AppendMapItem(cfg, k.prometheusField, finalLimit)
	}

	return cg.WithMinimumVersion(k.minVersion).AppendMapItem(cfg, k.prometheusField, finalLimit)
}

// AddHonorTimestamps adds the honor_timestamps field into scrape configurations.
// honor_timestamps is false only when the user specified it or when the global
// override applies.
// For backwards compatibility with Prometheus <2.9.0 we don't set
// honor_timestamps.
func (cg *ConfigGenerator) AddHonorTimestamps(cfg yaml.MapSlice, userHonorTimestamps *bool) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()
	// Fast path.
	if userHonorTimestamps == nil && !cpf.OverrideHonorTimestamps {
		return cfg
	}

	honor := false
	if userHonorTimestamps != nil {
		honor = *userHonorTimestamps
	}

	return cg.WithMinimumVersion("2.9.0").AppendMapItem(cfg, "honor_timestamps", honor && !cpf.OverrideHonorTimestamps)
}

// AddTrackTimestampsStaleness adds the track_timestamps_staleness field into scrape configurations.
// For backwards compatibility with Prometheus <2.48.0 we don't set
// track_timestamps_staleness.
func (cg *ConfigGenerator) AddTrackTimestampsStaleness(cfg yaml.MapSlice, trackTimestampsStaleness *bool) yaml.MapSlice {
	// Fast path.
	if trackTimestampsStaleness == nil {
		return cfg
	}

	return cg.WithMinimumVersion("2.48.0").AppendMapItem(cfg, "track_timestamps_staleness", *trackTimestampsStaleness)
}

// AddScrapeProtocols adds the scrape_protocols field into scrape configurations.
// For backwards compatibility with Prometheus <2.49.0 we don't set scrape_protocols.
func (cg *ConfigGenerator) AddScrapeProtocols(cfg yaml.MapSlice, scrapeProtocols []monitoringv1.ScrapeProtocol) yaml.MapSlice {
	if len(scrapeProtocols) == 0 {
		return cfg
	}

	return cg.WithMinimumVersion("2.49.0").AppendMapItem(cfg, "scrape_protocols", scrapeProtocols)
}

// AddHonorLabels adds the honor_labels field into scrape configurations.
// if OverrideHonorLabels is true then honor_labels is always false.
func (cg *ConfigGenerator) AddHonorLabels(cfg yaml.MapSlice, honorLabels bool) yaml.MapSlice {
	if cg.prom.GetCommonPrometheusFields().OverrideHonorLabels {
		honorLabels = false
	}

	return cg.AppendMapItem(cfg, "honor_labels", honorLabels)
}

func (cg *ConfigGenerator) EndpointSliceSupported() bool {
	return cg.version.GTE(semver.MustParse("2.21.0")) && cg.endpointSliceSupported
}

// stringMapToMapSlice returns a yaml.MapSlice from a string map to ensure that
// the output is deterministic.
func stringMapToMapSlice[V any](m map[string]V) yaml.MapSlice {
	res := yaml.MapSlice{}
	ks := make([]string, 0, len(m))

	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	for _, k := range ks {
		res = append(res, yaml.MapItem{Key: k, Value: m[k]})
	}

	return res
}

func addSafeTLStoYaml(cfg yaml.MapSlice, namespace string, tls monitoringv1.SafeTLSConfig) yaml.MapSlice {
	pathForSelector := func(sel monitoringv1.SecretOrConfigMap) string {
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
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "key_file", Value: pathForSelector(monitoringv1.SecretOrConfigMap{Secret: tls.KeySecret})})
	}
	if tls.ServerName != "" {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "server_name", Value: tls.ServerName})
	}
	cfg = append(cfg, yaml.MapItem{Key: "tls_config", Value: tlsConfig})
	return cfg
}

func addTLStoYaml(cfg yaml.MapSlice, namespace string, tls *monitoringv1.TLSConfig) yaml.MapSlice {
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

func (cg *ConfigGenerator) MergeTLSConfigWithScrapeClass(tlsConfig *monitoringv1.TLSConfig, scrapeClass *monitoringv1.ScrapeClass) *monitoringv1.TLSConfig {
	scrapeClassName := cg.defaultScrapeClassName

	if scrapeClass != nil {
		scrapeClassName = scrapeClass.Name
	}

	scrapeClass, found := cg.scrapeClasses[scrapeClassName]

	if !found {
		return tlsConfig
	}

	if tlsConfig == nil && !found {
		return nil
	}

	if tlsConfig != nil && !found {
		return tlsConfig
	}

	if tlsConfig == nil && found {
		return scrapeClass.TLSConfig
	}

	if tlsConfig.CAFile == "" && tlsConfig.SafeTLSConfig.CA == (monitoringv1.SecretOrConfigMap{}) {
		tlsConfig.CAFile = scrapeClass.TLSConfig.CAFile
	}

	if tlsConfig.CertFile == "" && tlsConfig.SafeTLSConfig.Cert == (monitoringv1.SecretOrConfigMap{}) {
		tlsConfig.CertFile = scrapeClass.TLSConfig.CertFile
	}

	if tlsConfig.KeyFile == "" && tlsConfig.SafeTLSConfig.KeySecret == nil {
		tlsConfig.KeyFile = scrapeClass.TLSConfig.KeyFile
	}

	return tlsConfig
}

func (cg *ConfigGenerator) addBasicAuthToYaml(cfg yaml.MapSlice,
	assetStoreKey string,
	store *assets.Store,
	basicAuth *monitoringv1.BasicAuth,
) yaml.MapSlice {
	if basicAuth == nil {
		return cfg
	}

	var authCfg assets.BasicAuthCredentials
	if s, ok := store.BasicAuthAssets[assetStoreKey]; ok {
		authCfg = s
	}

	return cg.WithKeyVals("component", strings.Split(assetStoreKey, "/")[0]).AppendMapItem(cfg, "basic_auth", authCfg)
}

func (cg *ConfigGenerator) addSigv4ToYaml(cfg yaml.MapSlice,
	assetStoreKey string,
	store *assets.Store,
	sigv4 *monitoringv1.Sigv4,
) yaml.MapSlice {
	if sigv4 == nil {
		return cfg
	}

	sigv4Cfg := yaml.MapSlice{}
	if sigv4.Region != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "region", Value: sigv4.Region})
	}
	if store.SigV4Assets[assetStoreKey].AccessKeyID != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "access_key", Value: store.SigV4Assets[assetStoreKey].AccessKeyID})
	}
	if store.SigV4Assets[assetStoreKey].SecretKeyID != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "secret_key", Value: store.SigV4Assets[assetStoreKey].SecretKeyID})
	}
	if sigv4.Profile != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "profile", Value: sigv4.Profile})
	}
	if sigv4.RoleArn != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "role_arn", Value: sigv4.RoleArn})
	}

	return cg.WithKeyVals("component", strings.Split(assetStoreKey, "/")[0]).AppendMapItem(cfg, "sigv4", sigv4Cfg)
}

func (cg *ConfigGenerator) addSafeAuthorizationToYaml(
	cfg yaml.MapSlice,
	assetStoreKey string,
	store *assets.Store,
	auth *monitoringv1.SafeAuthorization,
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
	auth *monitoringv1.Authorization,
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

func (cg *ConfigGenerator) buildExternalLabels() yaml.MapSlice {
	m := map[string]string{}
	cpf := cg.prom.GetCommonPrometheusFields()
	objMeta := cg.prom.GetObjectMeta()

	prometheusExternalLabelName := defaultPrometheusExternalLabelName
	if cpf.PrometheusExternalLabelName != nil {
		prometheusExternalLabelName = *cpf.PrometheusExternalLabelName
	}

	// Do not add the external label if the resulting value is empty.
	if prometheusExternalLabelName != "" {
		m[prometheusExternalLabelName] = fmt.Sprintf("%s/%s", objMeta.GetNamespace(), objMeta.GetName())
	}

	replicaExternalLabelName := defaultReplicaExternalLabelName
	if cpf.ReplicaExternalLabelName != nil {
		replicaExternalLabelName = *cpf.ReplicaExternalLabelName
	}

	// Do not add the external label if the resulting value is empty.
	if replicaExternalLabelName != "" {
		m[replicaExternalLabelName] = fmt.Sprintf("$(%s)", operator.PodNameEnvVar)
	}

	for k, v := range cpf.ExternalLabels {
		if _, found := m[k]; found {
			level.Warn(cg.logger).Log("msg", "ignoring external label because it is a reserved key", "key", k)
			continue
		}
		m[k] = v
	}

	return stringMapToMapSlice(m)
}

func (cg *ConfigGenerator) addProxyConfigtoYaml(
	ctx context.Context,
	cfg yaml.MapSlice,
	namespace string,
	store *assets.Store,
	proxyConfig monitoringv1.ProxyConfig,
) yaml.MapSlice {
	if reflect.ValueOf(proxyConfig).IsZero() {
		return cfg
	}

	var cgProxyConfig = cg.WithMinimumVersion("2.43.0")

	if proxyConfig.ProxyURL != nil {
		cfg = cgProxyConfig.AppendMapItem(cfg, "proxy_url", *proxyConfig.ProxyURL)
	}

	if proxyConfig.NoProxy != nil {
		cfg = cgProxyConfig.AppendMapItem(cfg, "no_proxy", *proxyConfig.NoProxy)
	}

	if proxyConfig.ProxyFromEnvironment != nil {
		cfg = cgProxyConfig.AppendMapItem(cfg, "proxy_from_environment", *proxyConfig.ProxyFromEnvironment)
	}

	if proxyConfig.ProxyConnectHeader != nil {
		proxyConnectHeader := make(map[string]string, len(proxyConfig.ProxyConnectHeader))

		for k, v := range proxyConfig.ProxyConnectHeader {
			value, _ := store.GetKey(ctx, namespace, monitoringv1.SecretOrConfigMap{
				Secret: &v,
			})
			proxyConnectHeader[k] = value
		}

		cfg = cgProxyConfig.AppendMapItem(cfg, "proxy_connect_header", stringMapToMapSlice(proxyConnectHeader))
	}

	return cfg
}

// CompareScrapeTimeoutToScrapeInterval validates value of scrapeTimeout based on scrapeInterval.
func CompareScrapeTimeoutToScrapeInterval(scrapeTimeout, scrapeInterval monitoringv1.Duration) error {
	var si, st model.Duration
	var err error

	if si, err = model.ParseDuration(string(scrapeInterval)); err != nil {
		return fmt.Errorf("invalid scrapeInterval %q: %w", scrapeInterval, err)
	}

	if st, err = model.ParseDuration(string(scrapeTimeout)); err != nil {
		return fmt.Errorf("invalid scrapeTimeout: %q: %w", scrapeTimeout, err)
	}

	if st > si {
		return fmt.Errorf("scrapeTimeout %q greater than scrapeInterval %q", scrapeTimeout, scrapeInterval)
	}

	return nil
}

// GenerateServerConfiguration creates a serialized YAML representation of a Prometheus Server configuration using the provided resources.
func (cg *ConfigGenerator) GenerateServerConfiguration(
	ctx context.Context,
	evaluationInterval monitoringv1.Duration,
	queryLogFile string,
	ruleSelector *metav1.LabelSelector,
	exemplars *monitoringv1.Exemplars,
	tsdb monitoringv1.TSDBSpec,
	alerting *monitoringv1.AlertingSpec,
	remoteRead []monitoringv1.RemoteReadSpec,
	sMons map[string]*monitoringv1.ServiceMonitor,
	pMons map[string]*monitoringv1.PodMonitor,
	probes map[string]*monitoringv1.Probe,
	sCons map[string]*monitoringv1alpha1.ScrapeConfig,
	store *assets.Store,
	additionalScrapeConfigs []byte,
	additionalAlertRelabelConfigs []byte,
	additionalAlertManagerConfigs []byte,
	ruleConfigMapNames []string,
) ([]byte, error) {
	cpf := cg.prom.GetCommonPrometheusFields()

	// validates the value of scrapeTimeout based on scrapeInterval
	if cpf.ScrapeTimeout != "" {
		if err := CompareScrapeTimeoutToScrapeInterval(cpf.ScrapeTimeout, cpf.ScrapeInterval); err != nil {
			return nil, err
		}
	}

	// Global config
	cfg := yaml.MapSlice{}
	globalItems := yaml.MapSlice{}
	globalItems = cg.appendEvaluationInterval(globalItems, evaluationInterval)
	globalItems = cg.appendScrapeIntervals(globalItems)
	globalItems = cg.appendScrapeProtocols(globalItems)
	globalItems = cg.appendExternalLabels(globalItems)
	globalItems = cg.appendQueryLogFile(globalItems, queryLogFile)
	globalItems = cg.appendScrapeLimits(globalItems)
	cfg = append(cfg, yaml.MapItem{Key: "global", Value: globalItems})

	// Rule Files config
	cfg = cg.appendRuleFiles(cfg, ruleConfigMapNames, ruleSelector)

	// Scrape config
	var (
		scrapeConfigs   []yaml.MapSlice
		apiserverConfig = cpf.APIServerConfig
		shards          = shardsNumber(cg.prom)
	)

	scrapeConfigs = cg.appendServiceMonitorConfigs(scrapeConfigs, sMons, apiserverConfig, store, shards)
	scrapeConfigs = cg.appendPodMonitorConfigs(scrapeConfigs, pMons, apiserverConfig, store, shards)
	scrapeConfigs = cg.appendProbeConfigs(scrapeConfigs, probes, apiserverConfig, store, shards)
	scrapeConfigs, err := cg.appendScrapeConfigs(ctx, scrapeConfigs, sCons, store, shards)
	if err != nil {
		return nil, fmt.Errorf("generate scrape configs: %w", err)
	}

	scrapeConfigs, err = cg.appendAdditionalScrapeConfigs(scrapeConfigs, additionalScrapeConfigs, shards)
	if err != nil {
		return nil, fmt.Errorf("generate additional scrape configs: %w", err)
	}
	cfg = append(cfg, yaml.MapItem{
		Key:   "scrape_configs",
		Value: scrapeConfigs,
	})

	// Storage config
	cfg, err = cg.appendStorageSettingsConfig(cfg, exemplars, tsdb)
	if err != nil {
		return nil, fmt.Errorf("generating storage_settings configuration failed: %w", err)
	}

	// Alerting config
	cfg, err = cg.appendAlertingConfig(cfg, alerting, additionalAlertRelabelConfigs, additionalAlertManagerConfigs, store)
	if err != nil {
		return nil, fmt.Errorf("generating alerting configuration failed: %w", err)
	}

	// Remote write config
	if len(cpf.RemoteWrite) > 0 {
		cfg = append(cfg, cg.generateRemoteWriteConfig(store))
	}

	// Remote read config
	if len(remoteRead) > 0 {
		cfg = append(cfg, cg.generateRemoteReadConfig(remoteRead, store))
	}

	if cpf.TracingConfig != nil {
		tracingcfg, err := cg.generateTracingConfig()

		if err != nil {
			return nil, fmt.Errorf("generating tracing configuration failed: %w", err)
		}

		cfg = append(cfg, tracingcfg)
	}

	return yaml.Marshal(cfg)
}

func (cg *ConfigGenerator) appendStorageSettingsConfig(cfg yaml.MapSlice, exemplars *monitoringv1.Exemplars, tsdb monitoringv1.TSDBSpec) (yaml.MapSlice, error) {
	var (
		storage   yaml.MapSlice
		cgStorage = cg.WithMinimumVersion("2.29.0")
	)

	if exemplars != nil && exemplars.MaxSize != nil {
		storage = cgStorage.AppendMapItem(storage, "exemplars", yaml.MapSlice{
			{
				Key:   "max_exemplars",
				Value: *exemplars.MaxSize,
			},
		})
	}

	if tsdb.OutOfOrderTimeWindow != "" {
		storage = cg.WithMinimumVersion("2.39.0").AppendMapItem(storage, "tsdb", yaml.MapSlice{
			{
				Key:   "out_of_order_time_window",
				Value: tsdb.OutOfOrderTimeWindow,
			},
		})
	}

	if len(storage) == 0 {
		return cfg, nil
	}

	return cgStorage.AppendMapItem(cfg, "storage", storage), nil
}

func (cg *ConfigGenerator) appendAlertingConfig(
	cfg yaml.MapSlice,
	alerting *monitoringv1.AlertingSpec,
	additionalAlertRelabelConfigs []byte,
	additionalAlertmanagerConfigs []byte,
	store *assets.Store,
) (yaml.MapSlice, error) {
	if alerting == nil && additionalAlertRelabelConfigs == nil && additionalAlertmanagerConfigs == nil {
		return cfg, nil
	}

	cpf := cg.prom.GetCommonPrometheusFields()

	alertmanagerConfigs := cg.generateAlertmanagerConfig(alerting, cpf.APIServerConfig, store)

	var additionalAlertmanagerConfigsYaml []yaml.MapSlice
	if err := yaml.Unmarshal(additionalAlertmanagerConfigs, &additionalAlertmanagerConfigsYaml); err != nil {
		return nil, fmt.Errorf("unmarshalling additional alertmanager configs failed")
	}
	alertmanagerConfigs = append(alertmanagerConfigs, additionalAlertmanagerConfigsYaml...)

	var alertRelabelConfigs []yaml.MapSlice

	replicaExternalLabelName := defaultReplicaExternalLabelName
	if cpf.ReplicaExternalLabelName != nil {
		replicaExternalLabelName = *cpf.ReplicaExternalLabelName
	}

	if replicaExternalLabelName != "" {
		// Drop the replica label to enable proper deduplication on the Alertmanager side.
		alertRelabelConfigs = append(alertRelabelConfigs, yaml.MapSlice{
			{Key: "action", Value: "labeldrop"},
			{Key: "regex", Value: regexp.QuoteMeta(replicaExternalLabelName)},
		})
	}

	var additionalAlertRelabelConfigsYaml []yaml.MapSlice
	if err := yaml.Unmarshal(additionalAlertRelabelConfigs, &additionalAlertRelabelConfigsYaml); err != nil {
		return nil, fmt.Errorf("unmarshalling additional alerting relabel configs failed: %w", err)
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
	m *monitoringv1.PodMonitor,
	ep monitoringv1.PodMetricsEndpoint,
	i int, apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.Store,
	shards int32,
) yaml.MapSlice {
	scrapeClass := cg.getScrapeClassOrDefault(m.Spec.ScrapeClassName)
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i),
		},
	}
	cfg = cg.AddHonorLabels(cfg, ep.HonorLabels)
	cfg = cg.AddHonorTimestamps(cfg, ep.HonorTimestamps)
	cfg = cg.AddTrackTimestampsStaleness(cfg, ep.TrackTimestampsStaleness)

	var attachMetaConfig *attachMetadataConfig
	if m.Spec.AttachMetadata != nil {
		attachMetaConfig = &attachMetadataConfig{
			MinimumVersion: "2.35.0",
			AttachMetadata: m.Spec.AttachMetadata,
		}
	}

	cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.NamespaceSelector, m.Namespace, apiserverConfig, store, kubernetesSDRolePod, attachMetaConfig))

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
	if ep.EnableHttp2 != nil {
		cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *ep.EnableHttp2)
	}

	if ep.TLSConfig != nil {
		tlsConfig := &monitoringv1.TLSConfig{SafeTLSConfig: *ep.TLSConfig}
		mergedTLSConfig := cg.MergeTLSConfigWithScrapeClass(tlsConfig, scrapeClass)
		cfg = addTLStoYaml(cfg, m.Namespace, mergedTLSConfig)
	}

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if ep.BearerTokenSecret.Name != "" {
		level.Warn(cg.logger).Log("msg", "'bearerTokenSecret' is deprecated, use 'authorization' instead.")
		if s, ok := store.TokenAssets[fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	cfg = cg.addBasicAuthToYaml(cfg, fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i), store, ep.BasicAuth)

	assetKey := fmt.Sprintf("podMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	cfg = cg.addOAuth2ToYaml(cfg, ep.OAuth2, store.OAuth2Assets, assetKey)

	cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("podMonitor/auth/%s/%s/%d", m.Namespace, m.Name, i), store, ep.Authorization)

	relabelings := initRelabelings()

	if ep.FilterRunning == nil || *ep.FilterRunning {
		relabelings = append(relabelings, generateRunningFilter())
	}

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
	cpf := cg.prom.GetCommonPrometheusFields()
	for _, l := range append(m.Spec.PodTargetLabels, cpf.PodTargetLabels...) {
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

	// Add scrape class relabelings if there is any.
	if scrapeClass != nil {
		relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)
	}

	labeler := namespacelabeler.New(cpf.EnforcedNamespaceLabel, cpf.ExcludedFromEnforcement, false)
	relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs))...)

	relabelings = generateAddressShardingRelabelingRules(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, m.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	cfg = cg.AddScrapeProtocols(cfg, m.Spec.ScrapeProtocols)

	if bodySizeLimit := getLowerByteSize(m.Spec.BodySizeLimit, &cpf); !isByteSizeEmpty(bodySizeLimit) {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", bodySizeLimit)
	}

	cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs))})

	return cfg
}

// generateProbeConfig builds the prometheus configuration for a probe. This function
// assumes that it will never receive a probe with empty targets, since the
// operator filters those in the validation step in SelectProbes().
func (cg *ConfigGenerator) generateProbeConfig(
	m *monitoringv1.Probe,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.Store,
	shards int32,
) yaml.MapSlice {
	scrapeClass := cg.getScrapeClassOrDefault(m.Spec.ScrapeClassName)

	jobName := fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: jobName,
		},
	}

	hTs := true
	cfg = cg.AddHonorTimestamps(cfg, &hTs)

	cfg = append(cfg, yaml.MapItem{Key: "metrics_path", Value: m.Spec.ProberSpec.Path})

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

	cpf := cg.prom.GetCommonPrometheusFields()
	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, m.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	cfg = cg.AddScrapeProtocols(cfg, m.Spec.ScrapeProtocols)

	if cpf.EnforcedBodySizeLimit != "" {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", cpf.EnforcedBodySizeLimit)
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
	labeler := namespacelabeler.New(cpf.EnforcedNamespaceLabel, cpf.ExcludedFromEnforcement, false)

	// As stated in the CRD documentation, if both StaticConfig and Ingress are
	// defined, the former takes precedence which is why the first case statement
	// checks for m.Spec.Targets.StaticConfig.
	switch {
	case m.Spec.Targets.StaticConfig != nil:
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

		// Add scrape class relabelings if there is any.
		if scrapeClass != nil {
			relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)
		}

		// Add configured relabelings.
		xc := labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.StaticConfig.RelabelConfigs)
		relabelings = append(relabelings, generateRelabelConfig(xc)...)

	case m.Spec.Targets.Ingress != nil:
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

		cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.Targets.Ingress.NamespaceSelector, m.Namespace, apiserverConfig, store, kubernetesSDRoleIngress, nil))

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

		// Add scrape class relabelings if there is any.
		if scrapeClass != nil {
			relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)
		}

		// Add configured relabelings.
		relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.Ingress.RelabelConfigs))...)
	}

	relabelings = generateAddressShardingRelabelingRulesForProbes(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	if m.Spec.TLSConfig != nil {
		tlsConfig := &monitoringv1.TLSConfig{SafeTLSConfig: *m.Spec.TLSConfig}
		mergedTLSConfig := cg.MergeTLSConfigWithScrapeClass(tlsConfig, scrapeClass)
		cfg = addTLStoYaml(cfg, m.Namespace, mergedTLSConfig)
	}

	if m.Spec.BearerTokenSecret.Name != "" {
		pnKey := fmt.Sprintf("probe/%s/%s", m.GetNamespace(), m.GetName())
		if s, ok := store.TokenAssets[pnKey]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	cfg = cg.addBasicAuthToYaml(cfg, fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name), store, m.Spec.BasicAuth)

	assetKey := fmt.Sprintf("probe/%s/%s", m.Namespace, m.Name)
	cfg = cg.addOAuth2ToYaml(cfg, m.Spec.OAuth2, store.OAuth2Assets, assetKey)

	cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("probe/auth/%s/%s", m.Namespace, m.Name), store, m.Spec.Authorization)

	cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.MetricRelabelConfigs))})

	return cfg
}

func (cg *ConfigGenerator) generateServiceMonitorConfig(
	m *monitoringv1.ServiceMonitor,
	ep monitoringv1.Endpoint,
	i int,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.Store,
	shards int32,
) yaml.MapSlice {
	scrapeClass := cg.getScrapeClassOrDefault(m.Spec.ScrapeClassName)

	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i),
		},
	}
	cfg = cg.AddHonorLabels(cfg, ep.HonorLabels)
	cfg = cg.AddHonorTimestamps(cfg, ep.HonorTimestamps)
	cfg = cg.AddTrackTimestampsStaleness(cfg, ep.TrackTimestampsStaleness)

	role := kubernetesSDRoleEndpoint
	if cg.EndpointSliceSupported() {
		role = kubernetesSDRoleEndpointSlice
	}

	var attachMetaConfig *attachMetadataConfig
	if m.Spec.AttachMetadata != nil {
		attachMetaConfig = &attachMetadataConfig{
			MinimumVersion: "2.37.0",
			AttachMetadata: m.Spec.AttachMetadata,
		}
	}

	cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.NamespaceSelector, m.Namespace, apiserverConfig, store, role, attachMetaConfig))

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
	if ep.EnableHttp2 != nil {
		cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *ep.EnableHttp2)
	}
	assetKey := fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)
	cfg = cg.addOAuth2ToYaml(cfg, ep.OAuth2, store.OAuth2Assets, assetKey)

	mergedTLSConfig := cg.MergeTLSConfigWithScrapeClass(ep.TLSConfig, scrapeClass)
	cfg = addTLStoYaml(cfg, m.Namespace, mergedTLSConfig)

	if ep.BearerTokenFile != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		level.Warn(cg.logger).Log("msg", "'bearerTokenFile' is deprecated, use 'authorization' instead.")
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: ep.BearerTokenFile}) //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	}

	if ep.BearerTokenSecret != nil && ep.BearerTokenSecret.Name != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		level.Warn(cg.logger).Log("msg", "'bearerTokenSecret' is deprecated, use 'authorization' instead.")
		if s, ok := store.TokenAssets[fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i)]; ok {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: s})
		}
	}

	cfg = cg.addBasicAuthToYaml(cfg, fmt.Sprintf("serviceMonitor/%s/%s/%d", m.Namespace, m.Name, i), store, ep.BasicAuth)

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

	if ptr.Deref(ep.FilterRunning, true) {
		relabelings = append(relabelings, generateRunningFilter())
	}

	// Relabel targetLabels from Service onto target.
	for _, l := range m.Spec.TargetLabels {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_label_" + sanitizeLabelName(l)}},
			{Key: "target_label", Value: sanitizeLabelName(l)},
			{Key: "regex", Value: "(.+)"},
			{Key: "replacement", Value: "${1}"},
		})
	}

	cpf := cg.prom.GetCommonPrometheusFields()
	for _, l := range append(m.Spec.PodTargetLabels, cpf.PodTargetLabels...) {
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

	// Add scrape class relabelings if there is any.
	if scrapeClass != nil {
		relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)
	}

	labeler := namespacelabeler.New(cpf.EnforcedNamespaceLabel, cpf.ExcludedFromEnforcement, false)
	relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs))...)

	relabelings = generateAddressShardingRelabelingRules(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, m.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	cfg = cg.AddScrapeProtocols(cfg, m.Spec.ScrapeProtocols)

	if bodySizeLimit := getLowerByteSize(m.Spec.BodySizeLimit, &cpf); !isByteSizeEmpty(bodySizeLimit) {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", bodySizeLimit)
	}

	cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs))})

	return cfg
}

func generateRunningFilter() yaml.MapSlice {
	return yaml.MapSlice{
		{Key: "action", Value: "drop"},
		{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_phase"}},
		{Key: "regex", Value: "(Failed|Succeeded)"},
	}
}

func getLimit(user *uint64, enforced *uint64) *uint64 {
	if enforced == nil {
		return user
	}

	if user == nil {
		return enforced
	}

	if *enforced > *user {
		return user
	}

	return enforced
}

func generateAddressShardingRelabelingRules(relabelings []yaml.MapSlice, shards int32) []yaml.MapSlice {
	return generateAddressShardingRelabelingRulesWithSourceLabel(relabelings, shards, "__address__")
}

func (cg *ConfigGenerator) generateAddressShardingRelabelingRulesIfMissing(relabelings []yaml.MapSlice, shards int32) []yaml.MapSlice {
	for i, relabeling := range relabelings {
		for _, relabelItem := range relabeling {
			if relabelItem.Key == "action" && relabelItem.Value == "hashmod" {
				level.Debug(cg.logger).Log("msg", "found existing hashmod relabeling rule, skipping", "idx", i)
				return relabelings
			}
		}
	}
	return generateAddressShardingRelabelingRules(relabelings, shards)
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
		{Key: "regex", Value: fmt.Sprintf("$(%s)", operator.ShardEnvVar)},
		{Key: "action", Value: "keep"},
	})
}

func generateRelabelConfig(rc []*monitoringv1.RelabelConfig) []yaml.MapSlice {
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
			relabeling = append(relabeling, yaml.MapItem{Key: "action", Value: strings.ToLower(c.Action)})
		}

		cfg = append(cfg, relabeling)
	}
	return cfg
}

// GetNamespacesFromNamespaceSelector gets a list of namespaces to select based on
// the given namespace selector, the given default namespace, and whether to ignore namespace selectors.
func (cg *ConfigGenerator) getNamespacesFromNamespaceSelector(nsel monitoringv1.NamespaceSelector, namespace string) []string {
	if cg.prom.GetCommonPrometheusFields().IgnoreNamespaceSelectors {
		return []string{namespace}
	} else if nsel.Any {
		return []string{}
	} else if len(nsel.MatchNames) == 0 {
		return []string{namespace}
	}
	return nsel.MatchNames
}

type attachMetadataConfig struct {
	MinimumVersion string
	AttachMetadata *monitoringv1.AttachMetadata
}

// generateK8SSDConfig generates a kubernetes_sd_configs entry.
func (cg *ConfigGenerator) generateK8SSDConfig(
	namespaceSelector monitoringv1.NamespaceSelector,
	namespace string,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.Store,
	role string,
	attachMetadataConfig *attachMetadataConfig,
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

		k8sSDConfig = cg.addBasicAuthToYaml(k8sSDConfig, "apiserver", store, apiserverConfig.BasicAuth)

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if apiserverConfig.BearerToken != "" {
			level.Warn(cg.logger).Log("msg", "'bearerToken' is deprecated, use 'authorization' instead.")
			k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token", Value: apiserverConfig.BearerToken})
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if apiserverConfig.BearerTokenFile != "" {
			level.Warn(cg.logger).Log("msg", "'bearerTokenFile' is deprecated, use 'authorization' instead.")
			k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token_file", Value: apiserverConfig.BearerTokenFile})
		}

		k8sSDConfig = cg.addAuthorizationToYaml(k8sSDConfig, "apiserver/auth", store, apiserverConfig.Authorization)

		// TODO: If we want to support secret refs for k8s service discovery tls
		// config as well, make sure to path the right namespace here.
		k8sSDConfig = addTLStoYaml(k8sSDConfig, "", apiserverConfig.TLSConfig)
	}
	if attachMetadataConfig != nil {
		k8sSDConfig = cg.WithMinimumVersion(attachMetadataConfig.MinimumVersion).AppendMapItem(k8sSDConfig, "attach_metadata", yaml.MapSlice{
			{Key: "node", Value: attachMetadataConfig.AttachMetadata.Node},
		})
	}

	return yaml.MapItem{
		Key: "kubernetes_sd_configs",
		Value: []yaml.MapSlice{
			k8sSDConfig,
		},
	}
}

func (cg *ConfigGenerator) generateAlertmanagerConfig(alerting *monitoringv1.AlertingSpec, apiserverConfig *monitoringv1.APIServerConfig, store *assets.Store) []yaml.MapSlice {
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

		if am.EnableHttp2 != nil {
			cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *am.EnableHttp2)
		}

		// TODO: If we want to support secret refs for alertmanager config tls
		// config as well, make sure to path the right namespace here.
		cfg = addTLStoYaml(cfg, "", am.TLSConfig)

		cfg = append(cfg, cg.generateK8SSDConfig(monitoringv1.NamespaceSelector{}, am.Namespace, apiserverConfig, store, kubernetesSDRoleEndpoint, nil))

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if am.BearerTokenFile != "" {
			level.Warn(cg.logger).Log("msg", "'bearerTokenFile' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: am.BearerTokenFile})
		}

		cfg = cg.WithMinimumVersion("2.26.0").addBasicAuthToYaml(cfg, fmt.Sprintf("alertmanager/auth/%d", i), store, am.BasicAuth)

		cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("alertmanager/auth/%d", i), store, am.Authorization)

		cfg = cg.WithMinimumVersion("2.48.0").addSigv4ToYaml(cfg, fmt.Sprintf("alertmanager/auth/%d", i), store, am.Sigv4)

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
	err := yaml.Unmarshal(additionalScrapeConfigs, &additionalScrapeConfigsYaml)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling additional scrape configs failed: %w", err)
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
				return nil, fmt.Errorf("error parsing relabel configs: %w", err)
			}
			for _, value := range values {
				relabeling, ok := value.(yaml.MapSlice)
				if !ok {
					return nil, fmt.Errorf("error parsing relabel config: %w", err)
				}
				relabelings = append(relabelings, relabeling)
			}
		}
		relabelings = cg.generateAddressShardingRelabelingRulesIfMissing(relabelings, shards)
		addlScrapeConfig = append(addlScrapeConfig, otherConfigItems...)
		addlScrapeConfig = append(addlScrapeConfig, yaml.MapItem{Key: "relabel_configs", Value: relabelings})
		addlScrapeConfigs = append(addlScrapeConfigs, addlScrapeConfig)
	}
	return addlScrapeConfigs, nil
}

func (cg *ConfigGenerator) generateRemoteReadConfig(
	remoteRead []monitoringv1.RemoteReadSpec,
	store *assets.Store,
) yaml.MapItem {
	cfgs := []yaml.MapSlice{}
	objMeta := cg.prom.GetObjectMeta()

	for i, spec := range remoteRead {
		// defaults
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

		cfg = cg.addBasicAuthToYaml(cfg, fmt.Sprintf("remoteRead/%d", i), store, spec.BasicAuth)

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerToken != "" {
			level.Warn(cg.logger).Log("msg", "'bearerToken' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: spec.BearerToken})
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerTokenFile != "" {
			level.Warn(cg.logger).Log("msg", "'bearerTokenFile' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: spec.BearerTokenFile})
		}

		cfg = cg.addOAuth2ToYaml(cfg, spec.OAuth2, store.OAuth2Assets, fmt.Sprintf("remoteRead/%d", i))

		cfg = addTLStoYaml(cfg, objMeta.GetNamespace(), spec.TLSConfig)

		cfg = cg.addAuthorizationToYaml(cfg, fmt.Sprintf("remoteRead/auth/%d", i), store, spec.Authorization)

		if spec.ProxyURL != "" {
			cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: spec.ProxyURL})
		}

		if spec.FollowRedirects != nil {
			cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "follow_redirects", spec.FollowRedirects)
		}

		if spec.FilterExternalLabels != nil {
			cfg = cg.WithMinimumVersion("2.34.0").AppendMapItem(cfg, "filter_external_labels", spec.FilterExternalLabels)
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
	oauth2 *monitoringv1.OAuth2,
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
	store *assets.Store,
) yaml.MapItem {
	cfgs := []yaml.MapSlice{}
	cpf := cg.prom.GetCommonPrometheusFields()
	objMeta := cg.prom.GetObjectMeta()

	for i, spec := range cpf.RemoteWrite {
		// defaults
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

		if spec.SendNativeHistograms != nil {
			cfg = cg.WithMinimumVersion("2.40.0").AppendMapItem(cfg, "send_native_histograms", spec.SendNativeHistograms)
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
					relabeling = append(relabeling, yaml.MapItem{Key: "action", Value: strings.ToLower(c.Action)})
				}
				relabelings = append(relabelings, relabeling)
			}

			cfg = append(cfg, yaml.MapItem{Key: "write_relabel_configs", Value: relabelings})

		}

		cfg = cg.addBasicAuthToYaml(cfg, fmt.Sprintf("remoteWrite/%d", i), store, spec.BasicAuth)

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerToken != "" {
			level.Warn(cg.logger).Log("msg", "'bearerToken' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: spec.BearerToken})
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerTokenFile != "" {
			level.Warn(cg.logger).Log("msg", "'bearerTokenFile' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: spec.BearerTokenFile})
		}

		cfg = cg.addOAuth2ToYaml(cfg, spec.OAuth2, store.OAuth2Assets, fmt.Sprintf("remoteWrite/%d", i))

		cfg = addTLStoYaml(cfg, objMeta.GetNamespace(), spec.TLSConfig)

		cfg = cg.addAuthorizationToYaml(cfg, fmt.Sprintf("remoteWrite/auth/%d", i), store, spec.Authorization)

		if spec.ProxyURL != "" {
			cfg = append(cfg, yaml.MapItem{Key: "proxy_url", Value: spec.ProxyURL})
		}

		cfg = cg.WithMinimumVersion("2.26.0").addSigv4ToYaml(cfg, fmt.Sprintf("remoteWrite/%d", i), store, spec.Sigv4)

		if spec.AzureAD != nil {
			azureAd := yaml.MapSlice{}

			if spec.AzureAD.ManagedIdentity != nil {
				azureAd = append(azureAd,
					yaml.MapItem{Key: "managed_identity", Value: yaml.MapSlice{
						{Key: "client_id", Value: spec.AzureAD.ManagedIdentity.ClientID},
					}},
				)
			}

			if spec.AzureAD.OAuth != nil {
				azureAd = cg.WithMinimumVersion("2.48.0").AppendMapItem(azureAd, "oauth", yaml.MapSlice{
					{Key: "client_id", Value: spec.AzureAD.OAuth.ClientID},
					{Key: "client_secret", Value: store.AzureOAuthAssets[fmt.Sprintf("remoteWrite/%d", i)].ClientSecret},
					{Key: "tenant_id", Value: spec.AzureAD.OAuth.TenantID},
				})
			}

			if spec.AzureAD.Cloud != nil {
				azureAd = append(azureAd, yaml.MapItem{Key: "cloud", Value: spec.AzureAD.Cloud})
			}

			cfg = cg.WithMinimumVersion("2.45.0").AppendMapItem(cfg, "azuread", azureAd)
		}

		if spec.EnableHttp2 != nil {
			cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *spec.EnableHttp2)
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

			if spec.QueueConfig.BatchSendDeadline != nil {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "batch_send_deadline", Value: string(*spec.QueueConfig.BatchSendDeadline)})
			}

			if spec.QueueConfig.MaxRetries != int(0) {
				queueConfig = cg.WithMaximumVersion("2.11.0").AppendMapItem(queueConfig, "max_retries", spec.QueueConfig.MaxRetries)
			}

			if spec.QueueConfig.MinBackoff != nil {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "min_backoff", Value: string(*spec.QueueConfig.MinBackoff)})
			}

			if spec.QueueConfig.MaxBackoff != nil {
				queueConfig = append(queueConfig, yaml.MapItem{Key: "max_backoff", Value: string(*spec.QueueConfig.MaxBackoff)})
			}

			if spec.QueueConfig.RetryOnRateLimit {
				queueConfig = cg.WithMinimumVersion("2.26.0").AppendMapItem(queueConfig, "retry_on_http_429", spec.QueueConfig.RetryOnRateLimit)
			}

			if spec.QueueConfig.SampleAgeLimit != nil {
				queueConfig = cg.WithMinimumVersion("2.50.0").AppendMapItem(queueConfig, "sample_age_limit", string(*spec.QueueConfig.SampleAgeLimit))
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

func (cg *ConfigGenerator) appendScrapeIntervals(slice yaml.MapSlice) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()
	slice = append(slice, yaml.MapItem{Key: "scrape_interval", Value: cpf.ScrapeInterval})

	if cpf.ScrapeTimeout != "" {
		slice = append(slice, yaml.MapItem{
			Key: "scrape_timeout", Value: cpf.ScrapeTimeout,
		})
	}

	return slice
}

func (cg *ConfigGenerator) appendScrapeProtocols(slice yaml.MapSlice) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()

	if len(cpf.ScrapeProtocols) == 0 {
		return slice
	}

	return cg.WithMinimumVersion("2.49.0").AppendMapItem(slice, "scrape_protocols", cpf.ScrapeProtocols)
}

func (cg *ConfigGenerator) appendEvaluationInterval(slice yaml.MapSlice, evaluationInterval monitoringv1.Duration) yaml.MapSlice {
	return append(slice, yaml.MapItem{Key: "evaluation_interval", Value: evaluationInterval})
}

func (cg *ConfigGenerator) appendScrapeLimits(slice yaml.MapSlice) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()
	if cpf.BodySizeLimit != nil {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "body_size_limit", cpf.BodySizeLimit)
	}
	if cpf.SampleLimit != nil {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "sample_limit", *cpf.SampleLimit)
	}
	if cpf.TargetLimit != nil {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "target_limit", *cpf.TargetLimit)
	}
	if cpf.LabelLimit != nil {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "label_limit", *cpf.LabelLimit)
	}
	if cpf.LabelNameLengthLimit != nil {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "label_name_length_limit", *cpf.LabelNameLengthLimit)
	}
	if cpf.LabelValueLengthLimit != nil {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "label_value_length_limit", *cpf.LabelValueLengthLimit)
	}
	if cpf.KeepDroppedTargets != nil {
		slice = cg.WithMinimumVersion("2.47.0").AppendMapItem(slice, "keep_dropped_targets", *cpf.KeepDroppedTargets)
	}

	return slice
}

func (cg *ConfigGenerator) appendExternalLabels(slice yaml.MapSlice) yaml.MapSlice {
	slice = append(slice, yaml.MapItem{
		Key:   "external_labels",
		Value: cg.buildExternalLabels(),
	})

	return slice
}

func (cg *ConfigGenerator) appendQueryLogFile(slice yaml.MapSlice, queryLogFile string) yaml.MapSlice {
	if queryLogFile != "" {
		slice = cg.WithMinimumVersion("2.16.0").AppendMapItem(slice, "query_log_file", queryLogFilePath(queryLogFile))
	}

	return slice
}

func (cg *ConfigGenerator) appendRuleFiles(slice yaml.MapSlice, ruleFiles []string, ruleSelector *metav1.LabelSelector) yaml.MapSlice {
	if ruleSelector != nil {
		ruleFilePaths := []string{}
		for _, name := range ruleFiles {
			ruleFilePaths = append(ruleFilePaths, RulesDir+"/"+name+"/*.yaml")
		}
		slice = append(slice, yaml.MapItem{
			Key:   "rule_files",
			Value: ruleFilePaths,
		})
	}

	return slice
}

func (cg *ConfigGenerator) appendServiceMonitorConfigs(
	slices []yaml.MapSlice,
	serviceMonitors map[string]*monitoringv1.ServiceMonitor,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.Store,
	shards int32) []yaml.MapSlice {
	sMonIdentifiers := make([]string, len(serviceMonitors))
	i := 0
	for k := range serviceMonitors {
		sMonIdentifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(sMonIdentifiers)

	for _, identifier := range sMonIdentifiers {
		for i, ep := range serviceMonitors[identifier].Spec.Endpoints {
			slices = append(slices,
				cg.WithKeyVals("service_monitor", identifier).generateServiceMonitorConfig(
					serviceMonitors[identifier],
					ep, i,
					apiserverConfig,
					store,
					shards,
				),
			)
		}
	}

	return slices
}

func (cg *ConfigGenerator) appendPodMonitorConfigs(
	slices []yaml.MapSlice,
	podMonitors map[string]*monitoringv1.PodMonitor,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.Store,
	shards int32) []yaml.MapSlice {
	pMonIdentifiers := make([]string, len(podMonitors))
	i := 0
	for k := range podMonitors {
		pMonIdentifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(pMonIdentifiers)

	for _, identifier := range pMonIdentifiers {
		for i, ep := range podMonitors[identifier].Spec.PodMetricsEndpoints {
			slices = append(slices,
				cg.WithKeyVals("pod_monitor", identifier).generatePodMonitorConfig(
					podMonitors[identifier], ep, i,
					apiserverConfig,
					store,
					shards,
				),
			)
		}
	}

	return slices
}

func (cg *ConfigGenerator) appendProbeConfigs(
	slices []yaml.MapSlice,
	probes map[string]*monitoringv1.Probe,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.Store,
	shards int32) []yaml.MapSlice {
	probeIdentifiers := make([]string, len(probes))
	i := 0
	for k := range probes {
		probeIdentifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(probeIdentifiers)

	for _, identifier := range probeIdentifiers {
		slices = append(slices,
			cg.WithKeyVals("probe", identifier).generateProbeConfig(
				probes[identifier],
				apiserverConfig,
				store,
				shards,
			),
		)
	}

	return slices
}

func (cg *ConfigGenerator) appendAdditionalScrapeConfigs(scrapeConfigs []yaml.MapSlice, additionalScrapeConfigs []byte, shards int32) ([]yaml.MapSlice, error) {
	addlScrapeConfigs, err := cg.generateAdditionalScrapeConfigs(additionalScrapeConfigs, shards)
	if err != nil {
		return nil, err
	}

	return append(scrapeConfigs, addlScrapeConfigs...), nil
}

// GenerateAgentConfiguration creates a serialized YAML representation of a Prometheus Agent configuration using the provided resources.
func (cg *ConfigGenerator) GenerateAgentConfiguration(
	ctx context.Context,
	sMons map[string]*monitoringv1.ServiceMonitor,
	pMons map[string]*monitoringv1.PodMonitor,
	probes map[string]*monitoringv1.Probe,
	sCons map[string]*monitoringv1alpha1.ScrapeConfig,
	store *assets.Store,
	additionalScrapeConfigs []byte,
) ([]byte, error) {
	cpf := cg.prom.GetCommonPrometheusFields()

	// validates the value of scrapeTimeout based on scrapeInterval
	if cpf.ScrapeTimeout != "" {
		if err := CompareScrapeTimeoutToScrapeInterval(cpf.ScrapeTimeout, cpf.ScrapeInterval); err != nil {
			return nil, err
		}
	}

	// Global config
	cfg := yaml.MapSlice{}
	globalItems := yaml.MapSlice{}
	globalItems = cg.appendScrapeIntervals(globalItems)
	globalItems = cg.appendScrapeProtocols(globalItems)
	globalItems = cg.appendExternalLabels(globalItems)
	globalItems = cg.appendScrapeLimits(globalItems)
	cfg = append(cfg, yaml.MapItem{Key: "global", Value: globalItems})

	// Scrape config
	var (
		scrapeConfigs   []yaml.MapSlice
		apiserverConfig = cpf.APIServerConfig
		shards          = shardsNumber(cg.prom)
	)

	scrapeConfigs = cg.appendServiceMonitorConfigs(scrapeConfigs, sMons, apiserverConfig, store, shards)
	scrapeConfigs = cg.appendPodMonitorConfigs(scrapeConfigs, pMons, apiserverConfig, store, shards)
	scrapeConfigs = cg.appendProbeConfigs(scrapeConfigs, probes, apiserverConfig, store, shards)
	scrapeConfigs, err := cg.appendScrapeConfigs(ctx, scrapeConfigs, sCons, store, shards)
	if err != nil {
		return nil, fmt.Errorf("generate scrape configs: %w", err)
	}

	scrapeConfigs, err = cg.appendAdditionalScrapeConfigs(scrapeConfigs, additionalScrapeConfigs, shards)
	if err != nil {
		return nil, fmt.Errorf("generate additional scrape configs: %w", err)
	}
	cfg = append(cfg, yaml.MapItem{
		Key:   "scrape_configs",
		Value: scrapeConfigs,
	})

	// Remote write config
	if len(cpf.RemoteWrite) > 0 {
		cfg = append(cfg, cg.generateRemoteWriteConfig(store))
	}

	if cpf.TracingConfig != nil {
		tracingcfg, err := cg.generateTracingConfig()
		if err != nil {
			return nil, fmt.Errorf("generating tracing configuration failed: %w", err)
		}

		cfg = append(cfg, tracingcfg)
	}

	return yaml.Marshal(cfg)
}

func (cg *ConfigGenerator) appendScrapeConfigs(
	ctx context.Context,
	slices []yaml.MapSlice,
	scrapeConfigs map[string]*monitoringv1alpha1.ScrapeConfig,
	store *assets.Store,
	shards int32) ([]yaml.MapSlice, error) {
	scrapeConfigIdentifiers := make([]string, len(scrapeConfigs))
	i := 0
	for k := range scrapeConfigs {
		scrapeConfigIdentifiers[i] = k
		i++
	}

	// Sorting ensures, that we always generate the config in the same order.
	sort.Strings(scrapeConfigIdentifiers)

	for _, identifier := range scrapeConfigIdentifiers {
		cfgGenerator := cg.WithKeyVals("scrapeconfig", identifier)
		scrapeConfig, err := cfgGenerator.generateScrapeConfig(ctx, scrapeConfigs[identifier], store, shards)

		if err != nil {
			return slices, err
		}

		slices = append(slices, scrapeConfig)
	}

	return slices, nil
}

func (cg *ConfigGenerator) generateScrapeConfig(
	ctx context.Context,
	sc *monitoringv1alpha1.ScrapeConfig,
	store *assets.Store,
	shards int32,
) (yaml.MapSlice, error) {
	scrapeClass := cg.getScrapeClassOrDefault(sc.Spec.ScrapeClassName)

	jobName := fmt.Sprintf("scrapeConfig/%s/%s", sc.Namespace, sc.Name)
	cfg := yaml.MapSlice{
		{
			Key:   "job_name",
			Value: jobName,
		},
	}

	cpf := cg.prom.GetCommonPrometheusFields()
	relabelings := initRelabelings()
	// Add scrape class relabelings if there is any.
	if scrapeClass != nil {
		relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)
	}
	labeler := namespacelabeler.New(cpf.EnforcedNamespaceLabel, cpf.ExcludedFromEnforcement, false)

	if sc.Spec.HonorTimestamps != nil {
		cfg = cg.AddHonorTimestamps(cfg, sc.Spec.HonorTimestamps)
	}

	if sc.Spec.TrackTimestampsStaleness != nil {
		cfg = cg.AddTrackTimestampsStaleness(cfg, sc.Spec.TrackTimestampsStaleness)
	}

	if sc.Spec.HonorLabels != nil {
		cfg = cg.AddHonorLabels(cfg, *sc.Spec.HonorLabels)
	}

	if sc.Spec.MetricsPath != nil {
		cfg = append(cfg, yaml.MapItem{Key: "metrics_path", Value: *sc.Spec.MetricsPath})
	}

	if len(sc.Spec.Params) > 0 {
		cfg = append(cfg, yaml.MapItem{Key: "params", Value: stringMapToMapSlice(sc.Spec.Params)})
	}

	if sc.Spec.EnableCompression != nil {
		cfg = cg.WithMinimumVersion("2.49.0").AppendMapItem(cfg, "enable_compression", *sc.Spec.EnableCompression)
	}

	if sc.Spec.ScrapeInterval != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_interval", Value: *sc.Spec.ScrapeInterval})
	}

	if sc.Spec.ScrapeTimeout != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_timeout", Value: *sc.Spec.ScrapeTimeout})
	}

	if len(sc.Spec.ScrapeProtocols) > 0 {
		cfg = cg.WithMinimumVersion("2.49.0").AppendMapItem(cfg, "scrape_protocols", sc.Spec.ScrapeProtocols)
	}

	if sc.Spec.Scheme != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: strings.ToLower(*sc.Spec.Scheme)})
	}

	cfg = cg.addProxyConfigtoYaml(ctx, cfg, sc.GetNamespace(), store, sc.Spec.ProxyConfig)

	cfg = cg.addBasicAuthToYaml(cfg, fmt.Sprintf("scrapeconfig/%s/%s", sc.Namespace, sc.Name), store, sc.Spec.BasicAuth)

	cfg = cg.addSafeAuthorizationToYaml(cfg, fmt.Sprintf("scrapeconfig/auth/%s/%s", sc.Namespace, sc.Name), store, sc.Spec.Authorization)

	if sc.Spec.TLSConfig != nil {
		tlsConfig := &monitoringv1.TLSConfig{SafeTLSConfig: *sc.Spec.TLSConfig}
		mergedTLSConfig := cg.MergeTLSConfigWithScrapeClass(tlsConfig, scrapeClass)
		cfg = addSafeTLStoYaml(cfg, sc.Namespace, mergedTLSConfig.SafeTLSConfig)
	}

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, sc.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, sc.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, sc.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, sc.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, sc.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, sc.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)

	if cpf.EnforcedBodySizeLimit != "" {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", cpf.EnforcedBodySizeLimit)
	}

	// StaticConfig
	if len(sc.Spec.StaticConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.StaticConfigs))
		for i, config := range sc.Spec.StaticConfigs {
			configs[i] = []yaml.MapItem{
				{
					Key:   "targets",
					Value: config.Targets,
				},
				{
					Key:   "labels",
					Value: config.Labels,
				},
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "static_configs",
			Value: configs,
		})
	}

	// FileSDConfig
	if len(sc.Spec.FileSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.FileSDConfigs))
		for i, config := range sc.Spec.FileSDConfigs {
			configs[i] = []yaml.MapItem{
				{
					Key:   "files",
					Value: config.Files,
				},
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "file_sd_configs",
			Value: configs,
		})
	}

	// HTTPSDConfig
	if len(sc.Spec.HTTPSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.HTTPSDConfigs))
		for i, config := range sc.Spec.HTTPSDConfigs {
			configs[i] = []yaml.MapItem{
				{
					Key:   "url",
					Value: config.URL,
				},
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			configs[i] = cg.addBasicAuthToYaml(configs[i], fmt.Sprintf("scrapeconfig/%s/%s/httpsdconfig/%d", sc.Namespace, sc.Name, i), store, config.BasicAuth)

			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], fmt.Sprintf("scrapeconfig/auth/%s/%s/httpsdconfig/%d", sc.Namespace, sc.Name, i), store, config.Authorization)

			if config.TLSConfig != nil {
				configs[i] = addSafeTLStoYaml(configs[i], sc.Namespace, *config.TLSConfig)
			}

			configs[i] = cg.addProxyConfigtoYaml(ctx, configs[i], sc.GetNamespace(), store, config.ProxyConfig)
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "http_sd_configs",
			Value: configs,
		})
	}

	// KubernetesSDConfig
	if len(sc.Spec.KubernetesSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.KubernetesSDConfigs))
		for i, config := range sc.Spec.KubernetesSDConfigs {
			assetStoreKey := fmt.Sprintf("scrapeconfig/%s/%s/kubernetessdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
			if config.APIServer != nil {
				configs[i] = []yaml.MapItem{
					{
						Key:   "api_server",
						Value: config.APIServer,
					},
				}
			}
			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "role",
				Value: strings.ToLower(string(config.Role)),
			})

			configs[i] = cg.addBasicAuthToYaml(configs[i], assetStoreKey, store, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], fmt.Sprintf("scrapeconfig/auth/%s/%s/kubernetessdconfig/%d", sc.GetNamespace(), sc.GetName(), i), store, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], config.OAuth2, store.OAuth2Assets, assetStoreKey)
			configs[i] = cg.addProxyConfigtoYaml(ctx, configs[i], sc.GetNamespace(), store, config.ProxyConfig)

			if config.FollowRedirects != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "follow_redirects",
					Value: config.FollowRedirects,
				})
			}

			if config.EnableHTTP2 != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "enable_http2",
					Value: config.EnableHTTP2,
				})
			}

			if config.TLSConfig != nil {
				configs[i] = addSafeTLStoYaml(configs[i], sc.GetNamespace(), *config.TLSConfig)
			}

			if config.Namespaces != nil {
				namespaces := []yaml.MapItem{
					{
						Key:   "names",
						Value: config.Namespaces.Names,
					},
				}

				if config.Namespaces.IncludeOwnNamespace != nil {
					namespaces = append(namespaces, yaml.MapItem{
						Key:   "own_namespace",
						Value: config.Namespaces.IncludeOwnNamespace,
					})
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "namespaces",
					Value: namespaces,
				})
			}

			selectors := make([][]yaml.MapItem, len(config.Selectors))
			for i, s := range config.Selectors {
				selectors[i] = []yaml.MapItem{
					{
						Key:   "role",
						Value: strings.ToLower(string(s.Role)),
					},
					{
						Key:   "label",
						Value: s.Label,
					},
					{
						Key:   "field",
						Value: s.Field,
					},
				}
			}

			if len(selectors) > 0 {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "selectors",
					Value: selectors,
				})
			}

			if config.AttachMetadata != nil {
				switch strings.ToLower(string(config.Role)) {
				case "pod":
					configs[i] = cg.WithMinimumVersion("2.35.0").AppendMapItem(configs[i], "attach_metadata", config.AttachMetadata)
				case "endpoints", "endpointslice":
					configs[i] = cg.WithMinimumVersion("2.37.0").AppendMapItem(configs[i], "attach_metadata", config.AttachMetadata)
				default:
					level.Warn(cg.logger).Log("msg", fmt.Sprintf("ignoring attachMetadata not supported by Prometheus for role: %s", config.Role))
				}
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "kubernetes_sd_configs",
			Value: configs,
		})
	}

	//ConsulSDConfig
	if len(sc.Spec.ConsulSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.ConsulSDConfigs))
		for i, config := range sc.Spec.ConsulSDConfigs {
			assetStoreKey := fmt.Sprintf("scrapeconfig/%s/%s/consulsdconfig/%d", sc.GetNamespace(), sc.GetName(), i)

			configs[i] = cg.addBasicAuthToYaml(configs[i], assetStoreKey, store, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], fmt.Sprintf("scrapeconfig/auth/%s/%s/consulsdconfig/%d", sc.GetNamespace(), sc.GetName(), i), store, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], config.Oauth2, store.OAuth2Assets, assetStoreKey)

			if config.TLSConfig != nil {
				configs[i] = addSafeTLStoYaml(configs[i], sc.GetNamespace(), *config.TLSConfig)
			}

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "server",
				Value: config.Server,
			})

			if config.TokenRef != nil {
				value, err := store.GetKey(ctx, sc.GetNamespace(), monitoringv1.SecretOrConfigMap{
					Secret: config.TokenRef,
				})

				if err != nil {
					return cfg, fmt.Errorf("failed to read %s secret %s: %w", config.TokenRef.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "token",
					Value: value,
				})
			}

			if config.Datacenter != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "datacenter",
					Value: config.Datacenter,
				})
			}

			if config.Namespace != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "namespace",
					Value: config.Namespace,
				})
			}

			if config.Partition != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "partition",
					Value: config.Partition,
				})
			}

			if config.Scheme != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "scheme",
					Value: strings.ToLower(*config.Scheme),
				})
			}

			if len(config.Services) > 0 {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "services",
					Value: config.Services,
				})
			}

			if len(config.Tags) > 0 {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "tags",
					Value: config.Tags,
				})
			}

			if config.TagSeparator != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "tag_separator",
					Value: config.TagSeparator,
				})
			}

			if len(config.NodeMeta) > 0 {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "node_meta",
					Value: stringMapToMapSlice(config.NodeMeta),
				})
			}

			if config.AllowStale != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "allow_stale",
					Value: config.AllowStale,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			configs[i] = cg.addProxyConfigtoYaml(ctx, configs[i], sc.GetNamespace(), store, config.ProxyConfig)

			if config.FollowRedirects != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "follow_redirects",
					Value: config.FollowRedirects,
				})
			}

			if config.EnableHttp2 != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "enable_http2",
					Value: config.EnableHttp2,
				})
			}
		}

		cfg = append(cfg, yaml.MapItem{
			Key:   "consul_sd_configs",
			Value: configs,
		})
	}

	// DNSSDConfig
	if len(sc.Spec.DNSSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.DNSSDConfigs))
		for i, config := range sc.Spec.DNSSDConfigs {
			configs[i] = []yaml.MapItem{
				{
					Key:   "names",
					Value: config.Names,
				},
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Type != nil {
				if *config.Type == "NS" {
					configs[i] = cg.WithMinimumVersion("2.49.0").AppendMapItem(configs[i], "type", config.Type)
				} else {
					configs[i] = append(configs[i], yaml.MapItem{
						Key:   "type",
						Value: config.Type,
					})
				}
			}

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "dns_sd_configs",
			Value: configs,
		})
	}

	// EC2SDConfig
	if len(sc.Spec.EC2SDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.EC2SDConfigs))
		for i, config := range sc.Spec.EC2SDConfigs {
			if config.Region != nil {
				configs[i] = []yaml.MapItem{
					{
						Key:   "region",
						Value: config.Region,
					},
				}
			}

			if config.AccessKey != nil && config.SecretKey != nil {

				value, err := store.GetKey(ctx, sc.GetNamespace(), monitoringv1.SecretOrConfigMap{
					Secret: config.AccessKey,
				})

				if err != nil {
					return cfg, fmt.Errorf("failed to get %s access key %s: %w", config.AccessKey.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "access_key",
					Value: value,
				})

				value, err = store.GetKey(ctx, sc.GetNamespace(), monitoringv1.SecretOrConfigMap{
					Secret: config.SecretKey,
				})

				if err != nil {
					return cfg, fmt.Errorf("failed to get %s access key %s: %w", config.SecretKey.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "secret_key",
					Value: value,
				})
			}

			if config.RoleARN != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "role_arn",
					Value: config.RoleARN,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}

			if config.Filters != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "filters",
					Value: config.Filters,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "ec2_sd_configs",
			Value: configs,
		})
	}

	// AzureSDConfig
	if len(sc.Spec.AzureSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.AzureSDConfigs))
		for i, config := range sc.Spec.AzureSDConfigs {
			if config.Environment != nil {
				configs[i] = []yaml.MapItem{
					{
						Key:   "environment",
						Value: config.Environment,
					},
				}
			}

			if config.AuthenticationMethod != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "authentication_method",
					Value: config.AuthenticationMethod,
				})
			}

			if config.SubscriptionID != "" {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "subscription_id",
					Value: config.SubscriptionID,
				})
			}

			if config.TenantID != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "tenant_id",
					Value: config.TenantID,
				})
			}

			if config.ClientID != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "client_id",
					Value: config.ClientID,
				})
			}

			if config.ClientSecret != nil {
				value, err := store.GetKey(ctx, sc.GetNamespace(), monitoringv1.SecretOrConfigMap{
					Secret: config.ClientSecret,
				})

				if err != nil {
					return cfg, fmt.Errorf("failed to get %s client secret %s: %w", config.ClientSecret.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "client_secret",
					Value: value,
				})
			}

			if config.ResourceGroup != nil {
				configs[i] = append(configs[i], yaml.MapItem{

					Key:   "resource_group",
					Value: config.ResourceGroup,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "azure_sd_configs",
			Value: configs,
		})
	}

	// GCESDConfig
	if len(sc.Spec.GCESDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.GCESDConfigs))
		for i, config := range sc.Spec.GCESDConfigs {
			configs[i] = []yaml.MapItem{
				{
					Key:   "project",
					Value: config.Project,
				},
			}

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "zone",
				Value: config.Zone,
			})

			if config.Filter != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "filter",
					Value: config.Filter,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}

			if config.TagSeparator != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "tag_separator",
					Value: config.TagSeparator,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "gce_sd_configs",
			Value: configs,
		})
	}

	// OpenStackSDConfig
	if len(sc.Spec.OpenStackSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.OpenStackSDConfigs))
		for i, config := range sc.Spec.OpenStackSDConfigs {
			configs[i] = []yaml.MapItem{
				{
					Key:   "role",
					Value: strings.ToLower(config.Role),
				},
			}

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "region",
				Value: config.Region,
			})

			if config.IdentityEndpoint != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "identity_endpoint",
					Value: config.IdentityEndpoint,
				})
			}

			if config.Username != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "username",
					Value: config.Username,
				})
			}

			if config.UserID != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "userid",
					Value: config.UserID,
				})
			}

			if config.Password != nil {
				password, err := store.GetKey(ctx, sc.GetNamespace(), monitoringv1.SecretOrConfigMap{
					Secret: config.Password,
				})

				if err != nil {
					return cfg, fmt.Errorf("failed to read %s secret %s: %w", config.Password.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "password",
					Value: password,
				})
			}

			if config.DomainName != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "domain_name",
					Value: config.DomainName,
				})
			}

			if config.DomainID != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "domain_id",
					Value: config.DomainID,
				})
			}

			if config.ProjectName != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "project_name",
					Value: config.ProjectName,
				})
			}

			if config.ProjectID != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "project_id",
					Value: config.ProjectID,
				})
			}

			if config.ApplicationCredentialName != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "application_credential_name",
					Value: config.ApplicationCredentialName,
				})
			}

			if config.ApplicationCredentialID != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "application_credential_id",
					Value: config.ApplicationCredentialID,
				})
			}

			if config.ApplicationCredentialSecret != nil {
				secret, err := store.GetKey(ctx, sc.GetNamespace(), monitoringv1.SecretOrConfigMap{
					Secret: config.ApplicationCredentialSecret,
				})

				if err != nil {
					return cfg, fmt.Errorf("failed to read %s secret %s: %w", config.ApplicationCredentialSecret.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "application_credential_secret",
					Value: secret,
				})
			}

			if config.AllTenants != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "all_tenants",
					Value: config.AllTenants,
				})
			}
			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}

			if config.Availability != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "availability",
					Value: config.Availability,
				})
			}

			if config.TLSConfig != nil {
				configs[i] = addSafeTLStoYaml(configs[i], sc.Namespace, *config.TLSConfig)
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "openstack_sd_configs",
			Value: configs,
		})
	}

	// DigitalOceanSDConfig
	if len(sc.Spec.DigitalOceanSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.DigitalOceanSDConfigs))
		for i, config := range sc.Spec.DigitalOceanSDConfigs {
			assetStoreKey := fmt.Sprintf("scrapeconfig/%s/%s/digitaloceansdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], fmt.Sprintf("scrapeconfig/auth/%s/%s/digitaloceansdconfig/%d", sc.GetNamespace(), sc.GetName(), i), store, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], config.OAuth2, store.OAuth2Assets, assetStoreKey)
			configs[i] = cg.addProxyConfigtoYaml(ctx, configs[i], sc.GetNamespace(), store, config.ProxyConfig)

			if config.FollowRedirects != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "follow_redirects",
					Value: config.FollowRedirects,
				})
			}

			if config.EnableHTTP2 != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "enable_http2",
					Value: config.EnableHTTP2,
				})
			}
			if config.TLSConfig != nil {
				configs[i] = addSafeTLStoYaml(configs[i], sc.GetNamespace(), *config.TLSConfig)
			}

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "digitalocean_sd_configs",
			Value: configs,
		})
	}

	// EurekaSDConfig
	if len(sc.Spec.EurekaSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.EurekaSDConfigs))
		for i, config := range sc.Spec.EurekaSDConfigs {
			assetStoreKey := fmt.Sprintf("scrapeconfig/%s/%s/eurekasdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
			configs[i] = cg.addBasicAuthToYaml(configs[i], assetStoreKey, store, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], fmt.Sprintf("scrapeconfig/auth/%s/%s/eurekasdconfig/%d", sc.GetNamespace(), sc.GetName(), i), store, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], config.OAuth2, store.OAuth2Assets, assetStoreKey)
			configs[i] = cg.addProxyConfigtoYaml(ctx, configs[i], sc.GetNamespace(), store, config.ProxyConfig)

			if config.FollowRedirects != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "follow_redirects",
					Value: config.FollowRedirects,
				})
			}

			if config.EnableHTTP2 != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "enable_http2",
					Value: config.EnableHTTP2,
				})
			}

			if config.TLSConfig != nil {
				configs[i] = addSafeTLStoYaml(configs[i], sc.GetNamespace(), *config.TLSConfig)
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Server != "" {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "server",
					Value: config.Server,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "eureka_sd_configs",
			Value: configs,
		})
	}

	if sc.Spec.MetricRelabelConfigs != nil {
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(labeler.GetRelabelingConfigs(sc.TypeMeta, sc.ObjectMeta, sc.Spec.MetricRelabelConfigs))})
	}

	if len(sc.Spec.RelabelConfigs) > 0 {
		relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(sc.TypeMeta, sc.ObjectMeta, sc.Spec.RelabelConfigs))...)
	}

	if shards != 1 {
		relabelings = cg.generateAddressShardingRelabelingRulesIfMissing(relabelings, shards)
	}

	// No need to check for the length because relabelings should always have
	// at least one item.
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	return cfg, nil
}

func (cg *ConfigGenerator) generateTracingConfig() (yaml.MapItem, error) {
	cfg := yaml.MapSlice{}
	objMeta := cg.prom.GetObjectMeta()

	tracingConfig := cg.prom.GetCommonPrometheusFields().TracingConfig

	cfg = append(cfg, yaml.MapItem{
		Key:   "endpoint",
		Value: tracingConfig.Endpoint,
	})

	if tracingConfig.ClientType != nil {
		cfg = append(cfg, yaml.MapItem{
			Key:   "client_type",
			Value: tracingConfig.ClientType,
		})
	}

	if tracingConfig.SamplingFraction != nil {
		cfg = append(cfg, yaml.MapItem{
			Key:   "sampling_fraction",
			Value: tracingConfig.SamplingFraction.AsApproximateFloat64(),
		})
	}

	if tracingConfig.Insecure != nil {
		cfg = append(cfg, yaml.MapItem{
			Key:   "insecure",
			Value: tracingConfig.Insecure,
		})
	}

	if len(tracingConfig.Headers) > 0 {
		headers := yaml.MapSlice{}
		for key, value := range tracingConfig.Headers {
			headers = append(headers, yaml.MapItem{
				Key:   key,
				Value: value,
			})
		}

		cfg = append(cfg, yaml.MapItem{
			Key:   "headers",
			Value: headers,
		})
	}

	if tracingConfig.Compression != nil {
		cfg = append(cfg, yaml.MapItem{
			Key:   "compression",
			Value: tracingConfig.Compression,
		})
	}

	if tracingConfig.Timeout != nil {
		cfg = append(cfg, yaml.MapItem{
			Key:   "timeout",
			Value: tracingConfig.Timeout,
		})
	}

	if tracingConfig.TLSConfig != nil {
		cfg = addTLStoYaml(cfg, objMeta.GetNamespace(), tracingConfig.TLSConfig)
	}

	return yaml.MapItem{
		Key:   "tracing",
		Value: cfg,
	}, nil
}

func validateProxyConfig(ctx context.Context, pc monitoringv1.ProxyConfig, store *assets.Store, namespace string) error {
	if reflect.ValueOf(pc).IsZero() {
		return nil
	}
	proxyFromEnvironmentDefined := ptr.Deref(pc.ProxyFromEnvironment, false)
	proxyURLDefined := ptr.Deref(pc.ProxyURL, "") != ""
	noProxyDefined := ptr.Deref(pc.NoProxy, "") != ""

	if len(pc.ProxyConnectHeader) > 0 && (!proxyFromEnvironmentDefined && !proxyURLDefined) {
		return fmt.Errorf("if proxyConnectHeader is configured, proxyUrl or proxyFromEnvironment must also be configured")
	}

	if proxyFromEnvironmentDefined && proxyURLDefined {
		return fmt.Errorf("if proxyFromEnvironment is configured, proxyUrl must not be configured")
	}

	if proxyFromEnvironmentDefined && noProxyDefined {
		return fmt.Errorf("if proxyFromEnvironment is configured, noProxy must not be configured")
	}

	if !proxyURLDefined && noProxyDefined {
		return fmt.Errorf("if noProxy is configured, proxyUrl must also be configured")
	}

	for k, v := range pc.ProxyConnectHeader {
		if _, err := store.GetSecretKey(ctx, namespace, v); err != nil {
			return fmt.Errorf("header[%s]: %w", k, err)
		}
	}

	return nil
}

func (cg *ConfigGenerator) getScrapeClassOrDefault(name *string) *monitoringv1.ScrapeClass {
	if name != nil {
		if scrapeClass, ok := cg.scrapeClasses[*name]; ok {
			return scrapeClass
		}
		return nil
	}
	if cg.defaultScrapeClassName != "" {
		if scrapeClass, ok := cg.scrapeClasses[cg.defaultScrapeClassName]; ok {
			return scrapeClass
		}
	}
	return nil
}

func getLowerByteSize(v *monitoringv1.ByteSize, cpf *monitoringv1.CommonPrometheusFields) *monitoringv1.ByteSize {
	if isByteSizeEmpty(&cpf.EnforcedBodySizeLimit) {
		return v
	}

	if isByteSizeEmpty(v) {
		return &cpf.EnforcedBodySizeLimit
	}

	vBytes, _ := units.ParseBase2Bytes(string(*v))
	pBytes, _ := units.ParseBase2Bytes(string(cpf.EnforcedBodySizeLimit))

	if vBytes > pBytes {
		return &cpf.EnforcedBodySizeLimit
	}

	return v
}

func isByteSizeEmpty(v *monitoringv1.ByteSize) bool {
	return v == nil || *v == ""
}
