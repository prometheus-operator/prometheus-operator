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
	"net/url"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

type ResourceSelector struct {
	l                  log.Logger
	p                  monitoringv1.PrometheusInterface
	store              *assets.Store
	namespaceInformers cache.SharedIndexInformer
	metrics            *operator.Metrics
	accessor           *operator.Accessor

	eventRecorder record.EventRecorder
}

type ListAllByNamespaceFn func(namespace string, selector labels.Selector, appendFn cache.AppendFunc) error

func NewResourceSelector(l log.Logger, p monitoringv1.PrometheusInterface, store *assets.Store, namespaceInformers cache.SharedIndexInformer, metrics *operator.Metrics, eventRecorder record.EventRecorder) *ResourceSelector {
	return &ResourceSelector{
		l:                  l,
		p:                  p,
		store:              store,
		namespaceInformers: namespaceInformers,
		metrics:            metrics,
		eventRecorder:      eventRecorder,
		accessor:           operator.NewAccessor(l),
	}
}

// SelectServiceMonitors selects ServiceMonitors based on the selectors in the Prometheus CR and filters them
// returning only those with a valid configuration. This function also populates authentication stores and performs validations against
// scrape intervals and relabel configs.
func (rs *ResourceSelector) SelectServiceMonitors(ctx context.Context, listFn ListAllByNamespaceFn) (map[string]*monitoringv1.ServiceMonitor, error) {
	cpf := rs.p.GetCommonPrometheusFields()
	objMeta := rs.p.GetObjectMeta()
	namespaces := []string{}
	// Selectors (<namespace>/<name>) might overlap. Deduplicate them along the keyFunc.
	serviceMonitors := make(map[string]*monitoringv1.ServiceMonitor)

	servMonSelector, err := metav1.LabelSelectorAsSelector(cpf.ServiceMonitorSelector)
	if err != nil {
		return nil, err
	}

	// If 'ServiceMonitorNamespaceSelector' is nil only check own namespace.
	if cpf.ServiceMonitorNamespaceSelector == nil {
		namespaces = append(namespaces, objMeta.GetNamespace())
	} else {
		servMonNSSelector, err := metav1.LabelSelectorAsSelector(cpf.ServiceMonitorNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = operator.ListMatchingNamespaces(servMonNSSelector, rs.namespaceInformers)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(rs.l).Log("msg", "filtering namespaces to select ServiceMonitors from", "namespaces", strings.Join(namespaces, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	for _, ns := range namespaces {
		err := listFn(ns, servMonSelector, func(obj interface{}) {
			k, ok := rs.accessor.MetaNamespaceKey(obj)
			if ok {
				svcMon := obj.(*monitoringv1.ServiceMonitor).DeepCopy()
				if err := k8sutil.AddTypeInformationToObject(svcMon); err != nil {
					level.Error(rs.l).Log("msg", "failed to set ServiceMonitor type information", "namespace", ns, "err", err)
					return
				}
				serviceMonitors[k] = svcMon
			}
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list service monitors in namespace %s: %w", ns, err)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1.ServiceMonitor, len(serviceMonitors))
	for namespaceAndName, sm := range serviceMonitors {
		var err error
		rejectFn := func(sm *monitoringv1.ServiceMonitor, err error) {
			rejected++
			level.Warn(rs.l).Log(
				"msg", "skipping servicemonitor",
				"error", err.Error(),
				"servicemonitor", namespaceAndName,
				"namespace", objMeta.GetNamespace(),
				"prometheus", objMeta.GetName(),
			)
			rs.eventRecorder.Eventf(sm, v1.EventTypeWarning, operator.InvalidConfigurationEvent, "ServiceMonitor %s was rejected due to invalid configuration: %v", sm.GetName(), err)
		}

		for i, endpoint := range sm.Spec.Endpoints {
			// If denied by Prometheus spec, filter out all service monitors that access
			// the file system.
			if cpf.ArbitraryFSAccessThroughSMs.Deny {
				if err = testForArbitraryFSAccess(endpoint); err != nil {
					rejectFn(sm, err)
					break
				}
			}

			smKey := fmt.Sprintf("serviceMonitor/%s/%s/%d", sm.GetNamespace(), sm.GetName(), i)

			//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
			if err = rs.store.AddBearerToken(ctx, sm.GetNamespace(), endpoint.BearerTokenSecret, smKey); err != nil {
				rejectFn(sm, err)
				break
			}

			if err = rs.store.AddBasicAuth(ctx, sm.GetNamespace(), endpoint.BasicAuth, smKey); err != nil {
				rejectFn(sm, err)
				break
			}

			if err = rs.store.AddTLSConfig(ctx, sm.GetNamespace(), endpoint.TLSConfig); err != nil {
				rejectFn(sm, err)
				break
			}

			if err = rs.store.AddOAuth2(ctx, sm.GetNamespace(), endpoint.OAuth2, smKey); err != nil {
				rejectFn(sm, err)
				break
			}

			smAuthKey := fmt.Sprintf("serviceMonitor/auth/%s/%s/%d", sm.GetNamespace(), sm.GetName(), i)
			if err = rs.store.AddSafeAuthorizationCredentials(ctx, sm.GetNamespace(), endpoint.Authorization, smAuthKey); err != nil {
				rejectFn(sm, err)
				break
			}

			if err = validateScrapeIntervalAndTimeout(rs.p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
				rejectFn(sm, err)
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.RelabelConfigs); err != nil {
				rejectFn(sm, fmt.Errorf("relabelConfigs: %w", err))
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.MetricRelabelConfigs); err != nil {
				rejectFn(sm, fmt.Errorf("metricRelabelConfigs: %w", err))
				break
			}

			if err = validateProxyURL(endpoint.ProxyURL); err != nil {
				rejectFn(sm, fmt.Errorf("proxyURL: %w", err))
				break
			}
		}

		if err != nil {
			continue
		}

		if err = validateScrapeClass(rs.p, sm.Spec.ScrapeClassName); err != nil {
			rejectFn(sm, err)
			continue
		}

		res[namespaceAndName] = sm
	}

	smKeys := []string{}
	for k := range res {
		smKeys = append(smKeys, k)
	}
	level.Debug(rs.l).Log("msg", "selected ServiceMonitors", "servicemonitors", strings.Join(smKeys, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	if pKey, ok := rs.accessor.MetaNamespaceKey(rs.p); ok {
		rs.metrics.SetSelectedResources(pKey, monitoringv1.ServiceMonitorsKind, len(res))
		rs.metrics.SetRejectedResources(pKey, monitoringv1.ServiceMonitorsKind, rejected)
	}

	return res, nil
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

func validateRelabelConfigs(p monitoringv1.PrometheusInterface, rcs []*monitoringv1.RelabelConfig) error {
	for i, rc := range rcs {
		if rc == nil {
			return fmt.Errorf("null relabel config")
		}

		if err := validateRelabelConfig(p, *rc); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateRelabelConfig(p monitoringv1.PrometheusInterface, rc monitoringv1.RelabelConfig) error {
	relabelTarget := regexp.MustCompile(`^(?:(?:[a-zA-Z_]|\$(?:\{\w+\}|\w+))+\w*)+$`)
	promVersion := operator.StringValOrDefault(p.GetCommonPrometheusFields().Version, operator.DefaultPrometheusVersion)

	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return fmt.Errorf("failed to parse Prometheus version: %w", err)
	}

	minimumVersionCaseActions := version.GTE(semver.MustParse("2.36.0"))
	minimumVersionEqualActions := version.GTE(semver.MustParse("2.41.0"))
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

	if (action == string(relabel.Lowercase) || action == string(relabel.Uppercase) || action == string(relabel.KeepEqual) || action == string(relabel.DropEqual)) && !(rc.Replacement == relabel.DefaultRelabelConfig.Replacement || rc.Replacement == "") {
		return fmt.Errorf("'replacement' can not be set for %s action", rc.Action)
	}

	if action == string(relabel.LabelMap) {
		if rc.Replacement != "" && !relabelTarget.MatchString(rc.Replacement) {
			return fmt.Errorf("%q is invalid 'replacement' for %s action", rc.Replacement, rc.Action)
		}
	}

	if action == string(relabel.HashMod) && !model.LabelName(rc.TargetLabel).IsValid() {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if action == string(relabel.KeepEqual) || action == string(relabel.DropEqual) {
		if !(rc.Regex == "" || rc.Regex == relabel.DefaultRelabelConfig.Regex.String()) ||
			!(rc.Modulus == uint64(0) ||
				rc.Modulus == relabel.DefaultRelabelConfig.Modulus) ||
			!(rc.Separator == nil ||
				*rc.Separator == relabel.DefaultRelabelConfig.Separator) ||
			!(rc.Replacement == relabel.DefaultRelabelConfig.Replacement ||
				rc.Replacement == "") {
			return fmt.Errorf("%s action requires only 'source_labels' and `target_label`, and no other fields", rc.Action)
		}
	}

	if action == string(relabel.LabelDrop) || action == string(relabel.LabelKeep) {
		if len(rc.SourceLabels) != 0 ||
			!(rc.TargetLabel == "" ||
				rc.TargetLabel == relabel.DefaultRelabelConfig.TargetLabel) ||
			!(rc.Modulus == uint64(0) ||
				rc.Modulus == relabel.DefaultRelabelConfig.Modulus) ||
			!(rc.Separator == nil ||
				*rc.Separator == relabel.DefaultRelabelConfig.Separator) ||
			!(rc.Replacement == relabel.DefaultRelabelConfig.Replacement ||
				rc.Replacement == "") {
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

// SelectPodMonitors selects PodMonitors based on the selectors in the Prometheus CR and filters them
// returning only those with a valid configuration. This function also populates authentication stores and performs validations against
// scrape intervals and relabel configs.
func (rs *ResourceSelector) SelectPodMonitors(ctx context.Context, listFn ListAllByNamespaceFn) (map[string]*monitoringv1.PodMonitor, error) {
	cpf := rs.p.GetCommonPrometheusFields()
	objMeta := rs.p.GetObjectMeta()
	namespaces := []string{}
	// Selectors (<namespace>/<name>) might overlap. Deduplicate them along the keyFunc.
	podMonitors := make(map[string]*monitoringv1.PodMonitor)

	podMonSelector, err := metav1.LabelSelectorAsSelector(cpf.PodMonitorSelector)
	if err != nil {
		return nil, err
	}

	// If 'PodMonitorNamespaceSelector' is nil only check own namespace.
	if cpf.PodMonitorNamespaceSelector == nil {
		namespaces = append(namespaces, objMeta.GetNamespace())
	} else {
		podMonNSSelector, err := metav1.LabelSelectorAsSelector(cpf.PodMonitorNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = operator.ListMatchingNamespaces(podMonNSSelector, rs.namespaceInformers)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(rs.l).Log("msg", "filtering namespaces to select PodMonitors from", "namespaces", strings.Join(namespaces, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	for _, ns := range namespaces {
		err := listFn(ns, podMonSelector, func(obj interface{}) {
			k, ok := rs.accessor.MetaNamespaceKey(obj)
			if ok {
				podMon := obj.(*monitoringv1.PodMonitor).DeepCopy()
				if err := k8sutil.AddTypeInformationToObject(podMon); err != nil {
					level.Error(rs.l).Log("msg", "failed to set PodMonitor type information", "namespace", ns, "err", err)
					return
				}
				podMonitors[k] = podMon
			}
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list pod monitors in namespace %s: %w", ns, err)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1.PodMonitor, len(podMonitors))
	for namespaceAndName, pm := range podMonitors {
		var err error
		rejectFn := func(pm *monitoringv1.PodMonitor, err error) {
			rejected++
			level.Warn(rs.l).Log(
				"msg", "skipping podmonitor",
				"error", err.Error(),
				"podmonitor", namespaceAndName,
				"namespace", objMeta.GetNamespace(),
				"prometheus", objMeta.GetName(),
			)
			rs.eventRecorder.Eventf(pm, v1.EventTypeWarning, operator.InvalidConfigurationEvent, "PodMonitor %s was rejected due to invalid configuration: %v", pm.GetName(), err)
		}

		for i, endpoint := range pm.Spec.PodMetricsEndpoints {
			pmKey := fmt.Sprintf("podMonitor/%s/%s/%d", pm.GetNamespace(), pm.GetName(), i)

			//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
			if err = rs.store.AddBearerToken(ctx, pm.GetNamespace(), &endpoint.BearerTokenSecret, pmKey); err != nil {
				rejectFn(pm, err)
				break
			}

			if err = rs.store.AddBasicAuth(ctx, pm.GetNamespace(), endpoint.BasicAuth, pmKey); err != nil {
				rejectFn(pm, err)
				break
			}

			if endpoint.TLSConfig != nil {
				if err = rs.store.AddSafeTLSConfig(ctx, pm.GetNamespace(), endpoint.TLSConfig); err != nil {
					rejectFn(pm, err)
					break
				}
			}

			if err = rs.store.AddOAuth2(ctx, pm.GetNamespace(), endpoint.OAuth2, pmKey); err != nil {
				rejectFn(pm, err)
				break
			}

			pmAuthKey := fmt.Sprintf("podMonitor/auth/%s/%s/%d", pm.GetNamespace(), pm.GetName(), i)
			if err = rs.store.AddSafeAuthorizationCredentials(ctx, pm.GetNamespace(), endpoint.Authorization, pmAuthKey); err != nil {
				rejectFn(pm, err)
				break
			}

			if err = validateScrapeIntervalAndTimeout(rs.p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
				rejectFn(pm, err)
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.RelabelConfigs); err != nil {
				rejectFn(pm, fmt.Errorf("relabelConfigs: %w", err))
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.MetricRelabelConfigs); err != nil {
				rejectFn(pm, fmt.Errorf("metricRelabelConfigs: %w", err))
				break
			}

			if err = validateProxyURL(endpoint.ProxyURL); err != nil {
				rejectFn(pm, fmt.Errorf("proxyURL: %w", err))
				break
			}
		}

		if err != nil {
			continue
		}

		if err = validateScrapeClass(rs.p, pm.Spec.ScrapeClassName); err != nil {
			rejectFn(pm, err)
			continue
		}

		res[namespaceAndName] = pm
	}

	pmKeys := []string{}
	for k := range res {
		pmKeys = append(pmKeys, k)
	}
	level.Debug(rs.l).Log("msg", "selected PodMonitors", "podmonitors", strings.Join(pmKeys, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	if pKey, ok := rs.accessor.MetaNamespaceKey(rs.p); ok {
		rs.metrics.SetSelectedResources(pKey, monitoringv1.PodMonitorsKind, len(res))
		rs.metrics.SetRejectedResources(pKey, monitoringv1.PodMonitorsKind, rejected)
	}

	return res, nil
}

// SelectProbes selects Probes based on the selectors in the Prometheus CR and filters them
// returning only those with a valid configuration. This function also populates authentication stores and performs validations against
// scrape intervals, relabel configs and Probe URLs.
func (rs *ResourceSelector) SelectProbes(ctx context.Context, listFn ListAllByNamespaceFn) (map[string]*monitoringv1.Probe, error) {
	cpf := rs.p.GetCommonPrometheusFields()
	objMeta := rs.p.GetObjectMeta()
	namespaces := []string{}
	// Selectors might overlap. Deduplicate them along the keyFunc.
	probes := make(map[string]*monitoringv1.Probe)

	bMonSelector, err := metav1.LabelSelectorAsSelector(cpf.ProbeSelector)
	if err != nil {
		return nil, err
	}

	// If 'ProbeNamespaceSelector' is nil only check own namespace.
	if cpf.ProbeNamespaceSelector == nil {
		namespaces = append(namespaces, objMeta.GetNamespace())
	} else {
		bMonNSSelector, err := metav1.LabelSelectorAsSelector(cpf.ProbeNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = operator.ListMatchingNamespaces(bMonNSSelector, rs.namespaceInformers)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(rs.l).Log("msg", "filtering namespaces to select Probes from", "namespaces", strings.Join(namespaces, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	for _, ns := range namespaces {
		err := listFn(ns, bMonSelector, func(obj interface{}) {
			if k, ok := rs.accessor.MetaNamespaceKey(obj); ok {
				probe := obj.(*monitoringv1.Probe).DeepCopy()
				if err := k8sutil.AddTypeInformationToObject(probe); err != nil {
					level.Error(rs.l).Log("msg", "failed to set Probe type information", "namespace", ns, "err", err)
					return
				}
				probes[k] = probe
			}
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list probes in namespace %s: %w", ns, err)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1.Probe, len(probes))

	for probeName, probe := range probes {
		rejectFn := func(probe *monitoringv1.Probe, err error) {
			rejected++
			level.Warn(rs.l).Log(
				"msg", "skipping probe",
				"error", err.Error(),
				"probe", probeName,
				"namespace", objMeta.GetNamespace(),
				"prometheus", objMeta.GetName(),
			)
			rs.eventRecorder.Eventf(probe, v1.EventTypeWarning, operator.InvalidConfigurationEvent, "Probe %s was rejected due to invalid configuration: %v", probe.GetName(), err)
		}

		if err = validateScrapeClass(rs.p, probe.Spec.ScrapeClassName); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = probe.Spec.Targets.Validate(); err != nil {
			rejectFn(probe, err)
			continue
		}

		pnKey := fmt.Sprintf("probe/%s/%s", probe.GetNamespace(), probe.GetName())
		if err = rs.store.AddBearerToken(ctx, probe.GetNamespace(), &probe.Spec.BearerTokenSecret, pnKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = rs.store.AddBasicAuth(ctx, probe.GetNamespace(), probe.Spec.BasicAuth, pnKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if probe.Spec.TLSConfig != nil {
			if err = rs.store.AddSafeTLSConfig(ctx, probe.GetNamespace(), probe.Spec.TLSConfig); err != nil {
				rejectFn(probe, err)
				continue
			}
		}
		pnAuthKey := fmt.Sprintf("probe/auth/%s/%s", probe.GetNamespace(), probe.GetName())
		if err = rs.store.AddSafeAuthorizationCredentials(ctx, probe.GetNamespace(), probe.Spec.Authorization, pnAuthKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = rs.store.AddOAuth2(ctx, probe.GetNamespace(), probe.Spec.OAuth2, pnKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = validateScrapeIntervalAndTimeout(rs.p, probe.Spec.Interval, probe.Spec.ScrapeTimeout); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = validateRelabelConfigs(rs.p, probe.Spec.MetricRelabelConfigs); err != nil {
			err = fmt.Errorf("metricRelabelConfigs: %w", err)
			rejectFn(probe, err)
			continue
		}

		if probe.Spec.Targets.StaticConfig != nil {
			if err = validateRelabelConfigs(rs.p, probe.Spec.Targets.StaticConfig.RelabelConfigs); err != nil {
				err = fmt.Errorf("targets.staticConfig.relabelConfigs: %w", err)
				rejectFn(probe, err)
				continue
			}
		}

		if probe.Spec.Targets.Ingress != nil {
			if err = validateRelabelConfigs(rs.p, probe.Spec.Targets.Ingress.RelabelConfigs); err != nil {
				err = fmt.Errorf("targets.ingress.relabelConfigs: %w", err)
				rejectFn(probe, err)
				continue
			}
		}

		if err = validateProxyURL(&probe.Spec.ProberSpec.ProxyURL); err != nil {
			rejectFn(probe, fmt.Errorf("proxyURL: %w", err))
			continue
		}

		if err = validateProberURL(probe.Spec.ProberSpec.URL); err != nil {
			err := fmt.Errorf("%s url specified in proberSpec is invalid, it should be of the format `hostname` or `hostname:port`: %w", probe.Spec.ProberSpec.URL, err)
			rejectFn(probe, err)
			continue
		}

		res[probeName] = probe
	}

	probeKeys := make([]string, 0)
	for k := range res {
		probeKeys = append(probeKeys, k)
	}
	level.Debug(rs.l).Log("msg", "selected Probes", "probes", strings.Join(probeKeys, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	if pKey, ok := rs.accessor.MetaNamespaceKey(rs.p); ok {
		rs.metrics.SetSelectedResources(pKey, monitoringv1.ProbesKind, len(res))
		rs.metrics.SetRejectedResources(pKey, monitoringv1.ProbesKind, rejected)
	}

	return res, nil
}

func validateProxyURL(proxyurl *string) error {
	if proxyurl == nil {
		return nil
	}

	_, err := url.Parse(*proxyurl)
	return err
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

// SelectScrapeConfigs selects ScrapeConfigs based on the selectors in the Prometheus CR and filters them
// returning only those with a valid configuration.
func (rs *ResourceSelector) SelectScrapeConfigs(ctx context.Context, listFn ListAllByNamespaceFn) (map[string]*monitoringv1alpha1.ScrapeConfig, error) {
	cpf := rs.p.GetCommonPrometheusFields()
	objMeta := rs.p.GetObjectMeta()
	namespaces := []string{}

	// Selectors might overlap. Deduplicate them along the keyFunc.
	scrapeConfigs := make(map[string]*monitoringv1alpha1.ScrapeConfig)

	sConSelector, err := metav1.LabelSelectorAsSelector(cpf.ScrapeConfigSelector)
	if err != nil {
		return nil, err
	}

	// If 'ScrapeConfigNamespaceSelector' is nil only check own namespace.
	if cpf.ScrapeConfigNamespaceSelector == nil {
		namespaces = append(namespaces, objMeta.GetNamespace())
	} else {
		sConNSSelector, err := metav1.LabelSelectorAsSelector(cpf.ScrapeConfigNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = operator.ListMatchingNamespaces(sConNSSelector, rs.namespaceInformers)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(rs.l).Log("msg", "filtering namespaces to select ScrapeConfigs from", "namespaces", strings.Join(namespaces, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	for _, ns := range namespaces {
		err := listFn(ns, sConSelector, func(obj interface{}) {
			if k, ok := rs.accessor.MetaNamespaceKey(obj); ok {
				scrapeConfig := obj.(*monitoringv1alpha1.ScrapeConfig).DeepCopy()
				if err := k8sutil.AddTypeInformationToObject(scrapeConfig); err != nil {
					level.Error(rs.l).Log("msg", "failed to set ScrapeConfig type information", "namespace", ns, "err", err)
					return
				}
				scrapeConfigs[k] = scrapeConfig
			}
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list ScrapeConfigs in namespace %s: %w", ns, err)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1alpha1.ScrapeConfig, len(scrapeConfigs))

	for scName, sc := range scrapeConfigs {
		rejectFn := func(sc *monitoringv1alpha1.ScrapeConfig, err error) {
			rejected++
			level.Warn(rs.l).Log(
				"msg", "skipping scrapeconfig",
				"error", err.Error(),
				"scrapeconfig", scName,
				"namespace", objMeta.GetNamespace(),
				"prometheus", objMeta.GetName(),
			)
			rs.eventRecorder.Eventf(sc, v1.EventTypeWarning, operator.InvalidConfigurationEvent, "ScrapeConfig %s was rejected due to invalid configuration: %v", sc.GetName(), err)
		}

		if err = validateScrapeClass(rs.p, sc.Spec.ScrapeClassName); err != nil {
			rejectFn(sc, err)
			continue
		}

		if err = validateRelabelConfigs(rs.p, sc.Spec.RelabelConfigs); err != nil {
			rejectFn(sc, fmt.Errorf("relabelConfigs: %w", err))
			continue
		}

		scKey := fmt.Sprintf("scrapeconfig/%s/%s", sc.GetNamespace(), sc.GetName())
		if err = rs.store.AddBasicAuth(ctx, sc.GetNamespace(), sc.Spec.BasicAuth, scKey); err != nil {
			rejectFn(sc, err)
			continue
		}

		scAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s", sc.GetNamespace(), sc.GetName())
		if err = rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), sc.Spec.Authorization, scAuthKey); err != nil {
			rejectFn(sc, err)
			continue
		}

		if err = rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), sc.Spec.TLSConfig); err != nil {
			rejectFn(sc, err)
			continue
		}

		var scrapeInterval, scrapeTimeout monitoringv1.Duration = "", ""
		if sc.Spec.ScrapeInterval != nil {
			scrapeInterval = *sc.Spec.ScrapeInterval
		}

		if sc.Spec.ScrapeTimeout != nil {
			scrapeTimeout = *sc.Spec.ScrapeTimeout
		}

		if err = validateScrapeIntervalAndTimeout(rs.p, scrapeInterval, scrapeTimeout); err != nil {
			rejectFn(sc, err)
			continue
		}

		if err = validateProxyConfig(ctx, sc.Spec.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			rejectFn(sc, err)
			continue
		}

		if err = validateRelabelConfigs(rs.p, sc.Spec.MetricRelabelConfigs); err != nil {
			rejectFn(sc, fmt.Errorf("metricRelabelConfigs: %w", err))
			continue
		}

		if err = rs.validateHTTPSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("httpSDConfigs: %w", err))
			continue
		}

		if err = rs.validateKubernetesSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("kubernetesSDConfigs: %w", err))
			continue
		}

		if err = rs.validateConsulSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("consulSDConfigs: %w", err))
			continue
		}

		if err = rs.validateDNSSDConfigs(sc); err != nil {
			rejectFn(sc, fmt.Errorf("dnsSDConfigs: %w", err))
			continue
		}

		if err = rs.validateEC2SDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("ec2SDConfigs: %w", err))
			continue
		}

		if err = rs.validateAzureSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("azureSDConfigs: %w", err))
			continue
		}

		if err = rs.validateOpenStackSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("openstackSDConfigs: %w", err))
			continue
		}

		if err = rs.validateDigitalOceanSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("digitalOceanSDConfigs: %w", err))
			continue
		}

		if err = rs.validateKumaSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("kumaSDConfigs: %w", err))
			continue
		}

		if err = rs.validateEurekaSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("eurekaSDConfigs: %w", err))
			continue
		}

		if err = rs.validateDockerSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("dockerSDConfigs: %w", err))
			continue
		}
		if err = rs.validateHetznerSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("hetznerSDConfigs: %w", err))
			continue
		}
		res[scName] = sc
	}

	scrapeConfigKeys := make([]string, 0)
	for k := range res {
		scrapeConfigKeys = append(scrapeConfigKeys, k)
	}
	level.Debug(rs.l).Log("msg", "selected ScrapeConfigs", "scrapeConfig", strings.Join(scrapeConfigKeys, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	if sKey, ok := rs.accessor.MetaNamespaceKey(rs.p); ok {
		rs.metrics.SetSelectedResources(sKey, monitoringv1alpha1.ScrapeConfigsKind, len(res))
		rs.metrics.SetRejectedResources(sKey, monitoringv1alpha1.ScrapeConfigsKind, rejected)
	}

	return res, nil
}

func (rs *ResourceSelector) validateKubernetesSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.KubernetesSDConfigs {
		configKey := fmt.Sprintf("scrapeconfig/%s/%s/kubernetessdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/kubernetessdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
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
			if _, err := fields.ParseSelector(s.Field); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}

			if _, err := labels.Parse(s.Label); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
	}
	return nil
}

func (rs *ResourceSelector) validateConsulSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.ConsulSDConfigs {
		configKey := fmt.Sprintf("scrapeconfig/%s/%s/consulsdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/consulsdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
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

		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func (rs *ResourceSelector) validateHTTPSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.HTTPSDConfigs {
		configKey := fmt.Sprintf("scrapeconfig/%s/%s/httpsdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/httpsdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
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
	}
	return nil
}

func (rs *ResourceSelector) validateAzureSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.AzureSDConfigs {
		// Since Prometheus uses default authentication method as "OAuth"
		if ptr.Deref(config.AuthenticationMethod, "") == "ManagedIdentity" {
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
	}
	return nil
}

func (rs *ResourceSelector) validateOpenStackSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.OpenStackSDConfigs {
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
	for i, config := range sc.Spec.DigitalOceanSDConfigs {
		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/digitaloceansdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		configKey := fmt.Sprintf("scrapeconfig/%s/%s/digitaloceansdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func (rs *ResourceSelector) validateDockerSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.DockerSDConfigs {
		configKey := fmt.Sprintf("scrapeconfig/%s/%s/dockersdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/dockersdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		configOAuthKey := fmt.Sprintf("scrapeconfig/%s/%s/dockersdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2, configOAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		// Validate the host daemon address url
		if _, err := url.Parse(config.Host); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
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

		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/kumasdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		configKey := fmt.Sprintf("scrapeconfig/%s/%s/kumasdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func (rs *ResourceSelector) validateEurekaSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.EurekaSDConfigs {
		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/eurekasdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		configKey := fmt.Sprintf("scrapeconfig/%s/%s/eurekasdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func (rs *ResourceSelector) validateHetznerSDConfigs(ctx context.Context, sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.HetznerSDConfigs {
		configKey := fmt.Sprintf("scrapeconfig/%s/%s/hetznersdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddBasicAuth(ctx, sc.GetNamespace(), config.BasicAuth, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		configAuthKey := fmt.Sprintf("scrapeconfig/auth/%s/%s/hetznersdconfig/%d", sc.GetNamespace(), sc.GetName(), i)
		if err := rs.store.AddSafeAuthorizationCredentials(ctx, sc.GetNamespace(), config.Authorization, configAuthKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if err := rs.store.AddOAuth2(ctx, sc.GetNamespace(), config.OAuth2, configKey); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := rs.store.AddSafeTLSConfig(ctx, sc.GetNamespace(), config.TLSConfig); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if err := validateProxyConfig(ctx, config.ProxyConfig, rs.store, sc.GetNamespace()); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}
