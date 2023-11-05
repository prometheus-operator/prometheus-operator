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
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

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
}

type ListAllByNamespaceFn func(namespace string, selector labels.Selector, appendFn cache.AppendFunc) error

func NewResourceSelector(l log.Logger, p monitoringv1.PrometheusInterface, store *assets.Store, namespaceInformers cache.SharedIndexInformer, metrics *operator.Metrics) *ResourceSelector {
	return &ResourceSelector{
		l:                  l,
		p:                  p,
		store:              store,
		namespaceInformers: namespaceInformers,
		metrics:            metrics,
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

		for i, endpoint := range sm.Spec.Endpoints {
			// If denied by Prometheus spec, filter out all service monitors that access
			// the file system.
			if cpf.ArbitraryFSAccessThroughSMs.Deny {
				if err = testForArbitraryFSAccess(endpoint); err != nil {
					break
				}
			}

			smKey := fmt.Sprintf("serviceMonitor/%s/%s/%d", sm.GetNamespace(), sm.GetName(), i)

			if err = rs.store.AddBearerToken(ctx, sm.GetNamespace(), endpoint.BearerTokenSecret, smKey); err != nil {
				break
			}

			if err = rs.store.AddBasicAuth(ctx, sm.GetNamespace(), endpoint.BasicAuth, smKey); err != nil {
				break
			}

			if endpoint.TLSConfig != nil {
				if err = rs.store.AddTLSConfig(ctx, sm.GetNamespace(), endpoint.TLSConfig); err != nil {
					break
				}
			}

			if err = rs.store.AddOAuth2(ctx, sm.GetNamespace(), endpoint.OAuth2, smKey); err != nil {
				break
			}

			smAuthKey := fmt.Sprintf("serviceMonitor/auth/%s/%s/%d", sm.GetNamespace(), sm.GetName(), i)
			if err = rs.store.AddSafeAuthorizationCredentials(ctx, sm.GetNamespace(), endpoint.Authorization, smAuthKey); err != nil {
				break
			}

			if err = validateScrapeIntervalAndTimeout(rs.p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.RelabelConfigs); err != nil {
				err = fmt.Errorf("relabelConfigs: %w", err)
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.MetricRelabelConfigs); err != nil {
				err = fmt.Errorf("metricRelabelConfigs: %w", err)
				break
			}
		}

		if err != nil {
			rejected++
			level.Warn(rs.l).Log(
				"msg", "skipping servicemonitor",
				"error", err.Error(),
				"servicemonitor", namespaceAndName,
				"namespace", objMeta.GetNamespace(),
				"prometheus", objMeta.GetName(),
			)
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
			!(rc.Separator == "" ||
				rc.Separator == relabel.DefaultRelabelConfig.Separator) ||
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
			!(rc.Separator == "" ||
				rc.Separator == relabel.DefaultRelabelConfig.Separator) ||
			!(rc.Replacement == relabel.DefaultRelabelConfig.Replacement ||
				rc.Replacement == "") {
			return fmt.Errorf("%s action requires only 'regex', and no other fields", rc.Action)
		}
	}
	return nil
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

		for i, endpoint := range pm.Spec.PodMetricsEndpoints {
			pmKey := fmt.Sprintf("podMonitor/%s/%s/%d", pm.GetNamespace(), pm.GetName(), i)

			if err = rs.store.AddBearerToken(ctx, pm.GetNamespace(), &endpoint.BearerTokenSecret, pmKey); err != nil {
				break
			}

			if err = rs.store.AddBasicAuth(ctx, pm.GetNamespace(), endpoint.BasicAuth, pmKey); err != nil {
				break
			}

			if endpoint.TLSConfig != nil {
				if err = rs.store.AddSafeTLSConfig(ctx, pm.GetNamespace(), &endpoint.TLSConfig.SafeTLSConfig); err != nil {
					break
				}
			}

			if err = rs.store.AddOAuth2(ctx, pm.GetNamespace(), endpoint.OAuth2, pmKey); err != nil {
				break
			}

			pmAuthKey := fmt.Sprintf("podMonitor/auth/%s/%s/%d", pm.GetNamespace(), pm.GetName(), i)
			if err = rs.store.AddSafeAuthorizationCredentials(ctx, pm.GetNamespace(), endpoint.Authorization, pmAuthKey); err != nil {
				break
			}

			if err = validateScrapeIntervalAndTimeout(rs.p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.RelabelConfigs); err != nil {
				err = fmt.Errorf("relabelConfigs: %w", err)
				break
			}

			if err = validateRelabelConfigs(rs.p, endpoint.MetricRelabelConfigs); err != nil {
				err = fmt.Errorf("metricRelabelConfigs: %w", err)
				break
			}
		}

		if err != nil {
			rejected++
			level.Warn(rs.l).Log(
				"msg", "skipping podmonitor",
				"error", err.Error(),
				"podmonitor", namespaceAndName,
				"namespace", objMeta.GetNamespace(),
				"prometheus", objMeta.GetName(),
			)
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
			if err = rs.store.AddSafeTLSConfig(ctx, probe.GetNamespace(), &probe.Spec.TLSConfig.SafeTLSConfig); err != nil {
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

		if err = validateRelabelConfigs(rs.p, sc.Spec.MetricRelabelConfigs); err != nil {
			rejectFn(sc, fmt.Errorf("metricRelabelConfigs: %w", err))
			continue
		}

		if err = rs.validateHTTPSDConfigs(ctx, sc); err != nil {
			rejectFn(sc, fmt.Errorf("httpSDConfigs: %w", err))
			continue
		}

		if err = rs.validateKubernetesSDConfigs(sc); err != nil {
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

func (rs *ResourceSelector) validateKubernetesSDConfigs(sc *monitoringv1alpha1.ScrapeConfig) error {
	for i, config := range sc.Spec.KubernetesSDConfigs {
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

		for k, v := range config.ProxyConnectHeader {
			if _, err := rs.store.GetSecretKey(context.Background(), sc.GetNamespace(), v); err != nil {
				return fmt.Errorf("[%d]: header[%s]: %w", i, k, err)
			}
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
