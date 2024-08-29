// Copyright 2019 The prometheus-operator Authors
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

package admission

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/conversion"

	validationv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation/v1alpha1"
	validationv1beta1 "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation/v1beta1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringv1beta1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1beta1"
	promoperator "github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	addFirstAnnotationPatch      = `{ "op": "add", "path": "/metadata/annotations", "value": {"prometheus-operator-validated": "true"}}`
	addAdditionalAnnotationPatch = `{ "op": "add", "path": "/metadata/annotations/prometheus-operator-validated", "value": "true" }`
	errUnmarshalAdmission        = "Cannot unmarshal admission request"
	errUnmarshalRules            = "Cannot unmarshal rules from spec"
	errUnmarshalConfig           = "Cannot unmarhsal config from spec"

	group                  = "monitoring.coreos.com"
	prometheusRuleResource = monitoringv1.PrometheusRuleName
	prometheusRuleVersion  = monitoringv1.Version

	alertManagerConfigResource = monitoringv1beta1.AlertmanagerConfigName
	alertManagerConfigKind     = monitoringv1beta1.AlertmanagerConfigKind

	prometheusRuleValidatePath     = "/admission-prometheusrules/validate"
	prometheusRuleMutatePath       = "/admission-prometheusrules/mutate"
	alertmanagerConfigValidatePath = "/admission-alertmanagerconfigs/validate"
	convertPath                    = "/convert"
)

var (
	deserializer      = kscheme.Codecs.UniversalDeserializer()
	prometheusRuleGVR = metav1.GroupVersionResource{
		Group:    group,
		Version:  prometheusRuleVersion,
		Resource: prometheusRuleResource,
	}
	alertManagerConfigGR = metav1.GroupResource{
		Group:    group,
		Resource: alertManagerConfigResource,
	}
)

// Admission control for:
// 1. PrometheusRules (validation, mutation) - ensuring created resources can be loaded by Promethues
// 2. monitoringv1alpha1.AlertmanagerConfig (validation) - ensuring.
type Admission struct {
	logger *slog.Logger
	wh     http.Handler
}

func New(logger *slog.Logger) *Admission {
	scheme := runtime.NewScheme()
	utilruntime.Must(monitoringv1alpha1.AddToScheme(scheme))
	utilruntime.Must(monitoringv1beta1.AddToScheme(scheme))

	return &Admission{
		logger: logger,
		wh:     conversion.NewWebhookHandler(scheme),
	}
}

func (a *Admission) Register(mux *http.ServeMux) {
	mux.HandleFunc(prometheusRuleValidatePath, a.servePrometheusRulesValidate)
	mux.HandleFunc(prometheusRuleMutatePath, a.servePrometheusRulesMutate)
	mux.HandleFunc(alertmanagerConfigValidatePath, a.serveAlertmanagerConfigValidate)
	mux.HandleFunc(convertPath, a.serveConvert)
}

type admitFunc func(ar v1.AdmissionReview) *v1.AdmissionResponse

func (a *Admission) servePrometheusRulesMutate(w http.ResponseWriter, r *http.Request) {
	a.serveAdmission(w, r, a.mutatePrometheusRules)
}

func (a *Admission) servePrometheusRulesValidate(w http.ResponseWriter, r *http.Request) {
	a.serveAdmission(w, r, a.validatePrometheusRules)
}

func (a *Admission) serveAlertmanagerConfigValidate(w http.ResponseWriter, r *http.Request) {
	a.serveAdmission(w, r, a.validateAlertmanagerConfig)
}

func (a *Admission) serveConvert(w http.ResponseWriter, r *http.Request) {
	a.wh.ServeHTTP(w, r)
}

func toAdmissionResponseFailure(message, resource string, errors []error) *v1.AdmissionResponse {
	r := &v1.AdmissionResponse{
		Result: &metav1.Status{
			Details: &metav1.StatusDetails{
				Causes: []metav1.StatusCause{}}}}

	r.Result.Status = metav1.StatusFailure
	r.Result.Reason = metav1.StatusReasonInvalid
	r.Result.Code = http.StatusUnprocessableEntity
	r.Result.Message = message

	for _, err := range errors {
		r.Result.Details.Name = resource
		r.Result.Details.Causes = append(r.Result.Details.Causes, metav1.StatusCause{Message: err.Error()})
	}

	return r
}

func (a *Admission) serveAdmission(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		a.logger.Warn("request has no body")
		http.Error(w, "request has no body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		a.logger.Warn(fmt.Sprintf("invalid Content-Type %s, want `application/json`", contentType))
		http.Error(w, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	a.logger.Debug("Received request", "content", string(body))

	requestedAdmissionReview := v1.AdmissionReview{}
	responseAdmissionReview := v1.AdmissionReview{}

	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		a.logger.Warn("Unable to deserialize request", "err", err)
		responseAdmissionReview.Response = toAdmissionResponseFailure("Unable to deserialize request", "", []error{err})
	} else {
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
	responseAdmissionReview.APIVersion = requestedAdmissionReview.APIVersion
	responseAdmissionReview.Kind = requestedAdmissionReview.Kind

	respBytes, err := json.Marshal(responseAdmissionReview)

	a.logger.Debug("sending response", "content", string(respBytes))

	if err != nil {
		a.logger.Error("Cannot serialize response", "err", err)
		http.Error(w, fmt.Sprintf("could not serialize response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(respBytes); err != nil {
		a.logger.Error("Cannot write response", "err", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (a *Admission) mutatePrometheusRules(ar v1.AdmissionReview) *v1.AdmissionResponse {
	a.logger.Debug("Mutating prometheusrules")

	if ar.Request.Resource != prometheusRuleGVR {
		err := fmt.Errorf("expected resource to be %v, but received %v", prometheusRuleResource, ar.Request.Resource)
		a.logger.Warn("", "err", err)
		return toAdmissionResponseFailure("Unexpected resource kind", prometheusRuleResource, []error{err})
	}

	rule := &PrometheusRules{}
	if err := json.Unmarshal(ar.Request.Object.Raw, rule); err != nil {
		a.logger.Info(errUnmarshalAdmission, "err", err)
		return toAdmissionResponseFailure(errUnmarshalAdmission, prometheusRuleResource, []error{err})
	}

	patches, err := generatePatchesForNonStringLabelsAnnotations(rule.Spec.Raw)
	if err != nil {
		a.logger.Info(errUnmarshalRules, "err", err)
		return toAdmissionResponseFailure(errUnmarshalRules, prometheusRuleResource, []error{err})
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
	a.logger.Debug("Validating prometheusrules")

	if ar.Request.Resource != prometheusRuleGVR {
		err := fmt.Errorf("expected resource to be %v, but received %v", prometheusRuleResource, ar.Request.Resource)
		a.logger.Warn("", "err", err)
		return toAdmissionResponseFailure("Unexpected resource kind", prometheusRuleResource, []error{err})
	}

	promRule := &monitoringv1.PrometheusRule{}
	if err := json.Unmarshal(ar.Request.Object.Raw, promRule); err != nil {
		a.logger.Info(errUnmarshalRules, "err", err)
		return toAdmissionResponseFailure(errUnmarshalRules, prometheusRuleResource, []error{err})
	}

	errors := promoperator.ValidateRule(promRule.Spec)
	if len(errors) != 0 {
		const m = "Invalid rule"
		a.logger.Debug(m, "content", promRule.Spec)
		for _, err := range errors {
			a.logger.Info(m, "err", err)
		}

		return toAdmissionResponseFailure("Rules are not valid", prometheusRuleResource, errors)
	}

	return &v1.AdmissionResponse{Allowed: true}
}

func (a *Admission) validateAlertmanagerConfig(ar v1.AdmissionReview) *v1.AdmissionResponse {
	a.logger.Debug("Validating alertmanagerconfigs")

	gr := metav1.GroupResource{Group: ar.Request.Resource.Group, Resource: ar.Request.Resource.Resource}
	if gr != alertManagerConfigGR {
		err := fmt.Errorf("expected resource to be %v, but received %v", alertManagerConfigResource, ar.Request.Resource)
		a.logger.Warn("", "err", err)
		return toAdmissionResponseFailure("Unexpected resource kind", alertManagerConfigResource, []error{err})
	}

	var amConf interface{}
	switch ar.Request.Resource.Version {
	case monitoringv1alpha1.Version:
		amConf = &monitoringv1alpha1.AlertmanagerConfig{}
	case monitoringv1beta1.Version:
		amConf = &monitoringv1beta1.AlertmanagerConfig{}
	default:
		err := fmt.Errorf("expected resource version to be 'v1alpha1' or 'v1beta1', but received %v", ar.Request.Resource.Version)
		return toAdmissionResponseFailure("Unexpected resource version", alertManagerConfigResource, []error{err})
	}

	if err := json.Unmarshal(ar.Request.Object.Raw, amConf); err != nil {
		a.logger.Info(errUnmarshalConfig, "err", err)
		return toAdmissionResponseFailure(errUnmarshalConfig, alertManagerConfigResource, []error{err})
	}

	var (
		err error
	)
	switch ar.Request.Resource.Version {
	case monitoringv1alpha1.Version:
		err = validationv1alpha1.ValidateAlertmanagerConfig(amConf.(*monitoringv1alpha1.AlertmanagerConfig))
	case monitoringv1beta1.Version:
		err = validationv1beta1.ValidateAlertmanagerConfig(amConf.(*monitoringv1beta1.AlertmanagerConfig))
	}

	if err != nil {
		msg := "invalid config"
		a.logger.Debug(msg, "content", string(ar.Request.Object.Raw))
		a.logger.Info(msg, "err", err)
		return toAdmissionResponseFailure("AlertmanagerConfig is invalid", alertManagerConfigResource, []error{err})
	}
	return &v1.AdmissionResponse{Allowed: true}
}
