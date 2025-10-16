// Copyright 2022 The prometheus-operator Authors
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

package operator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	sortutil "github.com/prometheus-operator/prometheus-operator/internal/sortutil"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
)

type RuleConfigurationFormat int

const (
	// PrometheusFormat indicates that the rule configuration should comply with the Prometheus format.
	PrometheusFormat RuleConfigurationFormat = iota
	// ThanosFormat indicates that the rule configuration should comply with the Thanos format.
	ThanosFormat
)

const (
	selectingPrometheusRuleResourcesAction = "SelectingPrometheusRuleResources"
)

// MaxConfigMapDataSize represents the maximum size for ConfigMap's data.  The
// maximum `Data` size of a ConfigMap seems to differ between environments.
// This is probably due to different meta data sizes which count into the
// overall maximum size of a ConfigMap. Thereby lets leave a large buffer.
var MaxConfigMapDataSize = int(float64(v1.MaxSecretSize) * 0.5)

// PrometheusRuleSelector selects PrometheusRule resources and translates them
// to Prometheus/Thanos configuration format.
type PrometheusRuleSelector struct {
	ruleFormat   RuleConfigurationFormat
	version      semver.Version
	ruleSelector labels.Selector
	nsLabeler    *namespacelabeler.Labeler
	ruleInformer *informers.ForResource

	eventRecorder *EventRecorder

	logger *slog.Logger
}

type PrometheusRuleSelection struct {
	selection TypedResourcesSelection[*monitoringv1.PrometheusRule] // PrometheusRules selected.
	ruleFiles map[string]string                                     // Map of rule configuration files serialized to the Prometheus format (key=filename).
}

func (prs *PrometheusRuleSelection) RuleFiles() map[string]string {
	return prs.ruleFiles
}

func (prs *PrometheusRuleSelection) Selected(log *slog.Logger) TypedResourcesSelection[*monitoringv1.PrometheusRule] {
	selected := make(TypedResourcesSelection[*monitoringv1.PrometheusRule], len(prs.selection))

	accessor := NewAccessor(log)
	for _, promRule := range prs.selection {
		k, ok := accessor.MetaNamespaceKey(promRule)
		if !ok {
			continue
		}

		selected[k] = promRule
	}
	return selected
}

func (prs *PrometheusRuleSelection) SelectedLen() int {
	return len(prs.selection)
}

func (prs *PrometheusRuleSelection) RejectedLen() int {
	return len(prs.selection) - len(prs.ruleFiles)
}

// NewPrometheusRuleSelector returns a PrometheusRuleSelector pointer.
func NewPrometheusRuleSelector(ruleFormat RuleConfigurationFormat, version string, labelSelector *metav1.LabelSelector, nsLabeler *namespacelabeler.Labeler, ruleInformer *informers.ForResource, eventRecorder *EventRecorder, logger *slog.Logger) (*PrometheusRuleSelector, error) {
	componentVersion, err := semver.ParseTolerant(version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	ruleSelector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("convert rule label selector to selector: %w", err)
	}

	return &PrometheusRuleSelector{
		ruleFormat:    ruleFormat,
		version:       componentVersion,
		ruleSelector:  ruleSelector,
		nsLabeler:     nsLabeler,
		ruleInformer:  ruleInformer,
		eventRecorder: eventRecorder,
		logger:        logger,
	}, nil
}

func (prs *PrometheusRuleSelector) generateRulesConfiguration(promRule *monitoringv1.PrometheusRule) (string, error) {
	logger := prs.logger.With("prometheusrule", promRule.Name, "prometheusrule-namespace", promRule.Namespace)
	promRuleSpec := promRule.Spec

	promRuleSpec = prs.sanitizePrometheusRulesSpec(promRuleSpec, logger)

	content, err := yaml.Marshal(promRuleSpec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	var validationScheme model.ValidationScheme
	if prs.ruleFormat == ThanosFormat {
		validationScheme = ValidationSchemeForThanos(prs.version)
	} else {
		validationScheme = ValidationSchemeForPrometheus(prs.version)
	}

	errs := ValidateRule(promRuleSpec, validationScheme)
	if len(errs) != 0 {
		const m = "invalid rule"
		logger.Debug(m, "content", content)
		for _, err := range errs {
			logger.Info(m, "err", err)
		}
		return "", errors.New(m)
	}

	return string(content), nil
}

// sanitizePrometheusRulesSpec sanitizes the PrometheusRules spec depending on the Prometheus/Thanos version.
func (prs *PrometheusRuleSelector) sanitizePrometheusRulesSpec(promRuleSpec monitoringv1.PrometheusRuleSpec, logger *slog.Logger) monitoringv1.PrometheusRuleSpec {
	minVersionKeepFiringFor := semver.MustParse("2.42.0")
	minVersionLimits := semver.MustParse("2.31.0")
	minVersionQueryOffset := semver.MustParse("2.53.0")
	minVersionRuleGroupLabels := semver.MustParse("3.0.0")
	component := "Prometheus"

	if prs.ruleFormat == ThanosFormat {
		minVersionKeepFiringFor = semver.MustParse("0.34.0")
		minVersionLimits = semver.MustParse("0.24.0")
		minVersionQueryOffset = semver.MustParse("100.0.0")     // Arbitrary very high major version because it's not yet supported by Thanos.
		minVersionRuleGroupLabels = semver.MustParse("100.0.0") // Arbitrary very high major version because it's not yet supported by Thanos.
		component = "Thanos"
	}

	for i := range promRuleSpec.Groups {
		if promRuleSpec.Groups[i].Limit != nil && prs.version.LT(minVersionLimits) {
			promRuleSpec.Groups[i].Limit = nil
			logger.Warn(fmt.Sprintf("ignoring `limit` not supported by %s", component), "minimum_version", minVersionLimits)
		}

		if promRuleSpec.Groups[i].QueryOffset != nil && prs.version.LT(minVersionQueryOffset) {
			promRuleSpec.Groups[i].QueryOffset = nil
			logger.Warn(fmt.Sprintf("ignoring `query_offset` not supported by %s", component), "minimum_version", minVersionQueryOffset)
		}

		if prs.ruleFormat == PrometheusFormat {
			// Unset partialResponseStrategy field.
			promRuleSpec.Groups[i].PartialResponseStrategy = ""
		}

		if len(promRuleSpec.Groups[i].Labels) > 0 && prs.version.LT(minVersionRuleGroupLabels) {
			promRuleSpec.Groups[i].Labels = nil
			logger.Warn(fmt.Sprintf("ignoring group labels since not supported by %s", component), "minimum_version", minVersionRuleGroupLabels)
		}

		for j := range promRuleSpec.Groups[i].Rules {
			if promRuleSpec.Groups[i].Rules[j].KeepFiringFor != nil && prs.version.LT(minVersionKeepFiringFor) {
				promRuleSpec.Groups[i].Rules[j].KeepFiringFor = nil
				logger.Warn(fmt.Sprintf("ignoring 'keep_firing_for' not supported by %s", component), "minimum_version", minVersionKeepFiringFor)
			}
		}
	}

	return promRuleSpec
}

// ValidateRule takes PrometheusRuleSpec and validates it using the upstream prometheus rule validator.
func ValidateRule(promRuleSpec monitoringv1.PrometheusRuleSpec, validationScheme model.ValidationScheme) []error {
	for i := range promRuleSpec.Groups {
		// The upstream Prometheus rule validator doesn't support the
		// partial_response_strategy field.
		promRuleSpec.Groups[i].PartialResponseStrategy = ""

		// Empty durations need to be translated to nil to be omitted from the
		// YAML output otherwise the generated configuration will not be valid.
		if promRuleSpec.Groups[i].Interval != nil && *promRuleSpec.Groups[i].Interval == "" {
			promRuleSpec.Groups[i].Interval = nil
		}

		for j := range promRuleSpec.Groups[i].Rules {
			if ptr.Deref(promRuleSpec.Groups[i].Rules[j].For, "") == "" {
				promRuleSpec.Groups[i].Rules[j].For = nil
			}
		}
	}

	content, err := yaml.Marshal(promRuleSpec)
	if err != nil {
		return []error{fmt.Errorf("failed to marshal content: %w", err)}
	}

	// Check if the serialized rules exceed our internal limit.
	promRuleSize := len(content)
	if promRuleSize > MaxConfigMapDataSize {
		return []error{fmt.Errorf("the length of rendered Prometheus Rule is %d bytes which is above the maximum limit of %d bytes", promRuleSize, MaxConfigMapDataSize)}
	}

	_, errs := rulefmt.Parse(content, false, validationScheme)
	return errs
}

// Select selects PrometheusRules by Prometheus or ThanosRuler.
func (prs *PrometheusRuleSelector) Select(namespaces []string) (PrometheusRuleSelection, error) {
	promRules := map[string]*monitoringv1.PrometheusRule{}

	for _, ns := range namespaces {
		err := prs.ruleInformer.ListAllByNamespace(ns, prs.ruleSelector, func(obj any) {
			promRule := obj.(*monitoringv1.PrometheusRule).DeepCopy()
			if err := k8sutil.AddTypeInformationToObject(promRule); err != nil {
				prs.logger.Error("failed to set rule type information", "namespace", ns, "err", err)
				return
			}

			// Generate a truly unique identifier for each PrometheusRule resource.
			// We use the UID to avoid collisions between foo-bar/fred and foo/bar-fred.
			promRules[fmt.Sprintf("%v-%v-%v.yaml", promRule.Namespace, promRule.Name, promRule.UID)] = promRule
		})
		if err != nil {
			return PrometheusRuleSelection{}, fmt.Errorf("failed to list PrometheusRule objects in namespace %s: %w", ns, err)
		}
	}

	var (
		marshalRules    = make(map[string]string, len(promRules))
		rules           = make(TypedResourcesSelection[*monitoringv1.PrometheusRule], len(promRules))
		namespacedNames = make([]string, 0, len(promRules))
	)
	for ruleName, promRule := range promRules {
		var err error
		var content string
		if err := prs.nsLabeler.EnforceNamespaceLabel(promRule); err != nil {
			continue
		}

		content, err = prs.generateRulesConfiguration(promRule)
		if err != nil {
			prs.logger.Warn(
				"skipping prometheusrule",
				"error", err.Error(),
				"prometheusrule", promRule.Name,
				"namespace", promRule.Namespace,
			)
			prs.eventRecorder.Eventf(promRule, v1.EventTypeWarning, InvalidConfigurationEvent, selectingPrometheusRuleResourcesAction, "PrometheusRule %s was rejected due to invalid configuration: %v", promRule.Name, err)
		}

		var reason string
		if err != nil {
			reason = InvalidConfigurationEvent
		}

		rules[ruleName] = TypedConfigurationResource[*monitoringv1.PrometheusRule]{
			resource:   promRule,
			err:        err,
			reason:     reason,
			generation: promRule.GetGeneration(),
		}
		if err == nil {
			marshalRules[ruleName] = content
			namespacedNames = append(namespacedNames, fmt.Sprintf("%s/%s", promRule.Namespace, promRule.Name))
		}
	}

	slices.Sort(namespacedNames)

	prs.logger.Debug(
		"selected Rules",
		"rules", strings.Join(namespacedNames, ","),
	)

	return PrometheusRuleSelection{
		selection: rules,
		ruleFiles: marshalRules,
	}, nil
}

// PrometheusRuleSyncer knows how to synchronize ConfigMaps holding
// Prometheus or Thanos rule configuration.
type PrometheusRuleSyncer struct {
	opts       []ObjectOption
	cmClient   clientv1.ConfigMapInterface
	cmSelector labels.Set

	logger *slog.Logger
}

func NewPrometheusRuleSyncer(
	logger *slog.Logger,
	cmClient clientv1.ConfigMapInterface,
	cmSelector labels.Set,
	options []ObjectOption,
) *PrometheusRuleSyncer {
	return &PrometheusRuleSyncer{
		logger:     logger,
		cmClient:   cmClient,
		cmSelector: cmSelector,
		opts:       options,
	}
}

// Sync synchronizes the ConfigMap(s) holding the provided list of rules.
// It returns the list of ConfigMap names.
func (prs *PrometheusRuleSyncer) Sync(ctx context.Context, rules map[string]string) ([]string, error) {
	listConfigMapOpts := metav1.ListOptions{LabelSelector: prs.cmSelector.String()}
	cmList, err := prs.cmClient.List(ctx, listConfigMapOpts)
	if err != nil {
		return nil, err
	}

	current := cmList.Items
	currentRules := map[string]string{}
	for _, cm := range current {
		maps.Copy(currentRules, cm.Data)
	}

	if equal := reflect.DeepEqual(rules, currentRules); equal && len(current) != 0 {
		// Return early because current and generated configmaps are identical
		// (and at least 1 configmap already exists).
		prs.logger.Debug("no PrometheusRule changes")
		return configMapNames(current), nil
	}

	configMaps, err := prs.makeConfigMapsFromRules(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ConfigMaps for PrometheusRule: %w", err)
	}

	// Delete and recreate the configmap(s). It's not very efficient but proved
	// to be robust enough so far.
	if len(current) > 0 {
		prs.logger.Debug("deleting ConfigMaps for PrometheusRule")
		// In theory we could use DeleteCollection but the method isn't
		// supported by the fake client.
		// See https://github.com/kubernetes/kubernetes/issues/105357
		for _, cm := range current {
			err := prs.cmClient.Delete(ctx, cm.Name, metav1.DeleteOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to delete ConfigMap %q: %w", cm.Name, err)
			}
		}
	}

	prs.logger.Debug("creating ConfigMaps for PrometheusRule")
	for _, cm := range configMaps {
		_, err = prs.cmClient.Create(ctx, &cm, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create ConfigMap %q: %w", cm.Name, err)
		}
	}

	return configMapNames(configMaps), nil
}

func configMapNames(cms []v1.ConfigMap) []string {
	return slices.Sorted(slices.Values(slices.Collect(func(yield func(string) bool) {
		for _, cm := range cms {
			if !yield(cm.Name) {
				return
			}
		}
	})))
}

type bucket struct {
	rules map[string]string
	size  int
}

// makeConfigMapsFromRules takes a map of rule files (keys: filenames) and returns
// a list of Kubernetes ConfigMaps to be later on mounted into the a container.
//
// If the total size of rule files exceeds the Kubernetes ConfigMap limit,
// they are split up via the simple first-fit [1] bin packing algorithm. In the
// future this can be replaced by a more sophisticated algorithm, but for now
// simplicity should be sufficient.
//
// [1] https://en.wikipedia.org/wiki/Bin_packing_problem#First-fit_algorithm
func (prs *PrometheusRuleSyncer) makeConfigMapsFromRules(rules map[string]string) ([]v1.ConfigMap, error) {
	var (
		i       int
		buckets = []bucket{{rules: map[string]string{}}}
	)
	// To make bin packing algorithm deterministic, sort ruleFiles filenames
	// and iterate over the filenames instead of the key/value pairs.
	for _, filename := range sortutil.SortedKeys(rules) {
		// If the rule file doesn't fit into the current bucket, create a new bucket.
		if (buckets[i].size + len(rules[filename])) > MaxConfigMapDataSize {
			buckets = append(buckets, bucket{rules: map[string]string{}})
			i++
		}

		buckets[i].rules[filename] = rules[filename]
		buckets[i].size += len(rules[filename])
	}

	configMaps := make([]v1.ConfigMap, 0, len(buckets))
	for i, bucket := range buckets {
		cm := v1.ConfigMap{
			Data: bucket.rules,
		}

		UpdateObject(
			&cm,
			prs.opts...,
		)

		UpdateObject(
			&cm,
			WithLabels(prs.cmSelector),
		)

		// Ensure that the ConfigMap's names are unique.
		cm.Name = fmt.Sprintf("%s-rulefiles-%d", cm.Name, i)

		configMaps = append(configMaps, cm)
	}

	return configMaps, nil
}
