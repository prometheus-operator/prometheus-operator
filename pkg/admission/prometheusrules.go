package admission

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ruleResource = metav1.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "prometheusrules",
	}
)

func (a *Admission) servePrometheusRulesMutate(w http.ResponseWriter, r *http.Request) {
	a.serveAdmission(w, r, a.mutatePrometheusRules)
}

func (a *Admission) servePrometheusRulesValidate(w http.ResponseWriter, r *http.Request) {
	a.serveAdmission(w, r, a.validatePrometheusRules)
}

func (a *Admission) mutatePrometheusRules(ar v1.AdmissionReview) *v1.AdmissionResponse {
	level.Debug(a.logger).Log("msg", "Mutating prometheusrules")

	if ar.Request.Resource != ruleResource {
		err := fmt.Errorf("expected resource to be %v, but received %v", ruleResource, ar.Request.Resource)
		level.Warn(a.logger).Log("err", err)
		return toAdmissionResponseFailure("Unexpected resource kind", []error{err})
	}

	rule := &PrometheusRules{}
	if err := json.Unmarshal(ar.Request.Object.Raw, rule); err != nil {
		level.Info(a.logger).Log("msg", errUnmarshalAdmission, "err", err)
		return toAdmissionResponseFailure(errUnmarshalAdmission, []error{err})
	}

	patches, err := genRulePatchesForNonStringLabelsAnnotations(rule.Spec.Raw)
	if err != nil {
		level.Info(a.logger).Log("msg", errUnmarshalRules, "err", err)
		return toAdmissionResponseFailure(errUnmarshalRules, []error{err})
	}

	reviewResponse := &v1.AdmissionResponse{Allowed: true}

	if len(rule.Annotations) == 0 {
		patches = append(patches, addFirstAnnotationPatch)
	} else {
		patches = append(patches, addAdditionalAnnotationPatch)
	}
	pt := v1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	reviewResponse.Patch = []byte(fmt.Sprintf("[%s]", strings.Join(patches, ",")))
	return reviewResponse
}

func (a *Admission) validatePrometheusRules(ar v1.AdmissionReview) *v1.AdmissionResponse {
	a.validationTriggeredCounter.Inc()
	level.Debug(a.logger).Log("msg", "Validating prometheusrules")

	if ar.Request.Resource != ruleResource {
		err := fmt.Errorf("expected resource to be %v, but received %v", ruleResource, ar.Request.Resource)
		level.Warn(a.logger).Log("err", err)
		a.validationErrorsCounter.Inc()
		return toAdmissionResponseFailure("Unexpected resource kind", []error{err})
	}

	promRule := &monitoringv1.PrometheusRule{}
	if err := json.Unmarshal(ar.Request.Object.Raw, promRule); err != nil {
		level.Info(a.logger).Log("msg", errUnmarshalRules, "err", err)
		a.validationErrorsCounter.Inc()
		return toAdmissionResponseFailure(errUnmarshalRules, []error{err})
	}

	rules := &PrometheusRules{}
	if err := json.Unmarshal(ar.Request.Object.Raw, rules); err != nil {
		level.Info(a.logger).Log("msg", errUnmarshalAdmission, "err", err)
		a.validationErrorsCounter.Inc()
		return toAdmissionResponseFailure(errUnmarshalAdmission, []error{err})
	}

	_, errors := rulefmt.Parse(rules.Spec.Raw)
	if len(errors) != 0 {
		const m = "Invalid rule"
		level.Debug(a.logger).Log("msg", m, "content", rules.Spec.Raw)
		for _, err := range errors {
			level.Info(a.logger).Log("msg", m, "err", err)
		}

		a.validationErrorsCounter.Inc()
		return toAdmissionResponseFailure("Rules are not valid", errors)
	}

	return &v1.AdmissionResponse{Allowed: true}
}

func genRulePatchesForNonStringLabelsAnnotations(content []byte) ([]string, error) {
	groups := &RuleGroups{}
	if err := json.Unmarshal(content, groups); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal RuleGroups")
	}

	patches := new([]string)
	for gi := range groups.Groups {
		for ri, rule := range groups.Groups[gi].Rules {
			for key, val := range rule.Annotations {
				patchRuleIfNotString(patches, gi, ri, "annotations", key, val)
			}
			for key, val := range rule.Labels {
				patchRuleIfNotString(patches, gi, ri, "labels", key, val)
			}
		}
	}

	return *patches, nil
}

func patchRuleIfNotString(patches *[]string, gi, ri int, typ, key string, val interface{}) {
	if _, ok := val.(string); ok || val == nil {
		// Kubernetes does not let nil values get this far.
		// Keeping it here for the sake of clarity of behavior.
		return
	}
	*patches = append(*patches,
		fmt.Sprintf(`{"op": "replace","path": "/spec/groups/%d/rules/%d/%s/%s","value": "%v"}`,
			gi, ri, typ, key, val))

}
