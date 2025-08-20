// Copyright 2023 The prometheus-operator Authors
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
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/blang/semver/v4"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	// Generic reason for selected resources that are not valid.
	invalidConfiguration = "InvalidConfiguration"
)

// ConfigurationResource is a type constraint that permits only the specific pointer types for configuration resources
// selectable by Prometheus or PrometheusAgent.
type ConfigurationResource interface {
	*monitoringv1.ServiceMonitor | *monitoringv1.PodMonitor | *monitoringv1.Probe | *monitoringv1alpha1.ScrapeConfig
}

// ResourceSelector knows how to select and verify scrape configuration
// resources that are matched by a Prometheus or PrometheusAgent object.
type ResourceSelector struct {
	l                  *slog.Logger
	p                  monitoringv1.PrometheusInterface
	version            semver.Version
	store              *assets.StoreBuilder
	namespaceInformers cache.SharedIndexInformer
	metrics            *operator.Metrics
	accessor           *operator.Accessor

	eventRecorder record.EventRecorder
}

// TypedConfigurationResource is a generic type that holds a configuration resource with its validation status.
type TypedConfigurationResource[T ConfigurationResource] struct {
	resource T
	err      error  // error encountered during selection or validation (nil if valid).
	reason   string // Reason for rejection; empty if accepted.
}

func (r *TypedConfigurationResource[T]) conditions(observedGeneration int64) []monitoringv1.ConfigResourceCondition {
	condition := monitoringv1.ConfigResourceCondition{
		Type:               monitoringv1.Accepted,
		Status:             monitoringv1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             r.reason,
		ObservedGeneration: observedGeneration,
	}

	if r.err != nil {
		condition.Status = monitoringv1.ConditionFalse
		condition.Message = r.err.Error()
	}

	return []monitoringv1.ConfigResourceCondition{condition}
}

// TypedResourcesSelection represents a map of configuration resources selected by Prometheus or PrometheusAgent.
type TypedResourcesSelection[T ConfigurationResource] map[string]TypedConfigurationResource[T]

// ValidResources returns only the resources which the operator considers to be valid.
// The keys of the returned map identify the resources using the `<namespace>/<name>` format.
func (resources TypedResourcesSelection[T]) ValidResources() map[string]T {
	validRes := make(map[string]T)
	for k, res := range resources {
		if res.err == nil {
			validRes[k] = res.resource
		}
	}
	return validRes
}

type ListAllByNamespaceFn func(namespace string, selector labels.Selector, appendFn cache.AppendFunc) error

func NewResourceSelector(
	l *slog.Logger,
	p monitoringv1.PrometheusInterface,
	store *assets.StoreBuilder,
	namespaceInformers cache.SharedIndexInformer,
	metrics *operator.Metrics,
	eventRecorder record.EventRecorder,
) (*ResourceSelector, error) {
	promVersion := operator.StringValOrDefault(p.GetCommonPrometheusFields().Version, operator.DefaultPrometheusVersion)
	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus version: %w", err)
	}

	return &ResourceSelector{
		l:                  l,
		p:                  p,
		version:            version,
		store:              store,
		namespaceInformers: namespaceInformers,
		metrics:            metrics,
		eventRecorder:      eventRecorder,
		accessor:           operator.NewAccessor(l),
	}, nil
}

func selectObjects[T ConfigurationResource](
	ctx context.Context,
	logger *slog.Logger,
	rs *ResourceSelector,
	kind string,
	selector *metav1.LabelSelector,
	nsSelector *metav1.LabelSelector,
	listFn ListAllByNamespaceFn,
	checkFn func(context.Context, T) error,
) (TypedResourcesSelection[T], error) {
	// Selectors (<namespace>/<name>) might overlap. Deduplicate them along the keyFunc.
	objects := make(map[string]runtime.Object)
	namespaces := []string{}

	labelSelector, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, err
	}

	if nsSelector == nil {
		namespaces = append(namespaces, rs.p.GetObjectMeta().GetNamespace())
	} else {
		nsLabelSelector, err := metav1.LabelSelectorAsSelector(nsSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = operator.ListMatchingNamespaces(nsLabelSelector, rs.namespaceInformers)
		if err != nil {
			return nil, err
		}
	}

	logger.Debug(
		"selecting objects",
		"namespaces", strings.Join(namespaces, ","),
	)

	for _, ns := range namespaces {
		err := listFn(ns, labelSelector, func(o interface{}) {
			k, ok := rs.accessor.MetaNamespaceKey(o)
			if !ok {
				return
			}

			obj := o.(runtime.Object)
			obj = obj.DeepCopyObject()
			if err := k8sutil.AddTypeInformationToObject(obj); err != nil {
				logger.Error("failed to set type information", "namespace", ns, "err", err)
				return
			}
			objects[k] = obj
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in namespace %s: %w", ns, err)
		}
	}

	var (
		rejected int
		valid    []string
		res      = make(TypedResourcesSelection[T], len(objects))
	)

	for namespaceAndName, obj := range objects {
		var reason string
		o := obj.(T)
		err := checkFn(ctx, o)
		if err != nil {
			rejected++
			reason = invalidConfiguration
			logger.Warn("skipping object", "error", err.Error(), "object", namespaceAndName)
			rs.eventRecorder.Eventf(obj, v1.EventTypeWarning, operator.InvalidConfigurationEvent, "%q was rejected due to invalid configuration: %v", namespaceAndName, err)
		} else {
			valid = append(valid, namespaceAndName)
		}

		res[namespaceAndName] = TypedConfigurationResource[T]{
			resource: o,
			err:      err,
			reason:   reason,
		}
	}

	logger.Debug("valid objects selected", "objects", strings.Join(valid, ","))

	if pKey, ok := rs.accessor.MetaNamespaceKey(rs.p); ok {
		rs.metrics.SetSelectedResources(pKey, kind, len(res))
		rs.metrics.SetRejectedResources(pKey, kind, rejected)
	}

	return res, nil
}

// SelectServiceMonitors returns the ServiceMonitors that match the selectors in the Prometheus custom resource.
// This function also populates authentication stores and
// performs validations against scrape intervals and relabel configs.
func (rs *ResourceSelector) SelectServiceMonitors(ctx context.Context, listFn ListAllByNamespaceFn) (TypedResourcesSelection[*monitoringv1.ServiceMonitor], error) {
	cpf := rs.p.GetCommonPrometheusFields()

	return selectObjects(
		ctx,
		rs.l.With("kind", monitoringv1.ServiceMonitorsKind),
		rs,
		monitoringv1.ServiceMonitorsKind,
		cpf.ServiceMonitorSelector,
		cpf.ServiceMonitorNamespaceSelector,
		listFn,
		rs.checkServiceMonitor,
	)
}

// checkServiceMonitor verifies that the ServiceMonitor object is valid.
func (rs *ResourceSelector) checkServiceMonitor(ctx context.Context, sm *monitoringv1.ServiceMonitor) error {
	cpf := rs.p.GetCommonPrometheusFields()

	if _, err := metav1.LabelSelectorAsSelector(&sm.Spec.Selector); err != nil {
		return err
	}

	if err := rs.validateMonitorSelectorMechanism(sm.Spec.SelectorMechanism); err != nil {
		return err
	}

	for i, endpoint := range sm.Spec.Endpoints {
		epErr := fmt.Errorf("endpoints[%d]", i)

		// If denied by the Prometheus spec, filter out all service monitors
		// that access the file system.
		if cpf.ArbitraryFSAccessThroughSMs.Deny {
			if err := testForArbitraryFSAccess(endpoint); err != nil {
				return fmt.Errorf("%w: %w", epErr, err)
			}
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if endpoint.BearerTokenSecret != nil && endpoint.BearerTokenSecret.Name != "" {
			if _, err := rs.store.GetSecretKey(ctx, sm.GetNamespace(), *endpoint.BearerTokenSecret); err != nil {
				return fmt.Errorf("%w: bearerTokenSecret: %w", epErr, err)
			}
		}

		if err := rs.store.AddBasicAuth(ctx, sm.GetNamespace(), endpoint.BasicAuth); err != nil {
			return fmt.Errorf("%w: basicAuth: %w", epErr, err)
		}

		if err := rs.store.AddTLSConfig(ctx, sm.GetNamespace(), endpoint.TLSConfig); err != nil {
			return fmt.Errorf("%w: tlsConfig: %w", epErr, err)
		}

		if err := rs.store.AddOAuth2(ctx, sm.GetNamespace(), endpoint.OAuth2); err != nil {
			return fmt.Errorf("%w: oauth2: %w", epErr, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sm.GetNamespace(), endpoint.Authorization); err != nil {
			return fmt.Errorf("%w: authorization: %w", epErr, err)
		}

		if err := validateScrapeIntervalAndTimeout(rs.p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
			return fmt.Errorf("%w: %w", epErr, err)
		}

		if err := rs.ValidateRelabelConfigs(endpoint.RelabelConfigs); err != nil {
			return fmt.Errorf("%w: relabelConfigs: %w", epErr, err)
		}

		if err := rs.ValidateRelabelConfigs(endpoint.MetricRelabelConfigs); err != nil {
			return fmt.Errorf("%w: metricRelabelConfigs: %w", epErr, err)
		}

		if err := addProxyConfigToStore(ctx, endpoint.ProxyConfig, rs.store, sm.GetNamespace()); err != nil {
			return err
		}
	}

	if err := validateScrapeClass(rs.p, sm.Spec.ScrapeClassName); err != nil {
		return fmt.Errorf("scrapeClassName: %w", err)
	}

	return nil
}

func (rs *ResourceSelector) ValidateRelabelConfigs(rcs []monitoringv1.RelabelConfig) error {
	lcv := &LabelConfigValidator{v: rs.version}
	return lcv.Validate(rcs)
}

func testForArbitraryFSAccess(e monitoringv1.Endpoint) error {
	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if e.BearerTokenFile != "" {
		return errors.New("it accesses file system via bearer token file which Prometheus specification prohibits")
	}

	tlsConf := e.TLSConfig
	if tlsConf == nil {
		return nil
	}

	if tlsConf.CAFile != "" || tlsConf.CertFile != "" || tlsConf.KeyFile != "" {
		return errors.New("it accesses file system via tls config which Prometheus specification prohibits")
	}

	return nil
}

func validateScrapeIntervalAndTimeout(p monitoringv1.PrometheusInterface, scrapeInterval, scrapeTimeout monitoringv1.Duration) error {
	if scrapeTimeout == "" {
		return nil
	}
	if scrapeInterval == "" {
		scrapeInterval = p.GetCommonPrometheusFields().ScrapeInterval
	}
	return CompareScrapeTimeoutToScrapeInterval(scrapeTimeout, scrapeInterval)
}

type LabelConfigValidator struct {
	v semver.Version
}

func NewLabelConfigValidator(p monitoringv1.PrometheusInterface) (*LabelConfigValidator, error) {
	promVersion := operator.StringValOrDefault(p.GetCommonPrometheusFields().Version, operator.DefaultPrometheusVersion)
	v, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus version: %w", err)
	}

	return &LabelConfigValidator{
		v: v,
	}, nil
}

func (lcv *LabelConfigValidator) Validate(rcs []monitoringv1.RelabelConfig) error {
	for i, rc := range rcs {
		if err := lcv.validate(rc); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (lcv *LabelConfigValidator) validate(rc monitoringv1.RelabelConfig) error {
	relabelTarget := regexp.MustCompile(`^(?:(?:[a-zA-Z_]|\$(?:\{\w+\}|\w+))+\w*)+$`)

	minimumVersionCaseActions := lcv.v.GTE(semver.MustParse("2.36.0"))
	minimumVersionEqualActions := lcv.v.GTE(semver.MustParse("2.41.0"))
	if rc.Action == "" {
		rc.Action = string(relabel.Replace)
	}
	action := strings.ToLower(rc.Action)

	if (action == string(relabel.Lowercase) || action == string(relabel.Uppercase)) && !minimumVersionCaseActions {
		return fmt.Errorf("%s relabel action is only supported from Prometheus version 2.36.0", rc.Action)
	}

	if (action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && !minimumVersionEqualActions {
		return fmt.Errorf("%s relabel action is only supported from Prometheus version 2.41.0", rc.Action)
	}

	if _, err := relabel.NewRegexp(rc.Regex); err != nil {
		return fmt.Errorf("invalid regex %s for relabel configuration: %w", rc.Regex, err)
	}

	if rc.Modulus == 0 && action == string(relabel.HashMod) {
		return fmt.Errorf("relabel configuration for hashmod requires non-zero modulus")
	}

	if (action == string(relabel.Replace) || action == string(relabel.HashMod) || action == string(relabel.Lowercase) || action == string(relabel.Uppercase) || action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && rc.TargetLabel == "" {
		return fmt.Errorf("relabel configuration for %s action needs targetLabel value", rc.Action)
	}

	if (action == string(relabel.Replace) || action == string(relabel.Lowercase) || action == string(relabel.Uppercase) || action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && !relabelTarget.MatchString(rc.TargetLabel) {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if (action == string(relabel.Lowercase) || action == string(relabel.Uppercase) || action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && (rc.Replacement != nil && *rc.Replacement != relabel.DefaultRelabelConfig.Replacement) {
		return fmt.Errorf("'replacement' can not be set for %s action", rc.Action)
	}

	if action == string(relabel.LabelMap) {
		if rc.Replacement != nil && !relabelTarget.MatchString(*rc.Replacement) {
			return fmt.Errorf("%q is invalid 'replacement' for %s action", *rc.Replacement, rc.Action)
		}
	}

	if action == string(relabel.HashMod) && !model.LabelName(rc.TargetLabel).IsValid() {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if action == string(relabel.KeepEqual) || action == string(relabel.DropEqual) {
		if (rc.Regex != "" && rc.Regex != relabel.DefaultRelabelConfig.Regex.String()) ||
			(rc.Modulus != uint64(0) &&
				rc.Modulus != relabel.DefaultRelabelConfig.Modulus) ||
			(rc.Separator != nil &&
				*rc.Separator != relabel.DefaultRelabelConfig.Separator) ||
			(rc.Replacement != nil && *rc.Replacement != relabel.DefaultRelabelConfig.Replacement) {
			return fmt.Errorf("%s action requires only 'source_labels' and `target_label`, and no other fields", rc.Action)
		}
	}

	if action == string(relabel.LabelDrop) || action == string(relabel.LabelKeep) {
		if len(rc.SourceLabels) != 0 ||
			(rc.TargetLabel != "" &&
				rc.TargetLabel != relabel.DefaultRelabelConfig.TargetLabel) ||
			(rc.Modulus != uint64(0) &&
				rc.Modulus != relabel.DefaultRelabelConfig.Modulus) ||
			(rc.Separator != nil &&
				*rc.Separator != relabel.DefaultRelabelConfig.Separator) ||
			(rc.Replacement != nil &&
				*rc.Replacement != relabel.DefaultRelabelConfig.Replacement) {
			return fmt.Errorf("%s action requires only 'regex', and no other fields", rc.Action)
		}
	}
	return nil
}

func validateScrapeClass(p monitoringv1.PrometheusInterface, sc *string) error {
	if ptr.Deref(sc, "") == "" {
		return nil
	}

	for _, c := range p.GetCommonPrometheusFields().ScrapeClasses {
		if c.Name == *sc {
			return nil
		}
	}

	return fmt.Errorf("scrapeClass %q not found in Prometheus scrapeClasses", *sc)
}

func (rs *ResourceSelector) validateMonitorSelectorMechanism(selectorMechanism *monitoringv1.SelectorMechanism) error {
	if ptr.Deref(selectorMechanism, monitoringv1.SelectorMechanismRelabel) == monitoringv1.SelectorMechanismRole && !rs.version.GTE(semver.MustParse("2.17.0")) {
		return fmt.Errorf("RoleSelector selectorMechanism is only supported in Prometheus 2.17.0 and newer")
	}

	return nil
}

// SelectPodMonitors returns the PodMonitors that match the selectors in the Prometheus custom resource.
// This function also populates authentication stores and
// performs validations against scrape intervals and relabel configs.
func (rs *ResourceSelector) SelectPodMonitors(ctx context.Context, listFn ListAllByNamespaceFn) (TypedResourcesSelection[*monitoringv1.PodMonitor], error) {
	cpf := rs.p.GetCommonPrometheusFields()

	return selectObjects(
		ctx,
		rs.l.With("kind", monitoringv1.PodMonitorsKind),
		rs,
		monitoringv1.PodMonitorsKind,
		cpf.PodMonitorSelector,
		cpf.PodMonitorNamespaceSelector,
		listFn,
		rs.checkPodMonitor,
	)
}

// checkPodMonitor verifies that the PodMonitor object is valid.
func (rs *ResourceSelector) checkPodMonitor(ctx context.Context, pm *monitoringv1.PodMonitor) error {
	if _, err := metav1.LabelSelectorAsSelector(&pm.Spec.Selector); err != nil {
		return fmt.Errorf("failed to parse label selector: %w", err)
	}

	if err := rs.validateMonitorSelectorMechanism(pm.Spec.SelectorMechanism); err != nil {
		return err
	}

	for i, endpoint := range pm.Spec.PodMetricsEndpoints {
		epErr := fmt.Errorf("endpoint[%d]", i)
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if endpoint.BearerTokenSecret.Name != "" && endpoint.BearerTokenSecret.Key != "" {
			if _, err := rs.store.GetSecretKey(ctx, pm.GetNamespace(), endpoint.BearerTokenSecret); err != nil {
				return fmt.Errorf("%w: bearerTokenSecret: %w", epErr, err)
			}
		}

		if err := rs.store.AddBasicAuth(ctx, pm.GetNamespace(), endpoint.BasicAuth); err != nil {
			return fmt.Errorf("%w: basicAuth: %w", epErr, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, pm.GetNamespace(), endpoint.TLSConfig); err != nil {
			return fmt.Errorf("%w: tlsConfig: %w", epErr, err)
		}

		if err := rs.store.AddOAuth2(ctx, pm.GetNamespace(), endpoint.OAuth2); err != nil {
			return fmt.Errorf("%w: oauth2: %w", epErr, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, pm.GetNamespace(), endpoint.Authorization); err != nil {
			return fmt.Errorf("%w: authorization: %w", epErr, err)
		}

		if err := validateScrapeIntervalAndTimeout(rs.p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
			return fmt.Errorf("%w: %w", epErr, err)
		}

		if err := rs.ValidateRelabelConfigs(endpoint.RelabelConfigs); err != nil {
			return fmt.Errorf("%w: relabelConfigs: %w", epErr, err)
		}

		if err := rs.ValidateRelabelConfigs(endpoint.MetricRelabelConfigs); err != nil {
			return fmt.Errorf("%w: metricRelabelConfigs: %w", epErr, err)
		}

		if err := addProxyConfigToStore(ctx, endpoint.ProxyConfig, rs.store, pm.GetNamespace()); err != nil {
			return fmt.Errorf("%w: proxyConfig: %w", epErr, err)
		}
	}

	if err := validateScrapeClass(rs.p, pm.Spec.ScrapeClassName); err != nil {
		return fmt.Errorf("scrapeClassName: %w", err)
	}

	return nil
}

// SelectProbes returns the probes matching the selectors specified in the Prometheus CR.
// This function also populates authentication stores and performs
// validations against scrape intervals, relabel configs and Probe URLs.
func (rs *ResourceSelector) SelectProbes(ctx context.Context, listFn ListAllByNamespaceFn) (TypedResourcesSelection[*monitoringv1.Probe], error) {
	cpf := rs.p.GetCommonPrometheusFields()

	return selectObjects(
		ctx,
		rs.l.With("kind", monitoringv1.ProbesKind),
		rs,
		monitoringv1.ProbesKind,
		cpf.ProbeSelector,
		cpf.ProbeNamespaceSelector,
		listFn,
		rs.checkProbe,
	)
}

// checkProbe verifies that the Probe object is valid.
func (rs *ResourceSelector) checkProbe(ctx context.Context, probe *monitoringv1.Probe) error {
	if err := validateScrapeClass(rs.p, probe.Spec.ScrapeClassName); err != nil {
		return fmt.Errorf("scrapeClassName: %w", err)
	}

	if err := probe.Spec.Targets.Validate(); err != nil {
		return err
	}

	if probe.Spec.BearerTokenSecret.Name != "" && probe.Spec.BearerTokenSecret.Key != "" {
		if _, err := rs.store.GetSecretKey(ctx, probe.GetNamespace(), probe.Spec.BearerTokenSecret); err != nil {
			return fmt.Errorf("bearerTokenSecret: %w", err)
		}
	}

	if err := rs.store.AddBasicAuth(ctx, probe.GetNamespace(), probe.Spec.BasicAuth); err != nil {
		return fmt.Errorf("basicAuth: %w", err)
	}

	if err := rs.store.AddSafeTLSConfig(ctx, probe.GetNamespace(), probe.Spec.TLSConfig); err != nil {
		return fmt.Errorf("tlsConfig: %w", err)
	}

	if err := rs.store.AddSafeAuthorizationCredentials(ctx, probe.GetNamespace(), probe.Spec.Authorization); err != nil {
		return fmt.Errorf("authorization: %w", err)
	}

	if err := rs.store.AddOAuth2(ctx, probe.GetNamespace(), probe.Spec.OAuth2); err != nil {
		return fmt.Errorf("oauth2: %w", err)
	}

	if err := validateScrapeIntervalAndTimeout(rs.p, probe.Spec.Interval, probe.Spec.ScrapeTimeout); err != nil {
		return err
	}

	if err := rs.ValidateRelabelConfigs(probe.Spec.MetricRelabelConfigs); err != nil {
		return fmt.Errorf("metricRelabelConfigs: %w", err)
	}

	if probe.Spec.Targets.StaticConfig != nil {
		if err := rs.ValidateRelabelConfigs(probe.Spec.Targets.StaticConfig.RelabelConfigs); err != nil {
			return fmt.Errorf("targets.staticConfig.relabelConfigs: %w", err)
		}
	}

	if probe.Spec.Targets.Ingress != nil {
		if err := rs.ValidateRelabelConfigs(probe.Spec.Targets.Ingress.RelabelConfigs); err != nil {
			return fmt.Errorf("targets.ingress.relabelConfigs: %w", err)
		}
	}

	if err := addProxyConfigToStore(ctx, probe.Spec.ProberSpec.ProxyConfig, rs.store, probe.GetNamespace()); err != nil {
		return fmt.Errorf("proxy configuration: %w", err)
	}

	if err := validateProberURL(probe.Spec.ProberSpec.URL); err != nil {
		return fmt.Errorf("%q url specified in proberSpec is invalid, it should be of the format `hostname` or `hostname:port`: %w", probe.Spec.ProberSpec.URL, err)
	}

	return nil
}

func validateProberURL(url string) error {
	hostPort := strings.Split(url, ":")

	if !govalidator.IsHost(hostPort[0]) {
		return fmt.Errorf("invalid host: %q", hostPort[0])
	}

	// handling cases with url specified as host:port
	if len(hostPort) > 1 {
		if !govalidator.IsPort(hostPort[1]) {
			return fmt.Errorf("invalid port: %q", hostPort[1])
		}
	}

	return nil
}

func validateServer(server string) error {
	parsedURL, err := url.Parse(server)
	if err != nil {
		return fmt.Errorf("cannot parse server: %s", err.Error())
	}

	if len(parsedURL.Scheme) == 0 || len(parsedURL.Host) == 0 {
		return fmt.Errorf("must not be empty and have a scheme: %s", server)
	}

	return nil
}

// SelectScrapeConfigs returns the ScrapeConfigs which match the selectors in the
// Prometheus CR and filters them returning all the configuration.
func (rs *ResourceSelector) SelectScrapeConfigs(ctx context.Context, listFn ListAllByNamespaceFn) (TypedResourcesSelection[*monitoringv1alpha1.ScrapeConfig], error) {
	cpf := rs.p.GetCommonPrometheusFields()

	return selectObjects(
		ctx,
		rs.l.With("kind", monitoringv1alpha1.ScrapeConfigsKind),
		rs,
		monitoringv1alpha1.ScrapeConfigsKind,
		cpf.ScrapeConfigSelector,
		cpf.ScrapeConfigNamespaceSelector,
		listFn,
		rs.checkScrapeConfig,
	)
}

// checkScrapeConfig verifies that the ScrapeConfig object is valid.
func (rs *ResourceSelector) checkScrapeConfig(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if err := validateScrapeClass(rs.p, sc.Spec.ScrapeClassName); err != nil {
		return err
	}

	if err := rs.ValidateRelabelConfigs(sc.Spec.RelabelConfigs); err != nil {
		return fmt.Errorf("relabelConfigs: %w", err)
	}

	if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), sc.Spec.BasicAuth); err != nil {
		return fmt.Errorf("basicAuth: %w", err)
	}

	if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), sc.Spec.Authorization); err != nil {
		return fmt.Errorf("authorization: %w", err)
	}

	if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), sc.Spec.OAuth2); err != nil {
		return fmt.Errorf("oauth2: %w", err)
	}

	if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), sc.Spec.TLSConfig); err != nil {
		return fmt.Errorf("tlsConfig: %w", err)
	}

	var scrapeInterval, scrapeTimeout monitoringv1.Duration = "", ""
	if sc.Spec.ScrapeInterval != nil {
		scrapeInterval = *sc.Spec.ScrapeInterval
	}

	if sc.Spec.ScrapeTimeout != nil {
		scrapeTimeout = *sc.Spec.ScrapeTimeout
	}

	if err := validateScrapeIntervalAndTimeout(rs.p, scrapeInterval, scrapeTimeout); err != nil {
		return err
	}

	if err := addProxyConfigToStore(ctx, sc.Spec.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
		return err
	}

	if err := rs.ValidateRelabelConfigs(sc.Spec.MetricRelabelConfigs); err != nil {
		return fmt.Errorf("metricRelabelConfigs: %w", err)
	}

	// The Kubernetes API can't do the validation (for now) because kubebuilder validation markers don't work on map keys with custom type.
	// https://github.com/prometheus-operator/prometheus-operator/issues/6889
	if err := rs.validateStaticConfig(sc); err != nil {
		return fmt.Errorf("staticConfigs: %w", err)
	}

	if err := rs.validateHTTPSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("httpSDConfigs: %w", err)
	}

	if err := rs.validateKubernetesSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("kubernetesSDConfigs: %w", err)
	}

	if err := rs.validateConsulSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("consulSDConfigs: %w", err)
	}

	if err := rs.validateDNSSDConfigs(sc); err != nil {
		return fmt.Errorf("dnsSDConfigs: %w", err)
	}

	if err := rs.validateEC2SDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("ec2SDConfigs: %w", err)
	}

	if err := rs.validateAzureSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("azureSDConfigs: %w", err)
	}

	if err := rs.validateOpenStackSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("openstackSDConfigs: %w", err)
	}

	if err := rs.validateDigitalOceanSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("digitalOceanSDConfigs: %w", err)
	}

	if err := rs.validateKumaSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("kumaSDConfigs: %w", err)
	}

	if err := rs.validateEurekaSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("eurekaSDConfigs: %w", err)
	}

	if err := rs.validateDockerSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("dockerSDConfigs: %w", err)
	}

	if err := rs.validateLinodeSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("linodeSDConfigs: %w", err)
	}

	if err := rs.validateHetznerSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("hetznerSDConfigs: %w", err)
	}

	if err := rs.validateNomadSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("nomadSDConfigs: %w", err)
	}

	if err := rs.validateDockerSwarmSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("dockerswarmSDConfigs: %w", err)
	}

	if err := rs.validatePuppetDBSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("puppetDBSDConfigs: %w", err)
	}

	if err := rs.validateLightSailSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("lightSailSDConfigs: %w", err)
	}

	if err := rs.validateOVHCloudSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("OVHCloudSDConfigs: %w", err)
	}

	if err := rs.validateScalewaySDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("ScalewaySDConfigs: %w", err)
	}

	if err := rs.validateIonosSDConfigs(ctx, sc); err != nil {
		return fmt.Errorf("IonosSDConfigs: %w", err)
	}

	return nil
}

func (rs *ResourceSelector) validateKubernetesSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.KubernetesSDConfigs {
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if config.APIServer != nil && config.Namespaces != nil {
			if ptr.Deref(config.Namespaces.IncludeOwnNamespace, false) {
				return fmt.Errorf("[%d]: %w", i, errors.New("cannot use 'apiServer' and 'namespaces.ownNamespace' simultaneously"))
			}
		}

		allowedSelectors := map[string][]string{
			monitoringv1.RolePod:           {string(monitoringv1.RolePod)},
			monitoringv1.RoleService:       {string(monitoringv1.RoleService)},
			monitoringv1.RoleEndpointSlice: {string(monitoringv1.RolePod), string(monitoringv1.RoleService), string(monitoringv1.RoleEndpointSlice)},
			monitoringv1.RoleEndpoint:      {string(monitoringv1.RolePod), string(monitoringv1.RoleService), string(monitoringv1.RoleEndpoint)},
			monitoringv1.RoleNode:          {string(monitoringv1.RoleNode)},
			monitoringv1.RoleIngress:       {string(monitoringv1.RoleIngress)},
		}

		for _, s := range config.Selectors {
			configRole := strings.ToLower(string(config.Role))
			if _, ok := allowedSelectors[configRole]; !ok {
				return fmt.Errorf("[%d]: invalid role: %q, expecting one of: pod, service, endpoints, endpointslice, node or ingress", i, s.Role)
			}

			var allowed bool

			for _, role := range allowedSelectors[configRole] {
				if role == strings.ToLower(string(s.Role)) {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("[%d] : %s role supports only %s selectors", i, config.Role, strings.Join(allowedSelectors[configRole], ", "))
			}
		}

		for _, s := range config.Selectors {
			if s.Field != nil {
				if _, err := fields.ParseSelector(*s.Field); err != nil {
					return fmt.Errorf("[%d]: %w", i, err)
				}
			}

			if s.Label != nil {
				if _, err := labels.Parse(*s.Label); err != nil {
					return fmt.Errorf("[%d]: %w", i, err)
				}
			}
		}
	}

	return nil
}

func (rs *ResourceSelector) validateConsulSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.ConsulSDConfigs {
		if config.PathPrefix != nil && rs.version.LT(semver.MustParse("2.45.0")) {
			return fmt.Errorf("field `config.PathPrefix` is only supported for Prometheus version >= 2.45.0")
		}

		if config.Namespace != nil && rs.version.LT(semver.MustParse("2.28.0")) {
			return fmt.Errorf("field `config.Namespace` is only supported for Prometheus version >= 2.28.0")
		}

		if config.Filter != nil && rs.version.Major < 3 {
			return fmt.Errorf("field `config.Filter` is only supported for Prometheus version >= 3.0.0")
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if config.TokenRef != nil {
			if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.TokenRef); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateHTTPSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.28.0")) {
		return fmt.Errorf("HTTP SD configuration is only supported for Prometheus version >= 2.28.0")
	}

	for i, config := range sc.Spec.HTTPSDConfigs {
		if _, err := url.Parse(config.URL); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateDNSSDConfigs(sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.DNSSDConfigs {
		if config.Type != nil {
			if *config.Type != "SRV" && config.Port == nil {
				return fmt.Errorf("[%d]: %s %q", i, "port required for record type", *config.Type)
			}
		}
	}

	return nil
}

func (rs *ResourceSelector) validateEC2SDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.EC2SDConfigs {
		if config.AccessKey != nil {
			if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.AccessKey); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}

		if config.SecretKey != nil {
			if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.SecretKey); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateAzureSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.AzureSDConfigs {
		authMethod := ptr.Deref(config.AuthenticationMethod, "")
		if authMethod == "SDK" && rs.version.LT(semver.MustParse("2.52.0")) {
			return fmt.Errorf("[%d]: SDK authentication is only supported from Prometheus version 2.52.0", i)
		}

		if config.ResourceGroup != nil && rs.version.LT(semver.MustParse("2.35.0")) {
			return fmt.Errorf("[%d]: ResourceGroup is only supported from Prometheus version >= 2.35.0", i)
		}

		// Since Prometheus uses default authentication method as "OAuth"
		if authMethod == "ManagedIdentity" || authMethod == "SDK" {
			continue
		}

		if len(ptr.Deref(config.TenantID, "")) == 0 {
			return fmt.Errorf("[%d]: configuration requires a tenantID", i)
		}

		if len(ptr.Deref(config.ClientID, "")) == 0 {
			return fmt.Errorf("[%d]: configuration requires a clientID", i)
		}

		if config.ClientSecret == nil {
			return fmt.Errorf("[%d]: configuration requires a clientSecret", i)
		}

		if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.ClientSecret); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

	}

	return nil
}

func (rs *ResourceSelector) validateOpenStackSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.OpenStackSDConfigs {
		if config.Role == monitoringv1alpha1.OpenStackRoleLoadBalancer && rs.version.LT(semver.MustParse("3.2.0")) {
			return fmt.Errorf("[%d]: The %s role is only supported from Prometheus version 3.2.0", i, string(config.Role))
		}
		if config.Password != nil {
			if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.Password); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}

		if config.ApplicationCredentialSecret != nil {
			if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.ApplicationCredentialSecret); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
	}
	return nil
}

func (rs *ResourceSelector) validateDigitalOceanSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.20.0")) {
		return fmt.Errorf("service discovery for Digital Ocean is only supported for Prometheus version >= 2.20.0")
	}

	for i, config := range sc.Spec.DigitalOceanSDConfigs {
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateDockerSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.DockerSDConfigs {
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		// Validate the host daemon address url
		if _, err := url.Parse(config.Host); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}
func (rs *ResourceSelector) validateLinodeSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if !rs.version.GTE(semver.MustParse("2.28.0")) {
		return fmt.Errorf("linode SD configuration is only supported for Prometheus version >= 2.28.0")
	}

	for i, config := range sc.Spec.LinodeSDConfigs {
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}
func (rs *ResourceSelector) validateKumaSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.KumaSDConfigs {
		if err := validateServer(config.Server); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateEurekaSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.EurekaSDConfigs {
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateHetznerSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.HetznerSDConfigs {
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateNomadSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.NomadSDConfigs {
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateDockerSwarmSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.20.0")) {
		return fmt.Errorf("dockerswarm SD configuration is only supported for Prometheus version >= 2.20.0")
	}

	for i, config := range sc.Spec.DockerSwarmSDConfigs {
		if _, err := url.Parse(config.Host); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validatePuppetDBSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.31.0")) {
		return fmt.Errorf("puppetDB SD configuration is only supported for Prometheus version >= 2.31.0")
	}

	for i, config := range sc.Spec.PuppetDBSDConfigs {
		parsedURL, err := url.Parse(config.URL)
		if err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("[%d]: URL scheme must be 'http' or 'https'", i)
		}
		if parsedURL.Host == "" {
			return fmt.Errorf("[%d]: host is missing in URL", i)
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateLightSailSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.27.0")) {
		return fmt.Errorf("lightSail SD configuration is only supported for Prometheus version >= 2.27.0")
	}

	for i, config := range sc.Spec.LightSailSDConfigs {
		if config.AccessKey != nil {
			if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.AccessKey); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
		if config.SecretKey != nil {
			if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), *config.SecretKey); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}

		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateOVHCloudSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.40.0")) {
		return fmt.Errorf("OVHCloud SD configuration is only supported for Prometheus version >= 2.40.0")
	}

	for i, config := range sc.Spec.OVHCloudSDConfigs {
		if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), config.ApplicationSecret); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), config.ConsumerKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateScalewaySDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.26.0")) {
		return fmt.Errorf("ScaleWay SD configuration is only supported for Prometheus version >= 2.26.0")
	}

	for i, config := range sc.Spec.ScalewaySDConfigs {
		if _, err := rs.store.GetSecretKey(ctx, sc.GetNamespace(), config.SecretKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateStaticConfig(sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.StaticConfigs {
		for labelName := range config.Labels {
			if !model.LabelName(labelName).IsValid() {
				return fmt.Errorf("[%d]: invalid label in map %s", i, labelName)
			}
		}
	}

	return nil
}

func (rs *ResourceSelector) validateIonosSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	if rs.version.LT(semver.MustParse("2.36.0")) {
		return fmt.Errorf("IONOS SD configuration is only supported for Prometheus version >= 2.36.0")
	}

	for i, config := range sc.Spec.IonosSDConfigs {
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), &config.Authorization); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := addProxyConfigToStore(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

	}
	return nil
}
