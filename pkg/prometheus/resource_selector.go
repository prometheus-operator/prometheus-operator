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
	"fmt"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			return nil, errors.Wrapf(err, "failed to list service monitors in namespace %s", ns)
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

			for _, rl := range endpoint.RelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(rs.p, *rl); err != nil {
						break
					}
				}
			}

			for _, rl := range endpoint.MetricRelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(rs.p, *rl); err != nil {
						break
					}
				}
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

func validateRelabelConfig(p monitoringv1.PrometheusInterface, rc monitoringv1.RelabelConfig) error {
	relabelTarget := regexp.MustCompile(`^(?:(?:[a-zA-Z_]|\$(?:\{\w+\}|\w+))+\w*)+$`)
	promVersion := operator.StringValOrDefault(p.GetCommonPrometheusFields().Version, operator.DefaultPrometheusVersion)
	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse Prometheus version")
	}
	minimumVersionCaseActions := version.GTE(semver.MustParse("2.36.0"))
	minimumVersionEqualActions := version.GTE(semver.MustParse("2.41.0"))

	if (rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase)) && !minimumVersionCaseActions {
		return errors.Errorf("%s relabel action is only supported from Prometheus version 2.36.0", rc.Action)
	}

	if (rc.Action == string(relabel.KeepEqual) || rc.Action == string(relabel.DropEqual)) && !minimumVersionEqualActions {
		return errors.Errorf("%s relabel action is only supported from Prometheus version 2.41.0", rc.Action)
	}

	if _, err := relabel.NewRegexp(rc.Regex); err != nil {
		return errors.Wrapf(err, "invalid regex %s for relabel configuration", rc.Regex)
	}

	if rc.Modulus == 0 && rc.Action == string(relabel.HashMod) {
		return errors.Errorf("relabel configuration for hashmod requires non-zero modulus")
	}

	if (rc.Action == string(relabel.Replace) || rc.Action == string(relabel.HashMod) || rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase) || rc.Action == string(relabel.KeepEqual) || rc.Action == string(relabel.DropEqual)) && rc.TargetLabel == "" {
		return errors.Errorf("relabel configuration for %s action needs targetLabel value", rc.Action)
	}

	if (rc.Action == string(relabel.Replace) || rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase) || rc.Action == string(relabel.KeepEqual) || rc.Action == string(relabel.DropEqual)) && !relabelTarget.MatchString(rc.TargetLabel) {
		return errors.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if (rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase) || rc.Action == string(relabel.KeepEqual) || rc.Action == string(relabel.DropEqual)) && !(rc.Replacement == relabel.DefaultRelabelConfig.Replacement || rc.Replacement == "") {
		return errors.Errorf("'replacement' can not be set for %s action", rc.Action)
	}

	if rc.Action == string(relabel.LabelMap) {
		if rc.Replacement != "" && !relabelTarget.MatchString(rc.Replacement) {
			return errors.Errorf("%q is invalid 'replacement' for %s action", rc.Replacement, rc.Action)
		}
	}

	if rc.Action == string(relabel.HashMod) && !model.LabelName(rc.TargetLabel).IsValid() {
		return errors.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if rc.Action == string(relabel.KeepEqual) || rc.Action == string(relabel.DropEqual) {
		if !(rc.Regex == "" || rc.Regex == relabel.DefaultRelabelConfig.Regex.String()) ||
			!(rc.Modulus == uint64(0) ||
				rc.Modulus == relabel.DefaultRelabelConfig.Modulus) ||
			!(rc.Separator == "" ||
				rc.Separator == relabel.DefaultRelabelConfig.Separator) ||
			!(rc.Replacement == relabel.DefaultRelabelConfig.Replacement ||
				rc.Replacement == "") {
			return errors.Errorf("%s action requires only 'source_labels' and `target_label`, and no other fields", rc.Action)
		}
	}

	if rc.Action == string(relabel.LabelDrop) || rc.Action == string(relabel.LabelKeep) {
		if len(rc.SourceLabels) != 0 ||
			!(rc.TargetLabel == "" ||
				rc.TargetLabel == relabel.DefaultRelabelConfig.TargetLabel) ||
			!(rc.Modulus == uint64(0) ||
				rc.Modulus == relabel.DefaultRelabelConfig.Modulus) ||
			!(rc.Separator == "" ||
				rc.Separator == relabel.DefaultRelabelConfig.Separator) ||
			!(rc.Replacement == relabel.DefaultRelabelConfig.Replacement ||
				rc.Replacement == "") {
			return errors.Errorf("%s action requires only 'regex', and no other fields", rc.Action)
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
			return nil, errors.Wrapf(err, "failed to list pod monitors in namespace %s", ns)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1.PodMonitor, len(podMonitors))
	for namespaceAndName, pm := range podMonitors {
		var err error

		for i, endpoint := range pm.Spec.PodMetricsEndpoints {
			pmKey := fmt.Sprintf("podMonitor/%s/%s/%d", pm.GetNamespace(), pm.GetName(), i)

			if err = rs.store.AddBearerToken(ctx, pm.GetNamespace(), endpoint.BearerTokenSecret, pmKey); err != nil {
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

			for _, rl := range endpoint.RelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(rs.p, *rl); err != nil {
						break
					}
				}
			}

			for _, rl := range endpoint.MetricRelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(rs.p, *rl); err != nil {
						break
					}
				}
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
			return nil, errors.Wrapf(err, "failed to list probes in namespace %s", ns)
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
				"probe", probe,
				"namespace", objMeta.GetNamespace(),
				"prometheus", objMeta.GetName(),
			)
		}

		if err = probe.Spec.Targets.Validate(); err != nil {
			rejectFn(probe, err)
			continue
		}

		pnKey := fmt.Sprintf("probe/%s/%s", probe.GetNamespace(), probe.GetName())
		if err = rs.store.AddBearerToken(ctx, probe.GetNamespace(), probe.Spec.BearerTokenSecret, pnKey); err != nil {
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

		for _, rl := range probe.Spec.MetricRelabelConfigs {
			if rl.Action != "" {
				if err = validateRelabelConfig(rs.p, *rl); err != nil {
					rejectFn(probe, err)
					continue
				}
			}
		}
		if err = validateProberURL(probe.Spec.ProberSpec.URL); err != nil {
			err := errors.Wrapf(err, "%s url specified in proberSpec is invalid, it should be of the format `hostname` or `hostname:port`", probe.Spec.ProberSpec.URL)
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
		return errors.Errorf("invalid host: %q", hostPort[0])
	}

	// handling cases with url specified as host:port
	if len(hostPort) > 1 {
		if !govalidator.IsPort(hostPort[1]) {
			return errors.Errorf("invalid port: %q", hostPort[1])
		}
	}
	return nil
}

// SelectScrapeConfigs selects ScrapeConfigs based on the selectors in the Prometheus CR and filters them
// returning only those with a valid configuration.
func (rs *ResourceSelector) SelectScrapeConfigs(listFn ListAllByNamespaceFn) (map[string]*monitoringv1alpha1.ScrapeConfig, error) {
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
			return nil, errors.Wrapf(err, "failed to list ScrapeConfigs in namespace %s", ns)
		}
	}

	// TODO(xiu): add validation steps
	scrapeConfigKeys := make([]string, 0)
	for k := range scrapeConfigs {
		scrapeConfigKeys = append(scrapeConfigKeys, k)
	}
	level.Debug(rs.l).Log("msg", "selected ScrapeConfigs", "scrapeConfig", strings.Join(scrapeConfigKeys, ","), "namespace", objMeta.GetNamespace(), "prometheus", objMeta.GetName())

	if sKey, ok := rs.accessor.MetaNamespaceKey(rs.p); ok {
		rs.metrics.SetSelectedResources(sKey, monitoringv1alpha1.ScrapeConfigsKind, len(scrapeConfigs))
		// since we don't have validation steps, we don't reject anything
		rs.metrics.SetRejectedResources(sKey, monitoringv1alpha1.ScrapeConfigsKind, 0)
	}

	return scrapeConfigs, nil
}
