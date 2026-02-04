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
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/alecthomas/units"
	"github.com/blang/semver/v4"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	sortutil "github.com/prometheus-operator/prometheus-operator/internal/sortutil"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/prometheus/validation"
)

const (
	kubernetesSDRoleEndpoint      = "endpoints"
	kubernetesSDRoleEndpointSlice = "endpointslice"
	kubernetesSDRolePod           = "pod"
	kubernetesSDRoleIngress       = "ingress"

	defaultPrometheusExternalLabelName = "prometheus"
	defaultReplicaExternalLabelName    = "prometheus_replica"

	hashLabelNameForSharding          = "__tmp_hash"
	hashLabelNameForDisablingSharding = "__tmp_disable_sharding"
)

var invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}

// ConfigGenerator knows how to generate a Prometheus configuration which is
// compatible with a given Prometheus version.
type ConfigGenerator struct {
	logger                     *slog.Logger
	version                    semver.Version
	notCompatible              bool
	prom                       monitoringv1.PrometheusInterface
	endpointSliceSupported     bool // True when the cluster supports EndpointSlice.
	scrapeClasses              map[string]monitoringv1.ScrapeClass
	defaultScrapeClassName     string
	daemonSet                  bool
	prometheusTopologySharding bool
	inlineTLSConfig            bool

	bypassVersionCheck bool
}

type ConfigGeneratorOption func(*ConfigGenerator)

func WithEndpointSliceSupport() ConfigGeneratorOption {
	return func(cg *ConfigGenerator) {
		cg.endpointSliceSupported = true
	}
}

func WithDaemonSet() ConfigGeneratorOption {
	return func(cg *ConfigGenerator) {
		cg.daemonSet = true
	}
}

func WithPrometheusTopologySharding() ConfigGeneratorOption {
	return func(cg *ConfigGenerator) {
		cg.prometheusTopologySharding = true
	}
}

func WithInlineTLSConfig() ConfigGeneratorOption {
	return func(cg *ConfigGenerator) {
		cg.inlineTLSConfig = true
	}
}

// WithoutVersionCheck returns a [ConfigGenerator] which doesn't perform any
// version check.
func WithoutVersionCheck() ConfigGeneratorOption {
	return func(cg *ConfigGenerator) {
		cg.bypassVersionCheck = true
	}
}

// NewConfigGenerator creates a ConfigGenerator for the provided Prometheus resource.
func NewConfigGenerator(
	logger *slog.Logger,
	p monitoringv1.PrometheusInterface,
	opts ...ConfigGeneratorOption,
) (*ConfigGenerator, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	cg := &ConfigGenerator{
		logger: logger,
		prom:   p,
	}

	if cg.prom == nil {
		for _, opt := range opts {
			opt(cg)
		}
		return cg, nil
	}

	cpf := p.GetCommonPrometheusFields()
	promVersion := operator.StringValOrDefault(cpf.Version, operator.DefaultPrometheusVersion)
	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus version: %w", err)
	}
	cg.version = version

	if version.Major != 2 && version.Major != 3 {
		return nil, fmt.Errorf("unsupported Prometheus version %q", promVersion)
	}

	cg.logger = logger.With("version", promVersion)

	scrapeClasses, defaultScrapeClassName, err := getScrapeClassConfig(p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scrape classes: %w", err)
	}
	cg.scrapeClasses = scrapeClasses
	cg.defaultScrapeClassName = defaultScrapeClassName

	for _, opt := range opts {
		opt(cg)
	}

	return cg, nil
}

// defaultEndpointRoleFlavor returns the default role (endpoints or
// endpointslice) to be used for Kubernetes service discovery configurations.
func (cg *ConfigGenerator) defaultEndpointRoleFlavor() string {
	return cg.endpointRoleFlavor(cg.prom.GetCommonPrometheusFields().ServiceDiscoveryRole)
}

// endpointRoleFlavor returns the Kubernetes service discovery's role
// (endpoints or endpointslice) corresponding to the given value.
func (cg *ConfigGenerator) endpointRoleFlavor(sdr *monitoringv1.ServiceDiscoveryRole) string {
	if !cg.endpointSliceSupported {
		return kubernetesSDRoleEndpoint
	}

	if cg.version.LT(semver.MustParse("2.21.0")) {
		return kubernetesSDRoleEndpoint
	}

	if ptr.Deref(sdr, monitoringv1.EndpointsRole) == monitoringv1.EndpointSliceRole {
		return kubernetesSDRoleEndpointSlice
	}

	return kubernetesSDRoleEndpoint
}

func getScrapeClassConfig(p monitoringv1.PrometheusInterface) (map[string]monitoringv1.ScrapeClass, string, error) {
	var (
		cpf                = p.GetCommonPrometheusFields()
		scrapeClasses      = make(map[string]monitoringv1.ScrapeClass, len(cpf.ScrapeClasses))
		defaultScrapeClass string
	)

	for _, scrapeClass := range cpf.ScrapeClasses {
		lcv, err := validation.NewLabelConfigValidator(p)
		if err != nil {
			return nil, "", err
		}

		if err := lcv.Validate(scrapeClass.Relabelings); err != nil {
			return nil, "", fmt.Errorf("invalid relabelings for scrapeClass %s: %w", scrapeClass.Name, err)
		}

		if err := lcv.Validate(scrapeClass.MetricRelabelings); err != nil {
			return nil, "", fmt.Errorf("invalid metric relabelings for scrapeClass %s: %w", scrapeClass.Name, err)
		}

		if err := scrapeClass.TLSConfig.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid TLS config for scrapeClass %s: %w", scrapeClass.Name, err)
		}

		if err := scrapeClass.Authorization.Validate(); err != nil {
			return nil, "", fmt.Errorf("invalid authorization for scrapeClass %s: %w", scrapeClass.Name, err)
		}

		if ptr.Deref(scrapeClass.Default, false) {
			if defaultScrapeClass != "" {
				return nil, "", fmt.Errorf("multiple default scrape classes defined")
			}

			defaultScrapeClass = scrapeClass.Name
		}

		scrapeClasses[scrapeClass.Name] = scrapeClass
	}

	return scrapeClasses, defaultScrapeClass, nil
}

// Version returns the currently configured Prometheus version.
func (cg *ConfigGenerator) Version() semver.Version {
	return cg.version
}

// WithKeyVals returns a new ConfigGenerator with the same characteristics as
// the current object, expect that the keyvals are appended to the existing
// logger.
func (cg *ConfigGenerator) WithKeyVals(keyvals ...any) *ConfigGenerator {
	return &ConfigGenerator{
		logger:                     cg.logger.With(keyvals...),
		version:                    cg.version,
		notCompatible:              cg.notCompatible,
		prom:                       cg.prom,
		endpointSliceSupported:     cg.endpointSliceSupported,
		scrapeClasses:              cg.scrapeClasses,
		defaultScrapeClassName:     cg.defaultScrapeClassName,
		daemonSet:                  cg.daemonSet,
		prometheusTopologySharding: cg.prometheusTopologySharding,
		inlineTLSConfig:            cg.inlineTLSConfig,
		bypassVersionCheck:         cg.bypassVersionCheck,
	}
}

// WithMinimumVersion returns a new ConfigGenerator that does nothing (except
// logging a warning message) if the Prometheus version is lesser than the
// given version.
// The method panics if version isn't a valid SemVer value.
func (cg *ConfigGenerator) WithMinimumVersion(version string) *ConfigGenerator {
	if cg.bypassVersionCheck {
		return cg
	}

	if cg.version.LT(semver.MustParse(version)) {
		return &ConfigGenerator{
			logger:                     cg.logger.With("minimum_version", version),
			version:                    cg.version,
			notCompatible:              true,
			prom:                       cg.prom,
			endpointSliceSupported:     cg.endpointSliceSupported,
			scrapeClasses:              cg.scrapeClasses,
			defaultScrapeClassName:     cg.defaultScrapeClassName,
			daemonSet:                  cg.daemonSet,
			prometheusTopologySharding: cg.prometheusTopologySharding,
			inlineTLSConfig:            cg.inlineTLSConfig,
			bypassVersionCheck:         cg.bypassVersionCheck,
		}
	}

	return cg
}

// WithMaximumVersion returns a new ConfigGenerator that does nothing (except
// logging a warning message) if the Prometheus version is greater than or
// equal to the given version.
// The method panics if version isn't a valid SemVer value.
func (cg *ConfigGenerator) WithMaximumVersion(version string) *ConfigGenerator {
	if cg.bypassVersionCheck {
		return cg
	}

	if cg.version.GTE(semver.MustParse(version)) {
		return &ConfigGenerator{
			logger:                     cg.logger.With("maximum_version", version),
			version:                    cg.version,
			notCompatible:              true,
			prom:                       cg.prom,
			endpointSliceSupported:     cg.endpointSliceSupported,
			scrapeClasses:              cg.scrapeClasses,
			defaultScrapeClassName:     cg.defaultScrapeClassName,
			daemonSet:                  cg.daemonSet,
			prometheusTopologySharding: cg.prometheusTopologySharding,
			inlineTLSConfig:            cg.inlineTLSConfig,
			bypassVersionCheck:         cg.bypassVersionCheck,
		}
	}

	return cg
}

// AppendMapItem appends the k/v item to the given yaml.MapSlice and returns
// the updated slice.
func (cg *ConfigGenerator) AppendMapItem(m yaml.MapSlice, k string, v any) yaml.MapSlice {
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
		cg.logger.Warn(fmt.Sprintf("ignoring command line argument %q=%q not supported by Prometheus", argument.Name, argument.Value))
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
	cg.logger.Warn(fmt.Sprintf("ignoring %q not supported by Prometheus", field))
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
	finalLimit := cg.getLimit(limit, enforcedLimit)
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

// addScrapeProtocols adds the scrape_protocols field into the configuration.
func (cg *ConfigGenerator) addScrapeProtocols(cfg yaml.MapSlice, scrapeProtocols []monitoringv1.ScrapeProtocol) yaml.MapSlice {
	if len(scrapeProtocols) == 0 {
		return cfg
	}

	sps := make([]string, 0, len(scrapeProtocols))
	for _, sp := range scrapeProtocols {
		// PrometheusText1.0.0 requires Prometheus v3.0.0 at least.
		if sp == monitoringv1.PrometheusText1_0_0 && !cg.WithMinimumVersion("3.0.0-rc.0").IsCompatible() {
			cg.Warn(fmt.Sprintf("scrapeProtocol=%s", monitoringv1.PrometheusText1_0_0))
			continue
		}

		sps = append(sps, string(sp))
	}

	return cg.WithMinimumVersion("2.49.0").AppendMapItem(cfg, "scrape_protocols", sps)
}

// addFallbackScrapeProtocol adds the fallback_scrape_protocol field into the configuration.
func (cg *ConfigGenerator) addFallbackScrapeProtocol(cfg yaml.MapSlice, fallbackScrapeProtocol *monitoringv1.ScrapeProtocol) yaml.MapSlice {
	if fallbackScrapeProtocol == nil {
		return cfg
	}

	return cg.WithMinimumVersion("3.0.0").AppendMapItem(cfg, "fallback_scrape_protocol", fallbackScrapeProtocol)
}

// AddHonorLabels adds the honor_labels field into scrape configurations.
// if OverrideHonorLabels is true then honor_labels is always false.
func (cg *ConfigGenerator) AddHonorLabels(cfg yaml.MapSlice, honorLabels bool) yaml.MapSlice {
	if cg.prom.GetCommonPrometheusFields().OverrideHonorLabels {
		honorLabels = false
	}

	return cg.AppendMapItem(cfg, "honor_labels", honorLabels)
}

// addNativeHistogramConfig adds the native histogram field into scrape configurations.
func (cg *ConfigGenerator) addNativeHistogramConfig(cfg yaml.MapSlice, nhc monitoringv1.NativeHistogramConfig) yaml.MapSlice {
	if reflect.ValueOf(nhc).IsZero() {
		return cfg
	}

	if nhc.ScrapeNativeHistograms != nil {
		cfg = cg.WithMinimumVersion("3.8.0").AppendMapItem(cfg, "scrape_native_histograms", nhc.ScrapeNativeHistograms)
	}

	if nhc.NativeHistogramBucketLimit != nil {
		cfg = cg.WithMinimumVersion("2.45.0").AppendMapItem(cfg, "native_histogram_bucket_limit", nhc.NativeHistogramBucketLimit)
	}

	if nhc.NativeHistogramMinBucketFactor != nil {
		cfg = cg.WithMinimumVersion("2.50.0").AppendMapItem(cfg, "native_histogram_min_bucket_factor", nhc.NativeHistogramMinBucketFactor.AsApproximateFloat64())
	}

	if nhc.ScrapeClassicHistograms != nil {
		switch cg.version.Major {
		case 3:
			cfg = cg.AppendMapItem(cfg, "always_scrape_classic_histograms", nhc.ScrapeClassicHistograms)
		default:
			cfg = cg.WithMinimumVersion("2.45.0").AppendMapItem(cfg, "scrape_classic_histograms", nhc.ScrapeClassicHistograms)
		}
	}

	if nhc.ConvertClassicHistogramsToNHCB != nil {
		cfg = cg.WithMinimumVersion("3.0.0").AppendMapItem(cfg, "convert_classic_histograms_to_nhcb", nhc.ConvertClassicHistogramsToNHCB)
	}

	return cfg
}

// stringMapToMapSlice returns a yaml.MapSlice from a string map to ensure that
// the output is deterministic.
func stringMapToMapSlice[V any](m map[string]V) yaml.MapSlice {
	res := yaml.MapSlice{}

	for _, k := range sortutil.SortedKeys(m) {
		res = append(res, yaml.MapItem{Key: k, Value: m[k]})
	}

	return res
}

func mergeSafeAuthorizationWithScrapeClass(authz *monitoringv1.SafeAuthorization, scrapeClass monitoringv1.ScrapeClass) *monitoringv1.Authorization {
	if authz == nil || reflect.ValueOf(*authz).IsZero() {
		return mergeAuthorizationWithScrapeClass(nil, scrapeClass)
	}

	return mergeAuthorizationWithScrapeClass(&monitoringv1.Authorization{SafeAuthorization: *authz}, scrapeClass)
}

func mergeAuthorizationWithScrapeClass(authz *monitoringv1.Authorization, scrapeClass monitoringv1.ScrapeClass) *monitoringv1.Authorization {
	if authz == nil {
		return scrapeClass.Authorization
	}

	if scrapeClass.Authorization == nil {
		return authz
	}

	if authz.Credentials == nil {
		authz.Credentials = scrapeClass.Authorization.Credentials
	}

	if authz.Credentials == nil && authz.CredentialsFile == "" {
		authz.Credentials = scrapeClass.Authorization.Credentials
		authz.CredentialsFile = scrapeClass.Authorization.CredentialsFile
	}

	return authz
}

func mergeSafeTLSConfigWithScrapeClass(tlsConfig *monitoringv1.SafeTLSConfig, scrapeClass monitoringv1.ScrapeClass) *monitoringv1.TLSConfig {
	if tlsConfig == nil || reflect.ValueOf(*tlsConfig).IsZero() {
		return mergeTLSConfigWithScrapeClass(nil, scrapeClass)
	}

	return mergeTLSConfigWithScrapeClass(&monitoringv1.TLSConfig{SafeTLSConfig: *tlsConfig}, scrapeClass)
}

func mergeTLSConfigWithScrapeClass(tlsConfig *monitoringv1.TLSConfig, scrapeClass monitoringv1.ScrapeClass) *monitoringv1.TLSConfig {
	if tlsConfig == nil {
		return scrapeClass.TLSConfig
	}

	if scrapeClass.TLSConfig == nil {
		return tlsConfig
	}

	if tlsConfig.CAFile == "" && tlsConfig.CA == (monitoringv1.SecretOrConfigMap{}) {
		tlsConfig.CAFile = scrapeClass.TLSConfig.CAFile
	}

	if tlsConfig.CertFile == "" && tlsConfig.Cert == (monitoringv1.SecretOrConfigMap{}) {
		tlsConfig.CertFile = scrapeClass.TLSConfig.CertFile
	}

	if tlsConfig.KeyFile == "" && tlsConfig.KeySecret == nil {
		tlsConfig.KeyFile = scrapeClass.TLSConfig.KeyFile
	}

	return tlsConfig
}

func mergeAttachMetadataWithScrapeClass(attachMetadata *monitoringv1.AttachMetadata, scrapeClass monitoringv1.ScrapeClass, minimumVersion string) *attachMetadataConfig {
	if attachMetadata == nil {
		attachMetadata = scrapeClass.AttachMetadata
	}

	if attachMetadata == nil {
		return nil
	}

	return &attachMetadataConfig{
		MinimumVersion: minimumVersion,
		attachMetadata: attachMetadata,
	}
}

func mergeFallbackScrapeProtocolWithScrapeClass(fallbackScrapeProtocol *monitoringv1.ScrapeProtocol, scrapeClass monitoringv1.ScrapeClass) *monitoringv1.ScrapeProtocol {
	if fallbackScrapeProtocol == nil {
		fallbackScrapeProtocol = scrapeClass.FallbackScrapeProtocol
	}

	return fallbackScrapeProtocol
}

func (cg *ConfigGenerator) addBasicAuthToYaml(
	cfg yaml.MapSlice,
	store assets.StoreGetter,
	basicAuth *monitoringv1.BasicAuth,
) yaml.MapSlice {
	if basicAuth == nil {
		return cfg
	}

	username, err := store.GetSecretKey(basicAuth.Username)
	if err != nil {
		cg.logger.Error("invalid username reference", "err", err)
	}

	password, err := store.GetSecretKey(basicAuth.Password)
	if err != nil {
		cg.logger.Error("invalid password reference", "err", err)
	}

	auth := yaml.MapSlice{
		yaml.MapItem{Key: "username", Value: string(username)},
		yaml.MapItem{Key: "password", Value: string(password)},
	}

	return cg.AppendMapItem(cfg, "basic_auth", auth)
}

func (cg *ConfigGenerator) addSigv4ToYaml(cfg yaml.MapSlice,
	assetStoreKey string,
	store assets.StoreGetter,
	sigv4 *monitoringv1.Sigv4,
) yaml.MapSlice {
	if sigv4 == nil {
		return cfg
	}

	sigv4Cfg := yaml.MapSlice{}
	if sigv4.Region != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "region", Value: sigv4.Region})
	}

	if sigv4.AccessKey != nil && sigv4.SecretKey != nil {
		var ak, sk []byte

		ak, err := store.GetSecretKey(*sigv4.AccessKey)
		if err != nil {
			cg.logger.Error("invalid SigV4 access key reference", "err", err)
		}

		sk, err = store.GetSecretKey(*sigv4.SecretKey)
		if err != nil {
			cg.logger.Error("invalid SigV4 secret key reference", "err", err)
		}

		if len(ak) > 0 && len(sk) > 0 {
			sigv4Cfg = append(sigv4Cfg,
				yaml.MapItem{Key: "access_key", Value: string(ak)},
				yaml.MapItem{Key: "secret_key", Value: string(sk)},
			)
		}
	}

	if sigv4.Profile != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "profile", Value: sigv4.Profile})
	}

	if sigv4.RoleArn != "" {
		sigv4Cfg = append(sigv4Cfg, yaml.MapItem{Key: "role_arn", Value: sigv4.RoleArn})
	}

	if sigv4.UseFIPSSTSEndpoint != nil {
		sigv4Cfg = cg.WithMinimumVersion("2.54.0").AppendMapItem(sigv4Cfg, "use_fips_sts_endpoint", *sigv4.UseFIPSSTSEndpoint)
	}

	return cg.WithKeyVals("component", strings.Split(assetStoreKey, "/")[0]).AppendMapItem(cfg, "sigv4", sigv4Cfg)
}

func (cg *ConfigGenerator) addSafeAuthorizationToYaml(
	cfg yaml.MapSlice,
	store assets.StoreGetter,
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
		b, err := store.GetSecretKey(*auth.Credentials)
		if err != nil {
			cg.logger.Error("invalid credentials reference", "err", err)
		} else {
			authCfg = append(authCfg, yaml.MapItem{Key: "credentials", Value: string(b)})
		}
	}

	return cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "authorization", authCfg)
}

func (cg *ConfigGenerator) addAuthorizationToYaml(
	cfg yaml.MapSlice,
	store assets.StoreGetter,
	auth *monitoringv1.Authorization,
) yaml.MapSlice {
	if auth == nil {
		return cfg
	}

	// reuse addSafeAuthorizationToYaml and unpack the part we're interested
	// in, namely the value under the "authorization" key
	authCfg := cg.addSafeAuthorizationToYaml(yaml.MapSlice{}, store, &auth.SafeAuthorization)[0].Value.(yaml.MapSlice)

	if auth.CredentialsFile != "" {
		authCfg = append(authCfg, yaml.MapItem{Key: "credentials_file", Value: auth.CredentialsFile})
	}

	return cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "authorization", authCfg)
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
			cg.logger.Warn("ignoring external label because it is a reserved key", "key", k)
			continue
		}
		m[k] = v
	}

	return stringMapToMapSlice(m)
}

func (cg *ConfigGenerator) addProxyConfigtoYaml(
	cfg yaml.MapSlice,
	store assets.StoreGetter,
	proxyConfig monitoringv1.ProxyConfig,
) yaml.MapSlice {
	if reflect.ValueOf(proxyConfig).IsZero() {
		return cfg
	}

	if proxyConfig.ProxyURL != nil {
		cfg = cg.AppendMapItem(cfg, "proxy_url", *proxyConfig.ProxyURL)
	}

	cgProxyConfig := cg.WithMinimumVersion("2.43.0")

	if proxyConfig.NoProxy != nil {
		cfg = cgProxyConfig.AppendMapItem(cfg, "no_proxy", *proxyConfig.NoProxy)
	}

	if proxyConfig.ProxyFromEnvironment != nil {
		cfg = cgProxyConfig.AppendMapItem(cfg, "proxy_from_environment", *proxyConfig.ProxyFromEnvironment)
	}

	if proxyConfig.ProxyConnectHeader != nil {
		proxyConnectHeader := make(map[string][]string, len(proxyConfig.ProxyConnectHeader))

		for k, v := range proxyConfig.ProxyConnectHeader {
			proxyConnectHeader[k] = []string{}
			for _, s := range v {
				value, _ := store.GetSecretKey(s)
				proxyConnectHeader[k] = append(proxyConnectHeader[k], string(value))
			}
		}

		cfg = cgProxyConfig.AppendMapItem(cfg, "proxy_connect_header", stringMapToMapSlice(proxyConnectHeader))
	}

	return cfg
}

func (cg *ConfigGenerator) addSafeTLStoYaml(
	cfg yaml.MapSlice,
	store assets.StoreGetter,
	safetls *monitoringv1.SafeTLSConfig,
) yaml.MapSlice {

	if safetls == nil {
		return cfg
	}

	safetlsConfig := yaml.MapSlice{}

	if safetls.InsecureSkipVerify != nil {
		safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "insecure_skip_verify", Value: *safetls.InsecureSkipVerify})
	}

	if safetls.CA.Secret != nil || safetls.CA.ConfigMap != nil {
		if cg.inlineTLSConfig {
			b, err := store.GetSecretOrConfigMapKey(safetls.CA)
			if err != nil {
				cg.logger.Error("invalid CA reference", "err", err)
			} else {
				safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "ca", Value: b})
			}
		} else {
			safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "ca_file", Value: path.Join(tlsAssetsDir, store.TLSAsset(safetls.CA))})
		}
	}

	if safetls.Cert.Secret != nil || safetls.Cert.ConfigMap != nil {
		if cg.inlineTLSConfig {
			b, err := store.GetSecretOrConfigMapKey(safetls.Cert)
			if err != nil {
				cg.logger.Error("invalid cert reference", "err", err)
			} else {
				safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "cert", Value: b})
			}
		} else {
			safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "cert_file", Value: path.Join(tlsAssetsDir, store.TLSAsset(safetls.Cert))})
		}
	}

	if safetls.KeySecret != nil {
		if cg.inlineTLSConfig {
			b, err := store.GetSecretKey(*safetls.KeySecret)
			if err != nil {
				cg.logger.Error("invalid key reference", "err", err)
			} else {
				safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "key", Value: string(b)})
			}
		} else {
			safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "key_file", Value: path.Join(tlsAssetsDir, store.TLSAsset(safetls.KeySecret))})
		}
	}

	if ptr.Deref(safetls.ServerName, "") != "" {
		safetlsConfig = append(safetlsConfig, yaml.MapItem{Key: "server_name", Value: *safetls.ServerName})
	}

	if safetls.MinVersion != nil {
		safetlsConfig = cg.WithMinimumVersion("2.35.0").AppendMapItem(safetlsConfig, "min_version", *safetls.MinVersion)
	}

	if safetls.MaxVersion != nil {
		safetlsConfig = cg.WithMinimumVersion("2.41.0").AppendMapItem(safetlsConfig, "max_version", *safetls.MaxVersion)
	}

	return cg.AppendMapItem(cfg, "tls_config", safetlsConfig)
}

func (cg *ConfigGenerator) addHTTPConfigToYAML(
	cfg yaml.MapSlice,
	store assets.StoreGetter,
	httpConfig *monitoringv1.HTTPConfig,
	scrapeClass monitoringv1.ScrapeClass,

) yaml.MapSlice {
	if httpConfig == nil {
		return cfg
	}

	if httpConfig.FollowRedirects != nil {
		cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "follow_redirects", *httpConfig.FollowRedirects)
	}

	if httpConfig.EnableHTTP2 != nil {
		cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *httpConfig.EnableHTTP2)
	}

	return cg.addTLStoYaml(cfg, store, mergeSafeTLSConfigWithScrapeClass(httpConfig.TLSConfig, scrapeClass))
}

func (cg *ConfigGenerator) addTLStoYaml(
	cfg yaml.MapSlice,
	store assets.StoreGetter,
	tls *monitoringv1.TLSConfig,
) yaml.MapSlice {
	if tls == nil {
		return cfg
	}

	tlsConfig := cg.addSafeTLStoYaml(yaml.MapSlice{}, store, &tls.SafeTLSConfig)[0].Value.(yaml.MapSlice)

	if tls.CAFile != "" {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "ca_file", Value: tls.CAFile})
	}

	if tls.CertFile != "" {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "cert_file", Value: tls.CertFile})
	}

	if tls.KeyFile != "" {
		tlsConfig = append(tlsConfig, yaml.MapItem{Key: "key_file", Value: tls.KeyFile})
	}

	return cg.AppendMapItem(cfg, "tls_config", tlsConfig)
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
	p *monitoringv1.Prometheus,
	sMons map[string]*monitoringv1.ServiceMonitor,
	pMons map[string]*monitoringv1.PodMonitor,
	probes map[string]*monitoringv1.Probe,
	sCons map[string]*monitoringv1alpha1.ScrapeConfig,
	store *assets.StoreBuilder,
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

	cfg := yaml.MapSlice{}

	// Global config
	globalCfg := cg.buildGlobalConfig()
	globalCfg = cg.appendEvaluationInterval(globalCfg, p.Spec.EvaluationInterval)
	globalCfg = cg.appendRuleQueryOffset(globalCfg, p.Spec.RuleQueryOffset)
	globalCfg = cg.appendQueryLogFile(globalCfg, p.Spec.QueryLogFile)
	cfg = append(cfg, yaml.MapItem{Key: "global", Value: globalCfg})

	// Runtime config
	cfg = cg.appendRuntime(cfg)

	// Rule Files config
	cfg = cg.appendRuleFiles(cfg, ruleConfigMapNames, p.Spec.RuleSelector)

	// Scrape config
	var (
		scrapeConfigs   []yaml.MapSlice
		apiserverConfig = cpf.APIServerConfig
		shards          = shardsNumber(cg.prom)
	)

	scrapeConfigs = cg.appendServiceMonitorConfigs(scrapeConfigs, sMons, apiserverConfig, store, shards)
	scrapeConfigs = cg.appendPodMonitorConfigs(scrapeConfigs, pMons, apiserverConfig, store, shards)
	scrapeConfigs = cg.appendProbeConfigs(scrapeConfigs, probes, apiserverConfig, store, shards)
	scrapeConfigs, err := cg.appendScrapeConfigs(scrapeConfigs, sCons, store, shards)
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
	cfg, err = cg.appendStorageSettingsConfig(cfg, p.Spec.Exemplars)
	if err != nil {
		return nil, fmt.Errorf("generating storage_settings configuration failed: %w", err)
	}

	s := store.ForNamespace(cg.prom.GetObjectMeta().GetNamespace())

	// Alerting config
	cfg, err = cg.appendAlertingConfig(cfg, p.Spec.Alerting, additionalAlertRelabelConfigs, additionalAlertManagerConfigs, s)
	if err != nil {
		return nil, fmt.Errorf("generating alerting configuration failed: %w", err)
	}

	// Remote write config
	if len(cpf.RemoteWrite) > 0 {
		cfg = append(cfg, cg.GenerateRemoteWriteConfig(cpf.RemoteWrite, s))
	}

	// Remote read config
	if len(p.Spec.RemoteRead) > 0 {
		cfg = append(cfg, cg.generateRemoteReadConfig(p.Spec.RemoteRead, s))
	}

	// OTLP config
	cfg, err = cg.appendOTLPConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTLP configuration: %w", err)
	}

	cfg, err = cg.appendTracingConfig(cfg, s)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tracing configuration: %w", err)
	}

	return yaml.Marshal(cfg)
}

func (cg *ConfigGenerator) appendStorageSettingsConfig(cfg yaml.MapSlice, exemplars *monitoringv1.Exemplars) (yaml.MapSlice, error) {
	var (
		storage   yaml.MapSlice
		cgStorage = cg.WithMinimumVersion("2.29.0")
		tsdb      = cg.prom.GetCommonPrometheusFields().TSDB
	)

	if exemplars != nil && exemplars.MaxSize != nil {
		storage = cgStorage.AppendMapItem(storage, "exemplars", yaml.MapSlice{
			{
				Key:   "max_exemplars",
				Value: *exemplars.MaxSize,
			},
		})
	}

	if tsdb != nil && tsdb.OutOfOrderTimeWindow != nil {
		storage = cg.WithMinimumVersion("2.39.0").AppendMapItem(storage, "tsdb", yaml.MapSlice{
			{
				Key:   "out_of_order_time_window",
				Value: *tsdb.OutOfOrderTimeWindow,
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
	store assets.StoreGetter,
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

// BuildCommonPrometheusArgs builds a slice of arguments that are common between Prometheus Server and Agent.
func (cg *ConfigGenerator) BuildCommonPrometheusArgs() []monitoringv1.Argument {
	cpf := cg.prom.GetCommonPrometheusFields()
	promArgs := []monitoringv1.Argument{
		{Name: "config.file", Value: path.Join(ConfOutDir, ConfigEnvsubstFilename)},
	}

	if cg.version.Major == 2 {
		// Add web.console.templates and web.console.libraries only if Prometheus version is v2.x.
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.console.templates", Value: "/etc/prometheus/consoles"},
			monitoringv1.Argument{Name: "web.console.libraries", Value: "/etc/prometheus/console_libraries"})
	}

	if ptr.Deref(cpf.ReloadStrategy, monitoringv1.HTTPReloadStrategyType) == monitoringv1.HTTPReloadStrategyType {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.enable-lifecycle"})
	}

	if cpf.Web != nil {
		if cpf.Web.PageTitle != nil {
			promArgs = cg.WithMinimumVersion("2.6.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "web.page-title", Value: *cpf.Web.PageTitle})
		}

		if cpf.Web.MaxConnections != nil {
			promArgs = append(promArgs, monitoringv1.Argument{Name: "web.max-connections", Value: fmt.Sprintf("%d", *cpf.Web.MaxConnections)})
		}
	}

	if cpf.EnableRemoteWriteReceiver {
		promArgs = cg.WithMinimumVersion("2.33.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "web.enable-remote-write-receiver"})
		if len(cpf.RemoteWriteReceiverMessageVersions) > 0 {
			versions := make([]string, 0, len(cpf.RemoteWriteReceiverMessageVersions))
			for _, v := range cpf.RemoteWriteReceiverMessageVersions {
				versions = append(versions, toProtobufMessageVersion(v))
			}
			promArgs = cg.WithMinimumVersion("2.54.0").AppendCommandlineArgument(
				promArgs,
				monitoringv1.Argument{
					Name:  "web.remote-write-receiver.accepted-protobuf-messages",
					Value: strings.Join(versions, ","),
				},
			)
		}
	}

	// Since metadata-wal-records is in the process of being deprecated as part of remote write v2 stabilization as described in issue.
	// Also seems to be cause some increase in resource usage overall, will stop being automatically added on prometheus 3.4.0 onwards.
	// For more context see https://github.com/prometheus-operator/prometheus-operator/issues/7889
	for _, rw := range cpf.RemoteWrite {
		if ptr.Deref(rw.MessageVersion, monitoringv1.RemoteWriteMessageVersion1_0) == monitoringv1.RemoteWriteMessageVersion2_0 {
			cg = cg.WithMinimumVersion("2.54.0")
			if cg.Version().LT(semver.MustParse("3.4.0")) {
				promArgs = cg.AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "enable-feature", Value: "metadata-wal-records"})
				break
			}
		}
	}

	// Turn on the OTLP receiver endpoint automatically if/when the OTLP config isn't empty.
	if (cpf.EnableOTLPReceiver != nil && *cpf.EnableOTLPReceiver) || (cpf.EnableOTLPReceiver == nil && cpf.OTLP != nil) {
		if cg.version.Major >= 3 {
			promArgs = cg.AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "web.enable-otlp-receiver"})
		} else {
			promArgs = cg.WithMinimumVersion("2.47.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "enable-feature", Value: "otlp-write-receiver"})
		}
	}

	if len(cpf.EnableFeatures) > 0 {
		efs := make([]string, len(cpf.EnableFeatures))
		for i := range cpf.EnableFeatures {
			efs[i] = string(cpf.EnableFeatures[i])
		}
		promArgs = cg.WithMinimumVersion("2.25.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "enable-feature", Value: strings.Join(efs, ",")})
	}

	if cpf.ExternalURL != "" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.external-url", Value: cpf.ExternalURL})
	}

	promArgs = append(promArgs, monitoringv1.Argument{Name: "web.route-prefix", Value: cpf.WebRoutePrefix()})

	if cpf.LogLevel != "" && cpf.LogLevel != "info" {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "log.level", Value: cpf.LogLevel})
	}

	if cpf.LogFormat != "" && cpf.LogFormat != "logfmt" {
		promArgs = cg.WithMinimumVersion("2.6.0").AppendCommandlineArgument(promArgs, monitoringv1.Argument{Name: "log.format", Value: cpf.LogFormat})
	}

	if cpf.ListenLocal {
		promArgs = append(promArgs, monitoringv1.Argument{Name: "web.listen-address", Value: "127.0.0.1:9090"})
	}

	return promArgs
}

func (cg *ConfigGenerator) BuildPodMetadata() (map[string]string, map[string]string) {
	podAnnotations := map[string]string{
		operator.DefaultContainerAnnotationKey: "prometheus",
	}

	podLabels := map[string]string{
		operator.ApplicationVersionLabelKey: cg.version.String(),
	}

	podMetadata := cg.prom.GetCommonPrometheusFields().PodMetadata
	if podMetadata != nil {
		maps.Copy(podLabels, podMetadata.Labels)

		maps.Copy(podAnnotations, podMetadata.Annotations)
	}

	return podAnnotations, podLabels
}

// BuildProbes returns a tuple of 3 probe definitions:
// 1. startup probe
// 2. readiness probe
// 3. liveness probe
//
// The /-/ready handler returns OK only after the TSDB initialization has
// completed. The WAL replay can take a significant time for large setups
// hence we enable the startup probe with a generous failure threshold (15
// minutes) to ensure that the readiness probe only comes into effect once
// Prometheus is effectively ready.
// We don't want to use the /-/healthy handler here because it returns OK as
// soon as the web server is started (irrespective of the WAL replay).
func (cg *ConfigGenerator) BuildProbes() (*v1.Probe, *v1.Probe, *v1.Probe) {
	readyProbeHandler := cg.buildProbeHandler("/-/ready")
	startupPeriodSeconds, startupFailureThreshold := getStatupProbePeriodSecondsAndFailureThreshold(cg.prom.GetCommonPrometheusFields().MaximumStartupDurationSeconds)

	startupProbe := &v1.Probe{
		ProbeHandler:     readyProbeHandler,
		TimeoutSeconds:   ProbeTimeoutSeconds,
		PeriodSeconds:    startupPeriodSeconds,
		FailureThreshold: startupFailureThreshold,
	}

	readinessProbe := &v1.Probe{
		ProbeHandler:     readyProbeHandler,
		TimeoutSeconds:   ProbeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}

	livenessProbe := &v1.Probe{
		ProbeHandler:     cg.buildProbeHandler("/-/healthy"),
		TimeoutSeconds:   ProbeTimeoutSeconds,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}

	return startupProbe, readinessProbe, livenessProbe
}

func (cg *ConfigGenerator) buildProbeHandler(probePath string) v1.ProbeHandler {
	cpf := cg.prom.GetCommonPrometheusFields()

	probePath = path.Clean(cpf.WebRoutePrefix() + probePath)
	handler := v1.ProbeHandler{}
	if cpf.ListenLocal {
		probeURL := url.URL{
			Scheme: "http",
			Host:   "localhost:9090",
			Path:   probePath,
		}
		handler.Exec = operator.ExecAction(probeURL.String())

		return handler
	}

	handler.HTTPGet = &v1.HTTPGetAction{
		Path: probePath,
		Port: intstr.FromString(cpf.PortName),
	}
	if cpf.Web != nil && cpf.Web.TLSConfig != nil && cg.IsCompatible() {
		handler.HTTPGet.Scheme = v1.URISchemeHTTPS
	}

	return handler
}

func getStatupProbePeriodSecondsAndFailureThreshold(maxStartupDurationSeconds *int32) (int32, int32) {
	var (
		startupPeriodSeconds    float64 = 15
		startupFailureThreshold float64 = 60
	)

	maximumStartupDurationSeconds := float64(ptr.Deref(maxStartupDurationSeconds, 0))

	if maximumStartupDurationSeconds >= 60 {
		startupFailureThreshold = math.Ceil(maximumStartupDurationSeconds / 60)
		startupPeriodSeconds = math.Ceil(maximumStartupDurationSeconds / startupFailureThreshold)
	}

	return int32(startupPeriodSeconds), int32(startupFailureThreshold)
}

func (cg *ConfigGenerator) generatePodMonitorConfig(
	m *monitoringv1.PodMonitor,
	ep monitoringv1.PodMetricsEndpoint,
	i int, apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.StoreBuilder,
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

	attachMetaConfig := mergeAttachMetadataWithScrapeClass(m.Spec.AttachMetadata, scrapeClass, "2.35.0")

	s := store.ForNamespace(m.Namespace)

	roleSelectors := []string{kubernetesSDRolePod}
	cfg = append(cfg,
		cg.generateK8SSDConfig(
			m.Spec.NamespaceSelector,
			m.Namespace,
			apiserverConfig,
			s,
			kubernetesSDRolePod,
			attachMetaConfig,
			cg.withK8SRoleSelectorConfig(m.Spec.Selector, m.Spec.SelectorMechanism, roleSelectors)))

	if ep.Interval != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_interval", Value: ep.Interval})
	}
	if ep.ScrapeTimeout != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_timeout", Value: ep.ScrapeTimeout})
	}
	if ep.Path != "" {
		cfg = append(cfg, yaml.MapItem{Key: "metrics_path", Value: ep.Path})
	}
	if ep.Params != nil {
		cfg = append(cfg, yaml.MapItem{Key: "params", Value: ep.Params})
	}
	if ep.Scheme != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: ep.Scheme.String()})
	}

	cfg = cg.addHTTPConfigToYAML(cfg, s, &ep.HTTPConfig, scrapeClass)

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if ep.BearerTokenSecret != nil && ep.BearerTokenSecret.Name != "" {
		cg.logger.Debug("'bearerTokenSecret' is deprecated, use 'authorization' instead.")

		b, err := s.GetSecretKey(*ep.HTTPConfig.BearerTokenSecret)
		if err != nil {
			cg.logger.Error("invalid bearer token secret reference", "err", err)
		} else {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: string(b)})
		}
	}

	cfg = cg.addBasicAuthToYaml(cfg, s, ep.BasicAuth)
	cfg = cg.addOAuth2ToYaml(cfg, s, ep.OAuth2)

	cfg = cg.addProxyConfigtoYaml(cfg, s, ep.ProxyConfig)

	cfg = cg.addAuthorizationToYaml(cfg, s, mergeSafeAuthorizationWithScrapeClass(ep.Authorization, scrapeClass))

	relabelings := initRelabelings()

	if ep.FilterRunning == nil || *ep.FilterRunning {
		relabelings = append(relabelings, generateRunningFilter())
	}

	// Filter targets by pods selected by the monitor.
	// Exact label matches.
	// If roleSelector is set, we don't need to add the service labels to the relabeling rules.
	if ptr.Deref(m.Spec.SelectorMechanism, monitoringv1.SelectorMechanismRelabel) == monitoringv1.SelectorMechanismRelabel {

		for _, k := range sortutil.SortedKeys(m.Spec.Selector.MatchLabels) {
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
	}

	// Filter targets based on correct port for the endpoint.
	if ptr.Deref(ep.Port, "") != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_name"}},
			{Key: "regex", Value: *ep.Port},
		})
	} else if ptr.Deref(ep.PortNumber, 0) != 0 {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_number"}},
			{Key: "regex", Value: *ep.PortNumber},
		})
	} else if ep.TargetPort != nil { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		cg.logger.Warn("'targetPort' is deprecated, use 'port' or 'portNumber' instead.")
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
				{Key: "regex", Value: ep.TargetPort.IntValue()},
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

	if ptr.Deref(ep.Port, "") != "" {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "endpoint"},
			{Key: "replacement", Value: *ep.Port},
		})
	} else if ep.TargetPort != nil && ep.TargetPort.String() != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "endpoint"},
			{Key: "replacement", Value: ep.TargetPort.String()}, //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		})
	}

	// Add scrape class relabelings if there is any.
	relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)

	labeler := namespacelabeler.New(cpf.EnforcedNamespaceLabel, cpf.ExcludedFromEnforcement, false)
	relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs))...)

	// DaemonSet mode doesn't support sharding.
	if !cg.daemonSet {
		relabelings = appendShardingRelabelingWithAddress(relabelings, shards)
	}

	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, m.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	cfg = cg.addNativeHistogramConfig(cfg, m.Spec.NativeHistogramConfig)
	cfg = cg.addScrapeProtocols(cfg, m.Spec.ScrapeProtocols)
	cfg = cg.addFallbackScrapeProtocol(cfg, mergeFallbackScrapeProtocolWithScrapeClass(m.Spec.FallbackScrapeProtocol, scrapeClass))

	if bodySizeLimit := getLowerByteSize(m.Spec.BodySizeLimit, &cpf); !isByteSizeEmpty(bodySizeLimit) {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", bodySizeLimit)
	}

	metricRelabelings := []monitoringv1.RelabelConfig{}
	metricRelabelings = append(metricRelabelings, scrapeClass.MetricRelabelings...)
	metricRelabelings = append(metricRelabelings, labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs)...)

	if len(metricRelabelings) > 0 {
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(metricRelabelings)})
	}

	return cfg
}

// generateProbeConfig builds the prometheus configuration for a probe. This function
// assumes that it will never receive a probe with empty targets, since the
// operator filters those in the validation step in SelectProbes().
func (cg *ConfigGenerator) generateProbeConfig(
	m *monitoringv1.Probe,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.StoreBuilder,
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
	if m.Spec.ProberSpec.Scheme != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: m.Spec.ProberSpec.Scheme.String()})
	}

	var paramsMapSlice yaml.MapSlice
	if m.Spec.Module != "" {
		paramsMapSlice = append(paramsMapSlice, yaml.MapItem{Key: "module", Value: []string{m.Spec.Module}})
	}

	for _, p := range m.Spec.Params {
		if m.Spec.Module != "" && p.Name == "module" {
			continue
		}

		paramsMapSlice = append(paramsMapSlice, yaml.MapItem{Key: p.Name, Value: p.Values})
	}

	if len(paramsMapSlice) != 0 {
		cfg = append(cfg, yaml.MapItem{Key: "params", Value: paramsMapSlice})
	}

	cpf := cg.prom.GetCommonPrometheusFields()
	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, m.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	cfg = cg.addNativeHistogramConfig(cfg, m.Spec.NativeHistogramConfig)
	cfg = cg.addScrapeProtocols(cfg, m.Spec.ScrapeProtocols)
	cfg = cg.addFallbackScrapeProtocol(cfg, mergeFallbackScrapeProtocolWithScrapeClass(m.Spec.FallbackScrapeProtocol, scrapeClass))

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

	s := store.ForNamespace(m.Namespace)

	cfg = cg.addProxyConfigtoYaml(cfg, s, m.Spec.ProberSpec.ProxyConfig)

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
		relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)

		// Add configured relabelings.
		xc := labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.StaticConfig.RelabelConfigs)
		relabelings = append(relabelings, generateRelabelConfig(xc)...)

	case m.Spec.Targets.Ingress != nil:
		// Generate kubernetes_sd_config section for the ingress resources.
		// Filter targets by ingresses selected by the monitor.
		// Exact label matches.
		for _, k := range sortutil.SortedKeys(m.Spec.Targets.Ingress.Selector.MatchLabels) {
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

		cfg = append(cfg, cg.generateK8SSDConfig(m.Spec.Targets.Ingress.NamespaceSelector, m.Namespace, apiserverConfig, s, kubernetesSDRoleIngress, nil))

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
		relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)

		// Add configured relabelings.
		relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.Targets.Ingress.RelabelConfigs))...)
	}

	relabelings = appendShardingRelabelingForProbes(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	cfg = cg.addTLStoYaml(cfg, s, mergeSafeTLSConfigWithScrapeClass(m.Spec.TLSConfig, scrapeClass))

	if m.Spec.BearerTokenSecret != nil { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		b, err := s.GetSecretKey(*m.Spec.BearerTokenSecret) //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if err != nil {
			cg.logger.Error("invalid bearer token reference", "err", err)
		} else {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: string(b)})
		}
	}

	cfg = cg.addBasicAuthToYaml(cfg, s, m.Spec.BasicAuth)
	cfg = cg.addOAuth2ToYaml(cfg, s, m.Spec.OAuth2)

	cfg = cg.addAuthorizationToYaml(cfg, s, mergeSafeAuthorizationWithScrapeClass(m.Spec.Authorization, scrapeClass))

	metricRelabelings := []monitoringv1.RelabelConfig{}
	metricRelabelings = append(metricRelabelings, scrapeClass.MetricRelabelings...)
	metricRelabelings = append(metricRelabelings, labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, m.Spec.MetricRelabelConfigs)...)

	if len(metricRelabelings) > 0 {
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(metricRelabelings)})
	}

	return cfg
}

func (cg *ConfigGenerator) generateServiceMonitorConfig(
	m *monitoringv1.ServiceMonitor,
	ep monitoringv1.Endpoint,
	i int,
	apiserverConfig *monitoringv1.APIServerConfig,
	store *assets.StoreBuilder,
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

	attachMetaConfig := mergeAttachMetadataWithScrapeClass(m.Spec.AttachMetadata, scrapeClass, "2.37.0")

	s := store.ForNamespace(m.Namespace)

	role := cg.defaultEndpointRoleFlavor()
	if m.Spec.ServiceDiscoveryRole != nil {
		role = cg.endpointRoleFlavor(m.Spec.ServiceDiscoveryRole)
	}
	roleSelectors := []string{role, strings.ToLower(string(monitoringv1alpha1.KubernetesRoleService))}

	cfg = append(cfg, cg.generateK8SSDConfig(
		m.Spec.NamespaceSelector,
		m.Namespace,
		apiserverConfig,
		s,
		role,
		attachMetaConfig,
		cg.withK8SRoleSelectorConfig(m.Spec.Selector, m.Spec.SelectorMechanism, roleSelectors)),
	)

	if ep.Interval != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_interval", Value: ep.Interval})
	}
	if ep.ScrapeTimeout != "" {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_timeout", Value: ep.ScrapeTimeout})
	}
	if ep.Path != "" {
		cfg = append(cfg, yaml.MapItem{Key: "metrics_path", Value: ep.Path})
	}
	if ep.Params != nil {
		cfg = append(cfg, yaml.MapItem{Key: "params", Value: ep.Params})
	}
	if ep.Scheme != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: ep.Scheme.String()})
	}
	if ep.FollowRedirects != nil {
		cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "follow_redirects", *ep.FollowRedirects)
	}
	if ep.EnableHTTP2 != nil {
		cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *ep.EnableHTTP2)
	}

	cfg = cg.addProxyConfigtoYaml(cfg, s, ep.ProxyConfig)

	cfg = cg.addOAuth2ToYaml(cfg, s, ep.OAuth2)

	cfg = cg.addTLStoYaml(cfg, s, mergeTLSConfigWithScrapeClass(ep.TLSConfig, scrapeClass))

	if ep.BearerTokenFile != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		cg.logger.Debug("'bearerTokenFile' is deprecated, use 'authorization' instead.")
		cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: ep.BearerTokenFile}) //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	}

	if ep.BearerTokenSecret != nil && ep.BearerTokenSecret.Name != "" { //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		cg.logger.Debug("'bearerTokenSecret' is deprecated, use 'authorization' instead.")

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		b, err := s.GetSecretKey(*ep.BearerTokenSecret)
		if err != nil {
			cg.logger.Error("invalid bearer token reference", "err", err)
		} else {
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: string(b)})
		}
	}

	cfg = cg.addBasicAuthToYaml(cfg, store.ForNamespace(m.Namespace), ep.BasicAuth)

	cfg = cg.addAuthorizationToYaml(cfg, s, mergeSafeAuthorizationWithScrapeClass(ep.Authorization, scrapeClass))

	relabelings := initRelabelings()

	// Filter targets by services selected by the monitor.
	// Exact label matches.
	// If roleSelector is set, we don't need to add the service labels to the relabeling rules.
	if ptr.Deref(m.Spec.SelectorMechanism, monitoringv1.SelectorMechanismRelabel) == monitoringv1.SelectorMechanismRelabel {
		for _, k := range sortutil.SortedKeys(m.Spec.Selector.MatchLabels) {
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
	}

	// Filter targets based on correct port for the endpoint.
	if ep.Port != "" {
		sourceLabels := []string{"__meta_kubernetes_endpoint_port_name"}
		if role == kubernetesSDRoleEndpointSlice {
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
				{Key: "regex", Value: ep.TargetPort.IntValue()},
			})
		}
	}

	sourceLabels := []string{"__meta_kubernetes_endpoint_address_target_kind", "__meta_kubernetes_endpoint_address_target_name"}
	if role == kubernetesSDRoleEndpointSlice {
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
	relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)

	labeler := namespacelabeler.New(cpf.EnforcedNamespaceLabel, cpf.ExcludedFromEnforcement, false)
	relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.RelabelConfigs))...)

	relabelings = appendShardingRelabelingWithAddress(relabelings, shards)
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, m.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, m.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, m.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, m.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, m.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, m.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	cfg = cg.addNativeHistogramConfig(cfg, m.Spec.NativeHistogramConfig)
	cfg = cg.addScrapeProtocols(cfg, m.Spec.ScrapeProtocols)
	cfg = cg.addFallbackScrapeProtocol(cfg, mergeFallbackScrapeProtocolWithScrapeClass(m.Spec.FallbackScrapeProtocol, scrapeClass))

	if bodySizeLimit := getLowerByteSize(m.Spec.BodySizeLimit, &cpf); !isByteSizeEmpty(bodySizeLimit) {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", bodySizeLimit)
	}

	metricRelabelings := []monitoringv1.RelabelConfig{}
	metricRelabelings = append(metricRelabelings, scrapeClass.MetricRelabelings...)
	metricRelabelings = append(metricRelabelings, labeler.GetRelabelingConfigs(m.TypeMeta, m.ObjectMeta, ep.MetricRelabelConfigs)...)

	if len(metricRelabelings) > 0 {
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(metricRelabelings)})
	}

	return cfg
}

func generateRunningFilter() yaml.MapSlice {
	return yaml.MapSlice{
		{Key: "action", Value: "drop"},
		{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_phase"}},
		{Key: "regex", Value: "(Failed|Succeeded)"},
	}
}

func (cg *ConfigGenerator) getLimit(user *uint64, enforced *uint64) *uint64 {
	if ptr.Deref(enforced, 0) == 0 {
		return user
	}

	if ptr.Deref(user, 0) == 0 {
		// With Prometheus >= 2.45.0, the limit value in the global section will always apply, hence there's no need to set the value explicitly.
		if cg.version.GTE(semver.MustParse("2.45.0")) {
			return nil
		}
		return enforced
	}

	if ptr.Deref(enforced, 0) > ptr.Deref(user, 0) {
		return user
	}

	return enforced
}

func appendShardingRelabelingWithAddress(relabelings []yaml.MapSlice, shards int32) []yaml.MapSlice {
	return appendShardingRelabelingWithLabel(relabelings, shards, "__address__")
}

func appendShardingRelabelingForProbes(relabelings []yaml.MapSlice, shards int32) []yaml.MapSlice {
	return appendShardingRelabelingWithLabel(relabelings, shards, "__param_target")
}

func (cg *ConfigGenerator) appendShardingRelabelingWithAddressIfMissing(relabelings []yaml.MapSlice, shards int32) []yaml.MapSlice {
	for i, relabeling := range relabelings {
		for _, relabelItem := range relabeling {
			if relabelItem.Key == "action" && relabelItem.Value == "hashmod" {
				cg.logger.Debug("found existing hashmod relabeling rule, skipping", "idx", i)
				return relabelings
			}
		}
	}
	return appendShardingRelabelingWithAddress(relabelings, shards)
}

func appendShardingRelabelingWithLabel(relabelings []yaml.MapSlice, shards int32, shardLabel string) []yaml.MapSlice {
	return append(relabelings,
		// Store the "shardLabel" value into the __tmp_hash label unless the
		// latter is already set.
		yaml.MapSlice{
			{Key: "source_labels", Value: []string{shardLabel, hashLabelNameForSharding}},
			{Key: "target_label", Value: hashLabelNameForSharding},
			{Key: "regex", Value: "(.+);"},
			{Key: "replacement", Value: "$1"},
			{Key: "action", Value: "replace"},
		}, yaml.MapSlice{
			{Key: "source_labels", Value: []string{hashLabelNameForSharding}},
			{Key: "target_label", Value: hashLabelNameForSharding},
			{Key: "modulus", Value: shards},
			{Key: "action", Value: "hashmod"},
		}, yaml.MapSlice{
			{Key: "source_labels", Value: []string{hashLabelNameForSharding, hashLabelNameForDisablingSharding}},
			{Key: "regex", Value: fmt.Sprintf("$(%s);|.+;.+", operator.ShardEnvVar)},
			{Key: "action", Value: "keep"},
		})
}

func generateRelabelConfig(rc []monitoringv1.RelabelConfig) []yaml.MapSlice {
	var cfg []yaml.MapSlice

	for _, c := range rc {
		relabeling := yaml.MapSlice{}

		if len(c.SourceLabels) > 0 {
			relabeling = append(relabeling, yaml.MapItem{Key: "source_labels", Value: c.SourceLabels})
		}

		if c.Separator != nil {
			relabeling = append(relabeling, yaml.MapItem{Key: "separator", Value: *c.Separator})
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

		if c.Replacement != nil {
			relabeling = append(relabeling, yaml.MapItem{Key: "replacement", Value: *c.Replacement})
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
	attachMetadata *monitoringv1.AttachMetadata
}

func (a *attachMetadataConfig) node() bool {
	return ptr.Deref(a.attachMetadata.Node, false)
}

// k8s sd config options.
type k8sSDConfigOptions func(k8sSDConfig yaml.MapSlice) yaml.MapSlice

func (cg *ConfigGenerator) withK8SRoleSelectorConfig(
	selector metav1.LabelSelector,
	selectorMechanism *monitoringv1.SelectorMechanism,
	roles []string) k8sSDConfigOptions {
	return func(k8sSDConfig yaml.MapSlice) yaml.MapSlice {
		if ptr.Deref(selectorMechanism, monitoringv1.SelectorMechanismRelabel) == monitoringv1.SelectorMechanismRelabel {
			return k8sSDConfig
		}
		return cg.generateRoleSelectorConfig(k8sSDConfig, roles, selector)
	}
}

// generateK8SSDConfig generates a kubernetes_sd_configs entry.
func (cg *ConfigGenerator) generateK8SSDConfig(
	namespaceSelector monitoringv1.NamespaceSelector,
	namespace string,
	apiserverConfig *monitoringv1.APIServerConfig,
	store assets.StoreGetter,
	role string,
	attachMetadataConfig *attachMetadataConfig,
	opts ...k8sSDConfigOptions,
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

		k8sSDConfig = cg.addBasicAuthToYaml(k8sSDConfig, store, apiserverConfig.BasicAuth)

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if apiserverConfig.BearerToken != "" {
			cg.logger.Warn("'bearerToken' is deprecated, use 'authorization' instead.")
			k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token", Value: apiserverConfig.BearerToken})
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if apiserverConfig.BearerTokenFile != "" {
			cg.logger.Debug("'bearerTokenFile' is deprecated, use 'authorization' instead.")
			k8sSDConfig = append(k8sSDConfig, yaml.MapItem{Key: "bearer_token_file", Value: apiserverConfig.BearerTokenFile})
		}

		k8sSDConfig = cg.addAuthorizationToYaml(k8sSDConfig, store, apiserverConfig.Authorization)

		k8sSDConfig = cg.addTLStoYaml(k8sSDConfig, store, apiserverConfig.TLSConfig)

		k8sSDConfig = cg.addProxyConfigtoYaml(k8sSDConfig, store, apiserverConfig.ProxyConfig)
	}

	if attachMetadataConfig != nil {
		k8sSDConfig = cg.WithMinimumVersion(attachMetadataConfig.MinimumVersion).AppendMapItem(
			k8sSDConfig,
			"attach_metadata",
			yaml.MapSlice{
				{Key: "node", Value: attachMetadataConfig.node()},
			})
	}

	// Specific configuration generated for DaemonSet mode.
	if cg.daemonSet {
		k8sSDConfig = cg.AppendMapItem(k8sSDConfig, "selectors", []yaml.MapSlice{
			{
				{
					Key:   "role",
					Value: "pod",
				},
				{
					Key:   "field",
					Value: "spec.nodeName=$(NODE_NAME)",
				},
			},
		})
	}

	for _, opt := range opts {
		k8sSDConfig = opt(k8sSDConfig)
	}

	return yaml.MapItem{
		Key: "kubernetes_sd_configs",
		Value: []yaml.MapSlice{
			k8sSDConfig,
		},
	}
}

func (cg *ConfigGenerator) generateRoleSelectorConfig(k8sSDConfig yaml.MapSlice, roles []string, selector metav1.LabelSelector) yaml.MapSlice {
	selectors := make([]yaml.MapSlice, 0, len(roles))
	labelSelector, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		// The field must have been validated by the controller beforehand.
		// If we fail here, it's a functional bug.
		panic(fmt.Errorf("failed to convert label selector to selector: %w", err))
	}

	for _, role := range roles {
		selectors = append(selectors, yaml.MapSlice{
			{Key: "role", Value: role},
			{Key: "label", Value: labelSelector.String()},
		})
	}

	for i, item := range k8sSDConfig {
		if item.Key == "selectors" {
			k8sSDConfig[i].Value = append(k8sSDConfig[i].Value.([]yaml.MapSlice), selectors...)
			return k8sSDConfig
		}
	}

	return cg.AppendMapItem(k8sSDConfig, "selectors", selectors)
}

func (cg *ConfigGenerator) generateAlertmanagerConfig(alerting *monitoringv1.AlertingSpec, apiserverConfig *monitoringv1.APIServerConfig, store assets.StoreGetter) []yaml.MapSlice {
	if alerting == nil || len(alerting.Alertmanagers) == 0 {
		return nil
	}

	alertmanagerConfigs := make([]yaml.MapSlice, 0, len(alerting.Alertmanagers))
	for i, am := range alerting.Alertmanagers {
		cfg := yaml.MapSlice{}
		if am.Scheme != nil {
			cfg = cg.AppendMapItem(cfg, "scheme", am.Scheme.String())
		}

		if am.PathPrefix != nil {
			cfg = cg.AppendMapItem(cfg, "path_prefix", am.PathPrefix)
		}

		if am.Timeout != nil {
			cfg = append(cfg, yaml.MapItem{Key: "timeout", Value: *am.Timeout})
		}

		if am.EnableHttp2 != nil {
			cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *am.EnableHttp2)
		}

		cfg = cg.addTLStoYaml(cfg, store, am.TLSConfig)

		cfg = cg.addProxyConfigtoYaml(cfg, store, am.ProxyConfig)

		ns := ptr.Deref(am.Namespace, cg.prom.GetObjectMeta().GetNamespace())
		cfg = append(cfg, cg.generateK8SSDConfig(monitoringv1.NamespaceSelector{}, ns, apiserverConfig, store, cg.defaultEndpointRoleFlavor(), nil))

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if am.BearerTokenFile != "" {
			cg.logger.Debug("'bearerTokenFile' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: am.BearerTokenFile})
		}

		cfg = cg.WithMinimumVersion("2.26.0").addBasicAuthToYaml(cfg, store, am.BasicAuth)

		cfg = cg.addSafeAuthorizationToYaml(cfg, store, am.Authorization)

		cfg = cg.WithMinimumVersion("2.48.0").addSigv4ToYaml(cfg, fmt.Sprintf("alertmanager/auth/%d", i), store, am.Sigv4)

		apiVersionCg := cg.WithMinimumVersion("2.11.0")
		if am.APIVersion != nil {
			switch monitoringv1.AlertmanagerAPIVersion(strings.ToUpper(string(*am.APIVersion))) {
			// API v1 isn't supported anymore by Prometheus v3.
			case monitoringv1.AlertmanagerAPIVersion1:
				if cg.version.Major <= 2 {
					cfg = apiVersionCg.AppendMapItem(cfg, "api_version", strings.ToLower(string(*am.APIVersion)))
				}
			case monitoringv1.AlertmanagerAPIVersion2:
				cfg = apiVersionCg.AppendMapItem(cfg, "api_version", strings.ToLower(string(*am.APIVersion)))
			}
		}

		var relabelings []yaml.MapSlice

		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "action", Value: "keep"},
			{Key: "source_labels", Value: []string{"__meta_kubernetes_service_name"}},
			{Key: "regex", Value: am.Name},
		})

		if am.Port.StrVal != "" {
			sourceLabels := []string{"__meta_kubernetes_endpoint_port_name"}
			if cg.defaultEndpointRoleFlavor() == kubernetesSDRoleEndpointSlice {
				sourceLabels = []string{"__meta_kubernetes_endpointslice_port_name"}
			}
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: sourceLabels},
				{Key: "regex", Value: am.Port.String()},
			})
		} else if am.Port.IntVal != 0 {
			relabelings = append(relabelings, yaml.MapSlice{
				{Key: "action", Value: "keep"},
				{Key: "source_labels", Value: []string{"__meta_kubernetes_pod_container_port_number"}},
				{Key: "regex", Value: am.Port.String()},
			})
		}

		if len(am.RelabelConfigs) != 0 {
			relabelings = append(relabelings, generateRelabelConfig(am.RelabelConfigs)...)
		}

		cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

		// Append alert_relabel_configs, if any, to the config
		if len(am.AlertRelabelConfigs) > 0 {
			cfg = cg.WithMinimumVersion("2.51.0").AppendMapItem(cfg, "alert_relabel_configs", generateRelabelConfig(am.AlertRelabelConfigs))
		}
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

	// DaemonSet mode doesn't support sharding.
	if cg.daemonSet || shards == 1 {
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
			values, ok := mapItem.Value.([]any)
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

		relabelings = cg.appendShardingRelabelingWithAddressIfMissing(relabelings, shards)

		addlScrapeConfig = append(addlScrapeConfig, otherConfigItems...)
		addlScrapeConfig = append(addlScrapeConfig, yaml.MapItem{Key: "relabel_configs", Value: relabelings})
		addlScrapeConfigs = append(addlScrapeConfigs, addlScrapeConfig)
	}

	return addlScrapeConfigs, nil
}

func (cg *ConfigGenerator) generateRemoteReadConfig(remoteRead []monitoringv1.RemoteReadSpec, s assets.StoreGetter) yaml.MapItem {
	cfgs := []yaml.MapSlice{}

	for _, spec := range remoteRead {
		cfg := yaml.MapSlice{
			{Key: "url", Value: spec.URL},
		}

		if spec.RemoteTimeout != nil {
			cfg = append(cfg, yaml.MapItem{Key: "remote_timeout", Value: *spec.RemoteTimeout})
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

		cfg = cg.addBasicAuthToYaml(cfg, s, spec.BasicAuth)

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerToken != "" {
			cg.logger.Warn("'bearerToken' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: spec.BearerToken})
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerTokenFile != "" {
			cg.logger.Debug("'bearerTokenFile' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: spec.BearerTokenFile})
		}

		cfg = cg.addOAuth2ToYaml(cfg, s, spec.OAuth2)

		cfg = cg.addTLStoYaml(cfg, s, spec.TLSConfig)

		cfg = cg.addAuthorizationToYaml(cfg, s, spec.Authorization)

		cfg = cg.addProxyConfigtoYaml(cfg, s, spec.ProxyConfig)

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
	store assets.StoreGetter,
	oauth2 *monitoringv1.OAuth2,
) yaml.MapSlice {
	if oauth2 == nil {
		return cfg
	}

	clientID, err := store.GetSecretOrConfigMapKey(oauth2.ClientID)
	if err != nil {
		cg.logger.Error("invalid OAuth2 client ID reference", "err", err)
		return cfg
	}

	clientSecret, err := store.GetSecretKey(oauth2.ClientSecret)
	if err != nil {
		cg.logger.Error("invalid OAuth2 client secret reference", "err", err)
		return cfg
	}

	oauth2Cfg := yaml.MapSlice{}
	oauth2Cfg = append(oauth2Cfg,
		yaml.MapItem{Key: "client_id", Value: clientID},
		yaml.MapItem{Key: "client_secret", Value: string(clientSecret)},
		yaml.MapItem{Key: "token_url", Value: oauth2.TokenURL},
	)

	if len(oauth2.Scopes) > 0 {
		oauth2Cfg = append(oauth2Cfg, yaml.MapItem{Key: "scopes", Value: oauth2.Scopes})
	}

	if len(oauth2.EndpointParams) > 0 {
		oauth2Cfg = append(oauth2Cfg, yaml.MapItem{Key: "endpoint_params", Value: oauth2.EndpointParams})
	}

	oauth2Cfg = cg.WithMinimumVersion("2.43.0").addProxyConfigtoYaml(oauth2Cfg, store, oauth2.ProxyConfig)
	oauth2Cfg = cg.WithMinimumVersion("2.43.0").addSafeTLStoYaml(oauth2Cfg, store, oauth2.TLSConfig)

	return cg.WithMinimumVersion("2.27.0").AppendMapItem(cfg, "oauth2", oauth2Cfg)
}

func toProtobufMessageVersion(mv monitoringv1.RemoteWriteMessageVersion) string {
	switch mv {
	case monitoringv1.RemoteWriteMessageVersion1_0:
		return "prometheus.WriteRequest"
	case monitoringv1.RemoteWriteMessageVersion2_0:
		return "io.prometheus.write.v2.Request"
	}

	// The API should allow only the values listed in the switch/case
	// statement but in case something goes wrong, let's return remote
	// write v1.
	return "prometheus.WriteRequest"
}

// AddRemoteWriteToStore validates the remote-write configurations and loads
// all secret/configmap references into the store.
func (cg *ConfigGenerator) AddRemoteWriteToStore(ctx context.Context, store *assets.StoreBuilder, namespace string, rws []monitoringv1.RemoteWriteSpec) error {
	for i, rw := range rws {
		if err := cg.validateRemoteWriteSpec(rw); err != nil {
			return fmt.Errorf("remoteWrite[%d]: %w", i, err)
		}

		if err := addRemoteWritesToStore(ctx, store, namespace, rw); err != nil {
			return fmt.Errorf("remoteWrite[%d]: %w", i, err)
		}
	}

	return nil
}

func (cg *ConfigGenerator) GenerateRemoteWriteConfig(rws []monitoringv1.RemoteWriteSpec, s assets.StoreGetter) yaml.MapItem {
	var cfgs []yaml.MapSlice

	for i, spec := range rws {
		cfg := yaml.MapSlice{
			{Key: "url", Value: spec.URL},
		}

		if spec.RemoteTimeout != nil {
			cfg = append(cfg, yaml.MapItem{Key: "remote_timeout", Value: *spec.RemoteTimeout})
		}

		if len(spec.Headers) > 0 {
			cfg = cg.WithMinimumVersion("2.15.0").AppendMapItem(cfg, "headers", stringMapToMapSlice(spec.Headers))
		}

		if ptr.Deref(spec.Name, "") != "" {
			cfg = cg.WithMinimumVersion("2.15.0").AppendMapItem(cfg, "name", *spec.Name)
		}

		if spec.MessageVersion != nil {
			cfg = cg.WithMinimumVersion("2.54.0").AppendMapItem(cfg, "protobuf_message", toProtobufMessageVersion(*spec.MessageVersion))
		}

		if spec.SendExemplars != nil {
			cfg = cg.WithMinimumVersion("2.27.0").AppendMapItem(cfg, "send_exemplars", spec.SendExemplars)
		}

		if spec.SendNativeHistograms != nil {
			cfg = cg.WithMinimumVersion("2.40.0").AppendMapItem(cfg, "send_native_histograms", spec.SendNativeHistograms)
		}

		var relabelings []yaml.MapSlice
		for _, c := range spec.WriteRelabelConfigs {
			var relabeling yaml.MapSlice

			if len(c.SourceLabels) > 0 {
				relabeling = append(relabeling, yaml.MapItem{Key: "source_labels", Value: c.SourceLabels})
			}

			if c.Separator != nil {
				relabeling = append(relabeling, yaml.MapItem{Key: "separator", Value: *c.Separator})
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

			if c.Replacement != nil {
				relabeling = append(relabeling, yaml.MapItem{Key: "replacement", Value: *c.Replacement})
			}

			if c.Action != "" {
				relabeling = append(relabeling, yaml.MapItem{Key: "action", Value: strings.ToLower(c.Action)})
			}

			relabelings = append(relabelings, relabeling)
		}

		if len(relabelings) > 0 {
			cfg = append(cfg, yaml.MapItem{Key: "write_relabel_configs", Value: relabelings})
		}

		cfg = cg.addBasicAuthToYaml(cfg, s, spec.BasicAuth)

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerToken != "" {
			cg.logger.Warn("'bearerToken' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token", Value: spec.BearerToken})
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if spec.BearerTokenFile != "" {
			cg.logger.Debug("'bearerTokenFile' is deprecated, use 'authorization' instead.")
			cfg = append(cfg, yaml.MapItem{Key: "bearer_token_file", Value: spec.BearerTokenFile})
		}

		cfg = cg.addOAuth2ToYaml(cfg, s, spec.OAuth2)

		cfg = cg.addTLStoYaml(cfg, s, spec.TLSConfig)

		cfg = cg.addAuthorizationToYaml(cfg, s, spec.Authorization)

		cfg = cg.addProxyConfigtoYaml(cfg, s, spec.ProxyConfig)

		cfg = cg.WithMinimumVersion("2.26.0").addSigv4ToYaml(cfg, fmt.Sprintf("remoteWrite/%d", i), s, spec.Sigv4)

		if spec.AzureAD != nil {
			azureAd := yaml.MapSlice{}

			if spec.AzureAD.ManagedIdentity != nil {
				managedIdentity := yaml.MapSlice{}
				if clientID := ptr.Deref(spec.AzureAD.ManagedIdentity.ClientID, ""); clientID != "" {
					managedIdentity = append(managedIdentity, yaml.MapItem{Key: "client_id", Value: clientID})
				}
				azureAd = append(azureAd, yaml.MapItem{Key: "managed_identity", Value: managedIdentity})
			}

			if spec.AzureAD.OAuth != nil {
				b, err := s.GetSecretKey(spec.AzureAD.OAuth.ClientSecret)
				if err != nil {
					cg.logger.Error("invalid Azure OAuth client secret", "err", err)
				} else {
					azureAd = cg.WithMinimumVersion("2.48.0").AppendMapItem(azureAd, "oauth", yaml.MapSlice{
						{Key: "client_id", Value: spec.AzureAD.OAuth.ClientID},
						{Key: "client_secret", Value: string(b)},
						{Key: "tenant_id", Value: spec.AzureAD.OAuth.TenantID},
					})
				}
			}

			if spec.AzureAD.SDK != nil {
				azureAd = cg.WithMinimumVersion("2.52.0").AppendMapItem(
					azureAd,
					"sdk",
					yaml.MapSlice{
						{Key: "tenant_id", Value: ptr.Deref(spec.AzureAD.SDK.TenantID, "")},
					})
			}

			if spec.AzureAD.WorkloadIdentity != nil {
				workloadIdentityConfig := yaml.MapSlice{
					{Key: "client_id", Value: spec.AzureAD.WorkloadIdentity.ClientID},
					{Key: "tenant_id", Value: spec.AzureAD.WorkloadIdentity.TenantID},
				}

				azureAd = cg.WithMinimumVersion("3.7.0").AppendMapItem(azureAd, "workload_identity", workloadIdentityConfig)
			}

			if spec.AzureAD.Cloud != nil {
				azureAd = append(azureAd, yaml.MapItem{Key: "cloud", Value: spec.AzureAD.Cloud})
			}

			if scope := ptr.Deref(spec.AzureAD.Scope, ""); scope != "" {
				azureAd = cg.WithMinimumVersion("3.9.0").AppendMapItem(azureAd, "scope", scope)
			}

			cfg = cg.WithMinimumVersion("2.45.0").AppendMapItem(cfg, "azuread", azureAd)
		}

		if spec.FollowRedirects != nil {
			cfg = cg.WithMinimumVersion("2.26.0").AppendMapItem(cfg, "follow_redirects", spec.FollowRedirects)
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
			if spec.MetadataConfig.MaxSamplesPerSend != nil {
				metadataConfig = cg.WithMinimumVersion("2.29.0").AppendMapItem(metadataConfig, "max_samples_per_send", *spec.MetadataConfig.MaxSamplesPerSend)
			}

			cfg = cg.WithMinimumVersion("2.23.0").AppendMapItem(cfg, "metadata_config", metadataConfig)
		}

		if spec.RoundRobinDNS != nil {
			cfg = cg.WithMinimumVersion("3.1.0").AppendMapItem(cfg, "round_robin_dns", spec.RoundRobinDNS)
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

func (cg *ConfigGenerator) appendRuntime(slice yaml.MapSlice) yaml.MapSlice {
	runtime := cg.prom.GetCommonPrometheusFields().Runtime
	if runtime == nil {
		return slice
	}
	if !cg.WithMinimumVersion("2.53.0").IsCompatible() {
		cg.Warn("runtime")
		return slice
	}

	var runtimeSlice yaml.MapSlice
	if runtime.GoGC != nil {
		runtimeSlice = append(runtimeSlice, yaml.MapItem{Key: "gogc", Value: *runtime.GoGC})
	}

	return cg.AppendMapItem(slice, "runtime", runtimeSlice)
}

func (cg *ConfigGenerator) appendEvaluationInterval(slice yaml.MapSlice, evaluationInterval monitoringv1.Duration) yaml.MapSlice {
	return append(slice, yaml.MapItem{Key: "evaluation_interval", Value: evaluationInterval})
}

func (cg *ConfigGenerator) appendGlobalLimits(slice yaml.MapSlice, limitKey string, limit *uint64, enforcedLimit *uint64) yaml.MapSlice {
	if ptr.Deref(limit, 0) > 0 {
		if ptr.Deref(enforcedLimit, 0) > 0 && *limit > *enforcedLimit {
			cg.logger.Warn(fmt.Sprintf("%q is greater than the enforced limit, using enforced limit", limitKey), "limit", *limit, "enforced_limit", *enforcedLimit)
			return cg.AppendMapItem(slice, limitKey, *enforcedLimit)
		}
		return cg.AppendMapItem(slice, limitKey, *limit)
	}

	// Use the enforced limit if no global limit is defined to ensure that scrape jobs without an explicit limit inherit the enforced limit value.
	if ptr.Deref(enforcedLimit, 0) > 0 {
		return cg.AppendMapItem(slice, limitKey, *enforcedLimit)
	}

	return slice
}

func (cg *ConfigGenerator) appendScrapeLimits(slice yaml.MapSlice) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()

	if cpf.BodySizeLimit != nil {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "body_size_limit", cpf.BodySizeLimit)
	} else if cpf.EnforcedBodySizeLimit != "" {
		slice = cg.WithMinimumVersion("2.45.0").AppendMapItem(slice, "body_size_limit", cpf.EnforcedBodySizeLimit)
	}

	slice = cg.WithMinimumVersion("2.45.0").appendGlobalLimits(slice, "sample_limit", cpf.SampleLimit, cpf.EnforcedSampleLimit)
	slice = cg.WithMinimumVersion("2.45.0").appendGlobalLimits(slice, "target_limit", cpf.TargetLimit, cpf.EnforcedTargetLimit)
	slice = cg.WithMinimumVersion("2.45.0").appendGlobalLimits(slice, "label_limit", cpf.LabelLimit, cpf.EnforcedLabelLimit)
	slice = cg.WithMinimumVersion("2.45.0").appendGlobalLimits(slice, "label_name_length_limit", cpf.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	slice = cg.WithMinimumVersion("2.45.0").appendGlobalLimits(slice, "label_value_length_limit", cpf.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	slice = cg.WithMinimumVersion("2.47.0").appendGlobalLimits(slice, "keep_dropped_targets", cpf.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	return slice
}

func (cg *ConfigGenerator) appendExternalLabels(slice yaml.MapSlice) yaml.MapSlice {
	slice = append(slice, yaml.MapItem{
		Key:   "external_labels",
		Value: cg.buildExternalLabels(),
	})

	return slice
}

func (cg *ConfigGenerator) appendRuleQueryOffset(slice yaml.MapSlice, ruleQueryOffset *monitoringv1.Duration) yaml.MapSlice {
	if ruleQueryOffset == nil {
		return slice
	}
	return cg.WithMinimumVersion("2.53.0").AppendMapItem(slice, "rule_query_offset", ruleQueryOffset)
}

func (cg *ConfigGenerator) appendQueryLogFile(slice yaml.MapSlice, queryLogFile string) yaml.MapSlice {
	if queryLogFile == "" {
		return slice
	}

	return cg.WithMinimumVersion("2.16.0").AppendMapItem(slice, "query_log_file", logFilePath(queryLogFile))
}

func (cg *ConfigGenerator) appendScrapeFailureLogFile(slice yaml.MapSlice, scrapeFailureLogFile *string) yaml.MapSlice {
	if scrapeFailureLogFile == nil {
		return slice
	}

	return cg.WithMinimumVersion("2.55.0").AppendMapItem(slice, "scrape_failure_log_file", logFilePath(*scrapeFailureLogFile))
}

func (cg *ConfigGenerator) appendRuleFiles(slice yaml.MapSlice, ruleFiles []string, ruleSelector *metav1.LabelSelector) yaml.MapSlice {
	if ruleSelector != nil {
		ruleFilePaths := []string{}
		for _, name := range ruleFiles {
			ruleFilePaths = append(ruleFilePaths, filepath.Join(RulesDir, name, "*.yaml"))
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
	store *assets.StoreBuilder,
	shards int32) []yaml.MapSlice {

	for _, identifier := range sortutil.SortedKeys(serviceMonitors) {
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
	store *assets.StoreBuilder,
	shards int32) []yaml.MapSlice {

	for _, identifier := range sortutil.SortedKeys(podMonitors) {
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
	store *assets.StoreBuilder,
	shards int32) []yaml.MapSlice {

	for _, identifier := range sortutil.SortedKeys(probes) {
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
	sMons map[string]*monitoringv1.ServiceMonitor,
	pMons map[string]*monitoringv1.PodMonitor,
	probes map[string]*monitoringv1.Probe,
	sCons map[string]*monitoringv1alpha1.ScrapeConfig,
	store *assets.StoreBuilder,
	additionalScrapeConfigs []byte,
) ([]byte, error) {
	cpf := cg.prom.GetCommonPrometheusFields()

	// validates the value of scrapeTimeout based on scrapeInterval
	if cpf.ScrapeTimeout != "" {
		if err := CompareScrapeTimeoutToScrapeInterval(cpf.ScrapeTimeout, cpf.ScrapeInterval); err != nil {
			return nil, err
		}
	}

	cfg := yaml.MapSlice{}

	// Global config
	cfg = append(cfg, yaml.MapItem{Key: "global", Value: cg.buildGlobalConfig()})

	// Runtime config
	cfg = cg.appendRuntime(cfg)

	// Scrape config
	var (
		scrapeConfigs   []yaml.MapSlice
		apiserverConfig = cpf.APIServerConfig
		shards          = shardsNumber(cg.prom)
	)

	scrapeConfigs = cg.appendPodMonitorConfigs(scrapeConfigs, pMons, apiserverConfig, store, shards)
	scrapeConfigs, err := cg.appendAdditionalScrapeConfigs(scrapeConfigs, additionalScrapeConfigs, shards)
	if err != nil {
		return nil, fmt.Errorf("generate additional scrape configs: %w", err)
	}

	// Currently, DaemonSet mode doesn't support these.
	if !cg.daemonSet {
		scrapeConfigs = cg.appendServiceMonitorConfigs(scrapeConfigs, sMons, apiserverConfig, store, shards)
		scrapeConfigs = cg.appendProbeConfigs(scrapeConfigs, probes, apiserverConfig, store, shards)
		scrapeConfigs, err = cg.appendScrapeConfigs(scrapeConfigs, sCons, store, shards)
		if err != nil {
			return nil, fmt.Errorf("generate scrape configs: %w", err)
		}
	}

	cfg = append(cfg, yaml.MapItem{
		Key:   "scrape_configs",
		Value: scrapeConfigs,
	})

	// TSDB
	tsdb := cpf.TSDB
	if tsdb != nil && tsdb.OutOfOrderTimeWindow != nil {
		var storage yaml.MapSlice
		storage = cg.AppendMapItem(storage, "tsdb", yaml.MapSlice{
			{
				Key:   "out_of_order_time_window",
				Value: *tsdb.OutOfOrderTimeWindow,
			},
		})
		cfg = cg.WithMinimumVersion("2.54.0").AppendMapItem(cfg, "storage", storage)
	}

	// Remote write config
	s := store.ForNamespace(cg.prom.GetObjectMeta().GetNamespace())
	if len(cpf.RemoteWrite) > 0 {
		cfg = append(cfg, cg.GenerateRemoteWriteConfig(cpf.RemoteWrite, s))
	}

	// OTLP config
	cfg, err = cg.appendOTLPConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTLP configuration: %w", err)
	}

	cfg, err = cg.appendTracingConfig(cfg, s)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tracing configuration: %w", err)
	}

	return yaml.Marshal(cfg)
}

func (cg *ConfigGenerator) appendScrapeConfigs(
	slices []yaml.MapSlice,
	scrapeConfigs map[string]*monitoringv1alpha1.ScrapeConfig,
	store *assets.StoreBuilder,
	shards int32) ([]yaml.MapSlice, error) {

	for _, identifier := range sortutil.SortedKeys(scrapeConfigs) {
		cfgGenerator := cg.WithKeyVals("scrapeconfig", identifier)
		scrapeConfig, err := cfgGenerator.generateScrapeConfig(scrapeConfigs[identifier], store.ForNamespace(scrapeConfigs[identifier].GetNamespace()), shards)

		if err != nil {
			return slices, err
		}

		slices = append(slices, scrapeConfig)
	}

	return slices, nil
}

func (cg *ConfigGenerator) generateScrapeConfig(
	sc *monitoringv1alpha1.ScrapeConfig,
	s assets.StoreGetter,
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
	relabelings = append(relabelings, generateRelabelConfig(scrapeClass.Relabelings)...)
	labeler := namespacelabeler.New(cpf.EnforcedNamespaceLabel, cpf.ExcludedFromEnforcement, false)

	if sc.Spec.JobName != nil {
		relabelings = append(relabelings, yaml.MapSlice{
			{Key: "target_label", Value: "job"},
			{Key: "action", Value: "replace"},
			{Key: "replacement", Value: sc.Spec.JobName},
		})
	}

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

	if sc.Spec.EnableHTTP2 != nil {
		cfg = cg.WithMinimumVersion("2.35.0").AppendMapItem(cfg, "enable_http2", *sc.Spec.EnableHTTP2)
	}

	if sc.Spec.ScrapeInterval != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_interval", Value: *sc.Spec.ScrapeInterval})
	}

	if sc.Spec.ScrapeTimeout != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scrape_timeout", Value: *sc.Spec.ScrapeTimeout})
	}

	cfg = cg.addScrapeProtocols(cfg, sc.Spec.ScrapeProtocols)
	cfg = cg.addFallbackScrapeProtocol(cfg, mergeFallbackScrapeProtocolWithScrapeClass(sc.Spec.FallbackScrapeProtocol, scrapeClass))

	if sc.Spec.Scheme != nil {
		cfg = append(cfg, yaml.MapItem{Key: "scheme", Value: sc.Spec.Scheme.String()})
	}

	cfg = cg.addProxyConfigtoYaml(cfg, s, sc.Spec.ProxyConfig)

	cfg = cg.addBasicAuthToYaml(cfg, s, sc.Spec.BasicAuth)

	cfg = cg.addAuthorizationToYaml(cfg, s, mergeSafeAuthorizationWithScrapeClass(sc.Spec.Authorization, scrapeClass))

	cfg = cg.addOAuth2ToYaml(cfg, s, sc.Spec.OAuth2)

	cfg = cg.addTLStoYaml(cfg, s, mergeSafeTLSConfigWithScrapeClass(sc.Spec.TLSConfig, scrapeClass))

	cfg = cg.AddLimitsToYAML(cfg, sampleLimitKey, sc.Spec.SampleLimit, cpf.EnforcedSampleLimit)
	cfg = cg.AddLimitsToYAML(cfg, targetLimitKey, sc.Spec.TargetLimit, cpf.EnforcedTargetLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelLimitKey, sc.Spec.LabelLimit, cpf.EnforcedLabelLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelNameLengthLimitKey, sc.Spec.LabelNameLengthLimit, cpf.EnforcedLabelNameLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, labelValueLengthLimitKey, sc.Spec.LabelValueLengthLimit, cpf.EnforcedLabelValueLengthLimit)
	cfg = cg.AddLimitsToYAML(cfg, keepDroppedTargetsKey, sc.Spec.KeepDroppedTargets, cpf.EnforcedKeepDroppedTargets)
	cfg = cg.addNativeHistogramConfig(cfg, sc.Spec.NativeHistogramConfig)

	if bodySizeLimit := getLowerByteSize(sc.Spec.BodySizeLimit, &cpf); !isByteSizeEmpty(bodySizeLimit) {
		cfg = cg.WithMinimumVersion("2.28.0").AppendMapItem(cfg, "body_size_limit", bodySizeLimit)
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
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "url",
				Value: config.URL,
			})

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

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
			if config.APIServer != nil {
				configs[i] = []yaml.MapItem{
					{
						Key:   "api_server",
						Value: config.APIServer,
					},
				}
			}

			switch config.Role {
			case monitoringv1alpha1.KubernetesRoleEndpointSlice:
				configs[i] = cg.WithMinimumVersion("2.21.0").AppendMapItem(configs[i], "role", strings.ToLower(string(config.Role)))
			default:
				configs[i] = cg.AppendMapItem(configs[i], "role", strings.ToLower(string(config.Role)))
			}

			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

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

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

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

			if len(config.Selectors) > 0 {
				selectors := make([][]yaml.MapItem, len(config.Selectors))
				for i, s := range config.Selectors {
					selectors[i] = cg.AppendMapItem(selectors[i], "role", strings.ToLower(string(s.Role)))

					if s.Label != nil {
						selectors[i] = cg.AppendMapItem(selectors[i], "label", *s.Label)
					}

					if s.Field != nil {
						selectors[i] = cg.AppendMapItem(selectors[i], "field", *s.Field)
					}
				}

				configs[i] = cg.WithMinimumVersion("2.17.0").AppendMapItem(configs[i], "selectors", selectors)
			}

			if config.AttachMetadata != nil {
				switch strings.ToLower(string(config.Role)) {
				case "pod":
					configs[i] = cg.WithMinimumVersion("2.35.0").AppendMapItem(configs[i], "attach_metadata", config.AttachMetadata)
				case "endpoints", "endpointslice":
					configs[i] = cg.WithMinimumVersion("2.37.0").AppendMapItem(configs[i], "attach_metadata", config.AttachMetadata)
				default:
					cg.logger.Warn(fmt.Sprintf("ignoring attachMetadata not supported by Prometheus for role: %s", config.Role))
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
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "server",
				Value: config.Server,
			})

			if config.PathPrefix != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "path_prefix",
					Value: config.PathPrefix,
				})
			}

			if config.TokenRef != nil {
				value, err := s.GetSecretKey(*config.TokenRef)
				if err != nil {
					return cfg, fmt.Errorf("failed to read %s secret %s: %w", config.TokenRef.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "token",
					Value: string(value),
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
					Value: config.Scheme.String(),
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

			if config.Filter != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "filter",
					Value: config.Filter,
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

		compatibilityMatrix := map[monitoringv1alpha1.DNSRecordType]string{
			monitoringv1alpha1.DNSRecordTypeNS: "2.49.0",
			monitoringv1alpha1.DNSRecordTypeMX: "2.38.0",
		}

		for i, config := range sc.Spec.DNSSDConfigs {
			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "names",
				Value: config.Names,
			})

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Type != nil {
				typecg := cg

				if minVersion, found := compatibilityMatrix[*config.Type]; found {
					typecg = typecg.WithMinimumVersion(minVersion)
				}

				configs[i] = typecg.AppendMapItem(configs[i], "type", config.Type)
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
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			if config.Region != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "region",
					Value: config.Region,
				})
			}

			if config.AccessKey != nil && config.SecretKey != nil {

				value, err := s.GetSecretKey(*config.AccessKey)
				if err != nil {
					return cfg, fmt.Errorf("failed to get %s access key %s: %w", config.AccessKey.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "access_key",
					Value: string(value),
				})

				value, err = s.GetSecretKey(*config.SecretKey)
				if err != nil {
					return cfg, fmt.Errorf("failed to get %s access key %s: %w", config.SecretKey.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "secret_key",
					Value: string(value),
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

			configs[i] = cg.WithMinimumVersion("2.3.0").addFiltersToYaml(configs[i], config.Filters)

			cgForHTTPClientConfig := cg.WithMinimumVersion("2.41.0")

			if config.FollowRedirects != nil {
				configs[i] = cgForHTTPClientConfig.AppendMapItem(configs[i], "follow_redirects", config.FollowRedirects)
			}

			if config.EnableHTTP2 != nil {
				configs[i] = cgForHTTPClientConfig.AppendMapItem(configs[i], "enable_http2", config.EnableHTTP2)
			}

			if config.TLSConfig != nil {
				configs[i] = cgForHTTPClientConfig.addSafeTLStoYaml(configs[i], s, config.TLSConfig)
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
				value, err := s.GetSecretKey(*config.ClientSecret)
				if err != nil {
					return cfg, fmt.Errorf("failed to get %s client secret %s: %w", config.ClientSecret.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "client_secret",
					Value: string(value),
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
					Value: strings.ToLower(string(config.Role)),
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
				password, err := s.GetSecretKey(*config.Password)
				if err != nil {
					return cfg, fmt.Errorf("failed to read %s secret %s: %w", config.Password.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "password",
					Value: string(password),
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
				secret, err := s.GetSecretKey(*config.ApplicationCredentialSecret)
				if err != nil {
					return cfg, fmt.Errorf("failed to read %s secret %s: %w", config.ApplicationCredentialSecret.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "application_credential_secret",
					Value: string(secret),
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

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)
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
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

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

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

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

	// KumaSDConfig
	if len(sc.Spec.KumaSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.KumaSDConfigs))
		for i, config := range sc.Spec.KumaSDConfigs {
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "server",
				Value: config.Server,
			})

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

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.FetchTimeout != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "fetch_timeout",
					Value: config.FetchTimeout,
				})
			}

			if config.ClientID != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "client_id",
					Value: config.ClientID,
				})
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "kuma_sd_configs",
			Value: configs,
		})
	}

	// EurekaSDConfig
	if len(sc.Spec.EurekaSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.EurekaSDConfigs))
		for i, config := range sc.Spec.EurekaSDConfigs {
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

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

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

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

	// DockerSDConfig
	if len(sc.Spec.DockerSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.DockerSDConfigs))

		for i, config := range sc.Spec.DockerSDConfigs {
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addFiltersToYaml(configs[i], config.Filters)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "host",
				Value: config.Host,
			})

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}

			if config.HostNetworkingHost != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "host_networking_host",
					Value: config.HostNetworkingHost})
			}

			if config.MatchFirstNetwork != nil {
				// ref: https://github.com/prometheus/prometheus/pull/14654
				configs[i] = cg.WithMinimumVersion("2.54.1").AppendMapItem(configs[i],
					"match_first_network",
					config.MatchFirstNetwork)
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

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

		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "docker_sd_configs",
			Value: configs,
		})
	}

	// LinodeSDConfig
	if len(sc.Spec.LinodeSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.LinodeSDConfigs))

		for i, config := range sc.Spec.LinodeSDConfigs {
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}

			if config.Region != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "region",
					Value: config.Region,
				})
			}

			if config.EnableHTTP2 != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "enable_http2",
					Value: config.EnableHTTP2,
				})
			}

			if config.TagSeparator != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "tag_separator",
					Value: config.TagSeparator,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.FollowRedirects != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "follow_redirects",
					Value: config.FollowRedirects,
				})
			}

		}

		cfg = append(cfg, yaml.MapItem{
			Key:   "linode_sd_configs",
			Value: configs,
		})

	}

	// HetznerSDConfig
	if len(sc.Spec.HetznerSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.HetznerSDConfigs))
		for i, config := range sc.Spec.HetznerSDConfigs {
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "role",
				Value: strings.ToLower(config.Role),
			})

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

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

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

			if config.LabelSelector != nil && len(*config.LabelSelector) > 0 {
				configs[i] = cg.WithMinimumVersion("3.5.0").AppendMapItem(configs[i],
					"label_selector",
					config.LabelSelector)
			}
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "hetzner_sd_configs",
			Value: configs,
		})
	}

	// NomadSDConfig
	if len(sc.Spec.NomadSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.NomadSDConfigs))
		for i, config := range sc.Spec.NomadSDConfigs {
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "server",
				Value: config.Server,
			})

			if config.AllowStale != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "allow_stale",
					Value: config.AllowStale,
				})
			}

			if config.Namespace != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "namespace",
					Value: config.Namespace,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			if config.Region != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "region",
					Value: config.Region,
				})
			}

			if config.TagSeparator != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "tag_separator",
					Value: config.TagSeparator,
				})
			}

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

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "nomad_sd_configs",
			Value: configs,
		})
	}

	// DockerswarmSDConfig
	if len(sc.Spec.DockerSwarmSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.DockerSwarmSDConfigs))
		for i, config := range sc.Spec.DockerSwarmSDConfigs {
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)
			configs[i] = cg.addFiltersToYaml(configs[i], config.Filters)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "host",
				Value: config.Host,
			})

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "role",
				Value: strings.ToLower(config.Role),
			})

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

		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "dockerswarm_sd_configs",
			Value: configs,
		})
	}

	// PuppetDBSDConfig
	if len(sc.Spec.PuppetDBSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.PuppetDBSDConfigs))
		for i, config := range sc.Spec.PuppetDBSDConfigs {
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "url",
				Value: config.URL,
			})

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "query",
				Value: config.Query,
			})

			if config.IncludeParameters != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "include_parameters",
					Value: config.IncludeParameters,
				})
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

			if config.FollowRedirects != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "follow_redirects",
					Value: config.FollowRedirects,
				})
			}

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

			if config.EnableHTTP2 != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "enable_http2",
					Value: config.EnableHTTP2,
				})
			}
		}

		cfg = append(cfg, yaml.MapItem{
			Key:   "puppetdb_sd_configs",
			Value: configs,
		})
	}

	if len(sc.Spec.LightSailSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.LightSailSDConfigs))
		for i, config := range sc.Spec.LightSailSDConfigs {
			configs[i] = cg.addBasicAuthToYaml(configs[i], s, config.BasicAuth)
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, config.Authorization)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)
			configs[i] = cg.addOAuth2ToYaml(configs[i], s, config.OAuth2)

			if config.Region != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "region",
					Value: config.Region,
				})
			}

			if config.Endpoint != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "endpoint",
					Value: config.Endpoint,
				})
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

			if config.AccessKey != nil && config.SecretKey != nil {

				value, err := s.GetSecretKey(*config.AccessKey)
				if err != nil {
					return cfg, fmt.Errorf("failed to get %s access key %s: %w", config.AccessKey.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "access_key",
					Value: string(value),
				})

				value, err = s.GetSecretKey(*config.SecretKey)
				if err != nil {
					return cfg, fmt.Errorf("failed to get %s access key %s: %w", config.SecretKey.Name, jobName, err)
				}

				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "secret_key",
					Value: string(value),
				})
			}

			if config.RoleARN != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "role_arn",
					Value: config.RoleARN,
				})
			}

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

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
		}
		cfg = append(cfg, yaml.MapItem{
			Key:   "lightsail_sd_configs",
			Value: configs,
		})
	}

	// OVHCloudSDConfigs
	if len(sc.Spec.OVHCloudSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.OVHCloudSDConfigs))
		for i, config := range sc.Spec.OVHCloudSDConfigs {
			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "application_key",
				Value: config.ApplicationKey,
			})

			value, _ := s.GetSecretKey(config.ApplicationSecret)
			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "application_secret",
				Value: string(value),
			})

			key, _ := s.GetSecretKey(config.ConsumerKey)
			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "consumer_key",
				Value: string(key),
			})

			switch config.Service {
			case monitoringv1alpha1.OVHServiceVPS:
				configs[i] = append(configs[i], yaml.MapItem{Key: "service", Value: "vps"})
			case monitoringv1alpha1.OVHServiceDedicatedServer:
				configs[i] = append(configs[i], yaml.MapItem{Key: "service", Value: "dedicated_server"})
			default:
				cg.logger.Warn(fmt.Sprintf("ignoring service not supported by Prometheus: %s", string(config.Service)))
			}

			if config.Endpoint != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "endpoint",
					Value: *config.Endpoint,
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
			Key:   "ovhcloud_sd_configs",
			Value: configs,
		})
	}

	// ScalewaySDConfig
	if len(sc.Spec.ScalewaySDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.ScalewaySDConfigs))
		for i, config := range sc.Spec.ScalewaySDConfigs {
			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "access_key",
				Value: config.AccessKey,
			})

			value, _ := s.GetSecretKey(config.SecretKey)
			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "secret_key",
				Value: string(value),
			})

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "project_id",
				Value: config.ProjectID,
			})

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "role",
				Value: strings.ToLower(string(config.Role)),
			})

			if config.Port != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "port",
					Value: config.Port,
				})
			}

			if config.ApiURL != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "api_url",
					Value: *config.ApiURL,
				})
			}

			if config.Zone != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "zone",
					Value: config.Zone,
				})
			}

			if config.NameFilter != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "name_filter",
					Value: config.NameFilter,
				})
			}

			if len(config.TagsFilter) > 0 {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "tags_filter",
					Value: config.TagsFilter,
				})
			}

			if config.RefreshInterval != nil {
				configs[i] = append(configs[i], yaml.MapItem{
					Key:   "refresh_interval",
					Value: config.RefreshInterval,
				})
			}

			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)

			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

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
		}

		cfg = append(cfg, yaml.MapItem{
			Key:   "scaleway_sd_configs",
			Value: configs,
		})
	}

	// IonosSDConfig
	if len(sc.Spec.IonosSDConfigs) > 0 {
		configs := make([][]yaml.MapItem, len(sc.Spec.IonosSDConfigs))
		for i, config := range sc.Spec.IonosSDConfigs {
			configs[i] = cg.addSafeAuthorizationToYaml(configs[i], s, &config.Authorization)
			configs[i] = cg.addProxyConfigtoYaml(configs[i], s, config.ProxyConfig)
			configs[i] = cg.addSafeTLStoYaml(configs[i], s, config.TLSConfig)

			configs[i] = append(configs[i], yaml.MapItem{
				Key:   "datacenter_id",
				Value: config.DataCenterID,
			})

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
			Key:   "ionos_sd_configs",
			Value: configs,
		})
	}

	if len(sc.Spec.RelabelConfigs) > 0 {
		relabelings = append(relabelings, generateRelabelConfig(labeler.GetRelabelingConfigs(sc.TypeMeta, sc.ObjectMeta, sc.Spec.RelabelConfigs))...)
	}

	if shards != 1 {
		relabelings = cg.appendShardingRelabelingWithAddressIfMissing(relabelings, shards)
	}

	// No need to check for the length because relabelings should always have
	// at least one item.
	cfg = append(cfg, yaml.MapItem{Key: "relabel_configs", Value: relabelings})

	metricRelabelings := []monitoringv1.RelabelConfig{}
	metricRelabelings = append(metricRelabelings, scrapeClass.MetricRelabelings...)
	metricRelabelings = append(metricRelabelings, labeler.GetRelabelingConfigs(sc.TypeMeta, sc.ObjectMeta, sc.Spec.MetricRelabelConfigs)...)

	if len(metricRelabelings) > 0 {
		cfg = append(cfg, yaml.MapItem{Key: "metric_relabel_configs", Value: generateRelabelConfig(metricRelabelings)})
	}

	cfg = cg.appendNameValidationScheme(cfg, sc.Spec.NameValidationScheme)
	cfg = cg.appendNameEscapingScheme(cfg, sc.Spec.NameEscapingScheme)

	return cfg, nil
}

func (cg *ConfigGenerator) appendOTLPConfig(cfg yaml.MapSlice) (yaml.MapSlice, error) {
	otlpConfig := cg.prom.GetCommonPrometheusFields().OTLP
	nameValidationScheme := cg.prom.GetCommonPrometheusFields().NameValidationScheme

	if otlpConfig == nil {
		return cfg, nil
	}

	if cg.version.LT(semver.MustParse("2.55.0")) {
		return cfg, fmt.Errorf("OTLP configuration is only supported from Prometheus version 2.55.0")
	}

	if ptr.Deref(otlpConfig.TranslationStrategy, "") == monitoringv1.NoUTF8EscapingWithSuffixes && ptr.Deref(nameValidationScheme, "") == monitoringv1.LegacyNameValidationScheme {
		return cfg, fmt.Errorf("nameValidationScheme %q is not compatible with OTLP translation strategy %q", monitoringv1.LegacyNameValidationScheme, monitoringv1.NoUTF8EscapingWithSuffixes)
	}

	if cg.version.LT(semver.MustParse("3.4.0")) && ptr.Deref(otlpConfig.TranslationStrategy, "") == monitoringv1.NoTranslation {
		return cfg, fmt.Errorf("nameValidationScheme %q is only supported from Prometheus version 3.4.0 ", monitoringv1.NoTranslation)
	}

	if cg.version.LT(semver.MustParse("3.6.0")) && ptr.Deref(otlpConfig.TranslationStrategy, "") == monitoringv1.UnderscoreEscapingWithoutSuffixes {
		return cfg, fmt.Errorf("nameValidationScheme %q is only supported from Prometheus version 3.6.0 ", monitoringv1.UnderscoreEscapingWithoutSuffixes)
	}

	if cg.version.GTE(semver.MustParse("3.5.0")) {
		err := otlpConfig.Validate()
		if err != nil {
			return cfg, err
		}
	}

	otlp := yaml.MapSlice{}

	if len(otlpConfig.PromoteResourceAttributes) > 0 {
		otlp = cg.WithMinimumVersion("2.55.0").AppendMapItem(otlp,
			"promote_resource_attributes",
			otlpConfig.PromoteResourceAttributes)
	}

	if otlpConfig.TranslationStrategy != nil {
		otlp = cg.WithMinimumVersion("3.0.0").AppendMapItem(otlp,
			"translation_strategy",
			otlpConfig.TranslationStrategy)
	}

	if otlpConfig.KeepIdentifyingResourceAttributes != nil {
		otlp = cg.WithMinimumVersion("3.1.0").AppendMapItem(otlp,
			"keep_identifying_resource_attributes",
			otlpConfig.KeepIdentifyingResourceAttributes)
	}

	if otlpConfig.ConvertHistogramsToNHCB != nil {
		otlp = cg.WithMinimumVersion("3.4.0").AppendMapItem(otlp,
			"convert_histograms_to_nhcb",
			otlpConfig.ConvertHistogramsToNHCB)
	}

	if otlpConfig.PromoteAllResourceAttributes != nil {
		otlp = cg.WithMinimumVersion("3.5.0").AppendMapItem(otlp,
			"promote_all_resource_attributes",
			otlpConfig.PromoteAllResourceAttributes)
	}

	if len(otlpConfig.IgnoreResourceAttributes) > 0 {
		otlp = cg.WithMinimumVersion("3.5.0").AppendMapItem(otlp,
			"ignore_resource_attributes",
			otlpConfig.IgnoreResourceAttributes)
	}

	if otlpConfig.PromoteScopeMetadata != nil {
		otlp = cg.WithMinimumVersion("3.6.0").AppendMapItem(otlp,
			"promote_scope_metadata",
			otlpConfig.PromoteScopeMetadata)
	}

	if len(otlp) == 0 {
		return cfg, nil
	}

	return cg.AppendMapItem(cfg, "otlp", otlp), nil
}

func (cg *ConfigGenerator) appendTracingConfig(cfg yaml.MapSlice, s assets.StoreGetter) (yaml.MapSlice, error) {
	tracingConfig := cg.prom.GetCommonPrometheusFields().TracingConfig
	if tracingConfig == nil {
		return cfg, nil
	}

	err := tracingConfig.Validate()
	if err != nil {
		return cfg, err
	}

	var tracing yaml.MapSlice
	tracing = append(tracing, yaml.MapItem{
		Key:   "endpoint",
		Value: tracingConfig.Endpoint,
	})

	if tracingConfig.ClientType != nil {
		tracing = append(tracing, yaml.MapItem{
			Key:   "client_type",
			Value: strings.ToLower(*tracingConfig.ClientType),
		})
	}

	if tracingConfig.SamplingFraction != nil {
		tracing = append(tracing, yaml.MapItem{
			Key:   "sampling_fraction",
			Value: tracingConfig.SamplingFraction.AsApproximateFloat64(),
		})
	}

	if tracingConfig.Insecure != nil {
		tracing = append(tracing, yaml.MapItem{
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

		tracing = append(tracing, yaml.MapItem{
			Key:   "headers",
			Value: headers,
		})
	}

	if tracingConfig.Compression != nil {
		tracing = append(tracing, yaml.MapItem{
			Key:   "compression",
			Value: strings.ToLower(*tracingConfig.Compression),
		})
	}

	if tracingConfig.Timeout != nil {
		tracing = append(tracing, yaml.MapItem{
			Key:   "timeout",
			Value: tracingConfig.Timeout,
		})
	}

	tracing = cg.addTLStoYaml(tracing, s, tracingConfig.TLSConfig)

	return append(
		cfg,
		yaml.MapItem{
			Key:   "tracing",
			Value: tracing,
		}), nil
}

func (cg *ConfigGenerator) appendNameValidationScheme(cfg yaml.MapSlice, nameValidationScheme *monitoringv1.NameValidationSchemeOptions) yaml.MapSlice {
	if nameValidationScheme == nil {
		return cfg
	}

	// need to cast it to a string in order to use strings.ToLower() to render the value in the way prometheus expects it
	nameValidationSchemeValue := string(*nameValidationScheme)

	return cg.WithMinimumVersion("3.0.0").AppendMapItem(cfg, "metric_name_validation_scheme", strings.ToLower(nameValidationSchemeValue))
}

func (cg *ConfigGenerator) appendNameEscapingScheme(cfg yaml.MapSlice, nameEscapingScheme *monitoringv1.NameEscapingSchemeOptions) yaml.MapSlice {
	if nameEscapingScheme == nil {
		return cfg
	}

	// conversion to prometheus values.
	nameMap := map[monitoringv1.NameEscapingSchemeOptions]string{
		monitoringv1.AllowUTF8NameEscapingScheme:   "allow-utf-8",
		monitoringv1.UnderscoresNameEscapingScheme: "underscores",
		monitoringv1.DotsNameEscapingScheme:        "dots",
		monitoringv1.ValuesNameEscapingScheme:      "values",
	}

	if v, ok := nameMap[*nameEscapingScheme]; ok {
		return cg.WithMinimumVersion("3.4.0").AppendMapItem(cfg, "metric_name_escaping_scheme", v)
	}

	return cfg
}

func (cg *ConfigGenerator) appendConvertClassicHistogramsToNHCB(cfg yaml.MapSlice) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()

	if cpf.ConvertClassicHistogramsToNHCB == nil {
		return cfg
	}

	return cg.WithMinimumVersion("3.4.0").AppendMapItem(cfg, "convert_classic_histograms_to_nhcb", *cpf.ConvertClassicHistogramsToNHCB)
}

func (cg *ConfigGenerator) appendConvertScrapeClassicHistograms(cfg yaml.MapSlice) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()

	if cpf.ScrapeClassicHistograms == nil {
		return cfg
	}

	return cg.WithMinimumVersion("3.5.0").AppendMapItem(cfg, "always_scrape_classic_histograms", *cpf.ScrapeClassicHistograms)
}

func (cg *ConfigGenerator) appendScrapeNativeHistograms(cfg yaml.MapSlice) yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()

	if cpf.ScrapeNativeHistograms == nil {
		return cfg
	}

	return cg.WithMinimumVersion("3.8.0").AppendMapItem(cfg, "scrape_native_histograms", *cpf.ScrapeNativeHistograms)
}

func (cg *ConfigGenerator) getScrapeClassOrDefault(name *string) monitoringv1.ScrapeClass {
	if name != nil {
		if scrapeClass, found := cg.scrapeClasses[*name]; found {
			return scrapeClass
		}
	}

	if cg.defaultScrapeClassName != "" {
		if scrapeClass, found := cg.scrapeClasses[cg.defaultScrapeClassName]; found {
			return scrapeClass
		}
	}

	return monitoringv1.ScrapeClass{}
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

func (cg *ConfigGenerator) addFiltersToYaml(cfg yaml.MapSlice, filters []monitoringv1alpha1.Filter) yaml.MapSlice {
	if len(filters) == 0 {
		return cfg
	}

	// Sort the filters by name to generate deterministic config.
	slices.SortStableFunc(filters, func(a, b monitoringv1alpha1.Filter) int {
		return cmp.Compare(a.Name, b.Name)
	})

	filtersYamlMap := []yaml.MapSlice{}
	for _, filter := range filters {
		filtersYamlMap = append(filtersYamlMap, yaml.MapSlice{
			{
				Key:   "name",
				Value: filter.Name,
			},
			{
				Key:   "values",
				Value: filter.Values,
			}})
	}

	return cg.AppendMapItem(cfg, "filters", filtersYamlMap)
}

func (cg *ConfigGenerator) buildGlobalConfig() yaml.MapSlice {
	cpf := cg.prom.GetCommonPrometheusFields()
	cfg := yaml.MapSlice{}
	cfg = cg.appendScrapeIntervals(cfg)
	cfg = cg.addScrapeProtocols(cfg, cg.prom.GetCommonPrometheusFields().ScrapeProtocols)
	cfg = cg.appendExternalLabels(cfg)
	cfg = cg.appendScrapeLimits(cfg)
	cfg = cg.appendScrapeFailureLogFile(cfg, cg.prom.GetCommonPrometheusFields().ScrapeFailureLogFile)
	cfg = cg.appendNameValidationScheme(cfg, cpf.NameValidationScheme)
	cfg = cg.appendNameEscapingScheme(cfg, cpf.NameEscapingScheme)
	cfg = cg.appendConvertClassicHistogramsToNHCB(cfg)
	cfg = cg.appendConvertScrapeClassicHistograms(cfg)
	cfg = cg.appendScrapeNativeHistograms(cfg)

	return cfg
}
