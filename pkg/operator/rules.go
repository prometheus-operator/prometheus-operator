// Copyright The prometheus-operator Authors
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

	metricsql "github.com/VictoriaMetrics/metricsql"
	"github.com/blang/semver/v4"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	sortutil "github.com/prometheus-operator/prometheus-operator/internal/sortutil"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
	namespacelabeler "github.com/prometheus-operator/prometheus-operator/pkg/namespacelabeler"
)

type RuleConfigurationFormat int

const (
	// PrometheusFormat indicates that the rule configuration should comply with the Prometheus format.
	PrometheusFormat RuleConfigurationFormat = iota
	// ThanosFormat indicates that the rule configuration should comply with the Thanos format.
	ThanosFormat
)

// ExpressionLanguage defines the query language used to validate rule expressions.
type ExpressionLanguage int

const (
	// PromQLLanguage uses the Prometheus PromQL parser for expression validation.
	// This is the default and is compatible with Prometheus and Thanos targets.
	PromQLLanguage ExpressionLanguage = iota
	// MetricsQLLanguage uses the VictoriaMetrics MetricsQL parser for expression validation.
	// MetricsQL is a superset of PromQL and is required for VictoriaMetrics targets.
	MetricsQLLanguage
)

const (
	selectingPrometheusRuleResourcesAction = "SelectingPrometheusRuleResources"
)

// MaxConfigMapDataSize represents the maximum size for ConfigMap's data.  The
// maximum `Data` size of a ConfigMap seems to differ between environments.
// This is probably due to different meta data sizes which count into the
// overall maximum size of a ConfigMap. Thereby lets leave a large buffer.
var MaxConfigMapDataSize = int(float64(corev1.MaxSecretSize) * 0.5)

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

func (prs *PrometheusRuleSelection) Selected() TypedResourcesSelection[*monitoringv1.PrometheusRule] {
	return prs.selection
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
		minVersionQueryOffset = semver.MustParse("0.38.0")
		minVersionRuleGroupLabels = semver.MustParse("0.39.0")
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
	return ValidateRuleWithExpressionLanguage(promRuleSpec, validationScheme, PromQLLanguage)
}

// ValidateRuleWithExpressionLanguage validates a PrometheusRuleSpec using the given expression language.
// Use PromQLLanguage (the default) for Prometheus and Thanos targets.
// Use MetricsQLLanguage for VictoriaMetrics targets, which accepts a superset of PromQL including
// functions such as share_eq_over_time, median_over_time, and other MetricsQL extensions.
func ValidateRuleWithExpressionLanguage(promRuleSpec monitoringv1.PrometheusRuleSpec, validationScheme model.ValidationScheme, exprLang ExpressionLanguage) []error {
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

	if exprLang == MetricsQLLanguage {
		return validateGroupsWithMetricsQL(promRuleSpec.Groups, validationScheme)
	}

	_, errs := rulefmt.Parse(content, false, validationScheme)
	return errs
}

// validateGroupsWithMetricsQL validates rule groups using the MetricsQL expression parser.
// It replicates the structural checks of rulefmt.Validate but uses metricsql.Parse for expressions.
func validateGroupsWithMetricsQL(groups []monitoringv1.RuleGroup, validationScheme model.ValidationScheme) []error {
	var errs []error
	seen := make(map[string]struct{})
	for _, g := range groups {
		if g.Name == "" {
			errs = append(errs, errors.New("group name must not be empty"))
		}
		if _, ok := seen[g.Name]; ok {
			errs = append(errs, fmt.Errorf("groupname: %q is repeated in the same file", g.Name))
		}
		seen[g.Name] = struct{}{}

		for k, v := range g.Labels {
			if !validationScheme.IsValidLabelName(k) || k == model.MetricNameLabel {
				errs = append(errs, fmt.Errorf("invalid label name: %s", k))
			}
			if !model.LabelValue(v).IsValid() {
				errs = append(errs, fmt.Errorf("invalid label value: %s", v))
			}
		}

		for i, r := range g.Rules {
			errs = append(errs, validateMetricsQLRule(g.Name, i+1, r, validationScheme)...)
		}
	}
	return errs
}

// validateMetricsQLRule validates a single rule using the MetricsQL expression parser.
func validateMetricsQLRule(groupName string, ruleIndex int, r monitoringv1.Rule, validationScheme model.ValidationScheme) []error {
	var errs []error
	ruleName := r.Alert
	if ruleName == "" {
		ruleName = r.Record
	}

	wrap := func(err error) error {
		return fmt.Errorf("group %q, rule %d, %q: %w", groupName, ruleIndex, ruleName, err)
	}

	if r.Record != "" && r.Alert != "" {
		errs = append(errs, wrap(errors.New("only one of 'record' and 'alert' must be set")))
	}
	if r.Record == "" && r.Alert == "" {
		errs = append(errs, wrap(errors.New("one of 'record' or 'alert' must be set")))
	}

	expr := r.Expr.String()
	if expr == "" {
		errs = append(errs, wrap(errors.New("field 'expr' must be set in rule")))
	} else if _, err := metricsql.Parse(expr); err != nil {
		errs = append(errs, wrap(fmt.Errorf("could not parse expression: %w", err)))
	}

	if r.Record != "" {
		if len(r.Annotations) > 0 {
			errs = append(errs, wrap(errors.New("invalid field 'annotations' in recording rule")))
		}
		if r.For != nil && *r.For != "" {
			errs = append(errs, wrap(errors.New("invalid field 'for' in recording rule")))
		}
		if r.KeepFiringFor != nil {
			errs = append(errs, wrap(errors.New("invalid field 'keep_firing_for' in recording rule")))
		}
		if !validationScheme.IsValidMetricName(r.Record) {
			errs = append(errs, wrap(fmt.Errorf("invalid recording rule name: %s", r.Record)))
		}
		if strings.Contains(r.Record, "{") || strings.Contains(r.Record, "}") {
			errs = append(errs, wrap(fmt.Errorf("braces present in the recording rule name; should it be in expr?: %s", r.Record)))
		}
	}

	for k, v := range r.Labels {
		if !validationScheme.IsValidLabelName(k) || k == model.MetricNameLabel {
			errs = append(errs, wrap(fmt.Errorf("invalid label name: %s", k)))
		}
		if !model.LabelValue(v).IsValid() {
			errs = append(errs, wrap(fmt.Errorf("invalid label value: %s", v)))
		}
	}

	for k := range r.Annotations {
		if !validationScheme.IsValidLabelName(k) {
			errs = append(errs, wrap(fmt.Errorf("invalid annotation name: %s", k)))
		}
	}

	return errs
}

// Select selects PrometheusRules by Prometheus or ThanosRuler.
func (prs *PrometheusRuleSelector) Select(namespaces []string) (PrometheusRuleSelection, error) {
	promRules := map[string]*monitoringv1.PrometheusRule{}

	for _, ns := range namespaces {
		err := prs.ruleInformer.ListAllByNamespace(ns, prs.ruleSelector, func(obj any) {
			promRule := obj.(*monitoringv1.PrometheusRule).DeepCopy()
			if err := k8s.AddTypeInformationToObject(promRule); err != nil {
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

	accessor := NewAccessor(prs.logger)

	for ruleName, promRule := range promRules {
		var err error
		var content string
		if err := prs.nsLabeler.EnforceNamespaceLabel(promRule); err != nil {
			continue
		}

		k, ok := accessor.MetaNamespaceKey(promRule)
		if !ok {
			continue
		}

		var reason string
		content, err = prs.generateRulesConfiguration(promRule)
		if err != nil {
			prs.logger.Warn(
				"skipping prometheusrule",
				"error", err.Error(),
				"prometheusrule", promRule.Name,
				"namespace", promRule.Namespace,
			)
			prs.eventRecorder.Eventf(promRule, corev1.EventTypeWarning, InvalidConfigurationEvent, selectingPrometheusRuleResourcesAction, "PrometheusRule %s was rejected due to invalid configuration: %v", promRule.Name, err)
			reason = InvalidConfigurationEvent
		} else {
			marshalRules[ruleName] = content
			namespacedNames = append(namespacedNames, fmt.Sprintf("%s/%s", promRule.Namespace, promRule.Name))
		}

		rules[k] = TypedConfigurationResource[*monitoringv1.PrometheusRule]{
			resource:   promRule,
			err:        err,
			reason:     reason,
			generation: promRule.GetGeneration(),
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
	namePrefix string
	opts       []ObjectOption
	cmClient   typedcorev1.ConfigMapInterface
	cmSelector labels.Set

	logger *slog.Logger
}

func NewPrometheusRuleSyncer(
	logger *slog.Logger,
	namePrefix string,
	cmClient typedcorev1.ConfigMapInterface,
	cmSelector labels.Set,
	options []ObjectOption,
) *PrometheusRuleSyncer {
	return &PrometheusRuleSyncer{
		logger:     logger,
		namePrefix: namePrefix,
		cmClient:   cmClient,
		cmSelector: cmSelector,
		opts:       options,
	}
}

// AppendConfigMapNames adds "virtual" ConfigMap names to the given slice until
// it reaches the limit.
//
// The goal is to avoid statefulset redeployments when the number of "concrete"
// ConfigMaps changes. Because rule's ConfigMaps are mounted with "optional:
// true", the statefulset will be rolled out successfully even if the
// ConfigMaps don't exist and if more ConfigMaps are generated, the operator
// doesn't have to update the statefulset's spec.
func (prs *PrometheusRuleSyncer) AppendConfigMapNames(configMapNames []string, limit int) []string {
	if len(configMapNames) >= limit {
		return configMapNames
	}

	for i := len(configMapNames); i < limit; i++ {
		configMapNames = append(configMapNames, prs.configMapNameAt(i))
	}

	return configMapNames
}

func (prs *PrometheusRuleSyncer) configMapNameAt(i int) string {
	return fmt.Sprintf("%s-rulefiles-%d", prs.namePrefix, i)
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

	// Build a set of desired ConfigMap names for quick lookup.
	desiredNames := make(map[string]struct{}, len(configMaps))
	for _, cm := range configMaps {
		desiredNames[cm.Name] = struct{}{}
	}

	// Create or update the desired ConfigMaps first to ensure rules are never
	// missing. This avoids the race condition where deleting ConfigMaps before
	// creating new ones could cause the config-reloader to reload Prometheus
	// with missing rules.
	prs.logger.Debug("creating/updating ConfigMaps for PrometheusRule")
	for i := range configMaps {
		if err := k8s.CreateOrUpdateConfigMap(ctx, prs.cmClient, &configMaps[i]); err != nil {
			return nil, fmt.Errorf("failed to create or update ConfigMap %q: %w", configMaps[i].Name, err)
		}
	}

	// Delete ConfigMaps that are no longer needed (excess shards from previous
	// reconciliations). This happens after creates/updates to ensure rules
	// remain available throughout the sync process.
	for _, cm := range current {
		if _, exists := desiredNames[cm.Name]; !exists {
			prs.logger.Debug("deleting excess ConfigMap for PrometheusRule", "configmap", cm.Name)
			if err := prs.cmClient.Delete(ctx, cm.Name, metav1.DeleteOptions{}); err != nil {
				return nil, fmt.Errorf("failed to delete excess ConfigMap %q: %w", cm.Name, err)
			}
		}
	}

	return configMapNames(configMaps), nil
}

func configMapNames(cms []corev1.ConfigMap) []string {
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
func (prs *PrometheusRuleSyncer) makeConfigMapsFromRules(rules map[string]string) ([]corev1.ConfigMap, error) {
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

	configMaps := make([]corev1.ConfigMap, 0, len(buckets))
	for i, bucket := range buckets {
		cm := corev1.ConfigMap{
			Data: bucket.rules,
		}

		UpdateObject(
			&cm,
			prs.opts...,
		)

		UpdateObject(
			&cm,
			WithLabels(prs.cmSelector),
			WithName(prs.configMapNameAt(i)),
		)

		configMaps = append(configMaps, cm)
	}

	return configMaps, nil
}
