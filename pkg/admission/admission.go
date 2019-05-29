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
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
)

const (
	addFirstAnnotationPatch string = `[
         { "op": "add", "path": "/metadata/annotations", "value": {"prometheus-operator-validated": "true"}}
     ]`
	addAdditionalAnnotationPatch string = `[
         { "op": "add", "path": "/metadata/annotations/prometheus-operator-validated", "value": "true" }
     ]`
)

var (
	deserializer = scheme.Codecs.UniversalDeserializer()
	ruleResource = metav1.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "prometheusrules",
	}
)

// Admission is a validating and mutating webhook that ensures PrometheusRules pushed into the cluster will be
// valid when loaded by a Prometheus
type Admission struct {
	validationErrorsCounter    *prometheus.Counter
	validationTriggeredCounter *prometheus.Counter
	logger                     log.Logger
}

func New(logger log.Logger) *Admission {
	return &Admission{logger: logger}
}

func (a *Admission) Register(mux *http.ServeMux) {
	mux.HandleFunc("/admission-prometheusrules/validate", a.servePrometheusRulesValidate)
	mux.HandleFunc("/admission-prometheusrules/mutate", a.servePrometheusRulesMutate)
}

func (a *Admission) RegisterMetrics(validationTriggeredCounter, validationErrorsCounter *prometheus.Counter) {
	a.validationTriggeredCounter = validationTriggeredCounter
	a.validationErrorsCounter = validationErrorsCounter
}

type admitFunc func(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

func (a *Admission) servePrometheusRulesMutate(w http.ResponseWriter, r *http.Request) {
	a.serveAdmission(w, r, a.mutatePrometheusRules)
}

func (a *Admission) servePrometheusRulesValidate(w http.ResponseWriter, r *http.Request) {
	a.serveAdmission(w, r, a.validatePrometheusRules)
}

func toAdmissionResponseFailure(message string, errors []error) *v1beta1.AdmissionResponse {
	r := &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Details: &metav1.StatusDetails{
				Causes: []metav1.StatusCause{}}}}

	r.Result.Status = metav1.StatusFailure
	r.Result.Reason = metav1.StatusReasonInvalid
	r.Result.Code = http.StatusUnprocessableEntity
	r.Result.Message = message

	for _, err := range errors {
		r.Result.Details.Name = "prometheusrules"
		r.Result.Details.Causes = append(r.Result.Details.Causes, metav1.StatusCause{Message: err.Error()})
	}

	return r
}

func (a *Admission) serveAdmission(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		level.Warn(a.logger).Log("msg", "request has no body")
		http.Error(w, "request has no body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		level.Warn(a.logger).Log("msg", fmt.Sprintf("invalid Content-Type %s, want `application/json`", contentType))
		http.Error(w, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	level.Debug(a.logger).Log("msg", "Received request", "content", string(body))

	requestedAdmissionReview := v1beta1.AdmissionReview{}
	responseAdmissionReview := v1beta1.AdmissionReview{}

	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		level.Warn(a.logger).Log("msg", "Unable to deserialize request", "err", err)
		responseAdmissionReview.Response = toAdmissionResponseFailure("Unable to deserialize request", []error{err})
	} else {
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	respBytes, err := json.Marshal(responseAdmissionReview)

	level.Debug(a.logger).Log("msg", "sending response", "content", string(respBytes))

	if err != nil {
		level.Error(a.logger).Log("msg", "Cannot serialize response", "err", err)
		http.Error(w, fmt.Sprintf("could not serialize response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(respBytes); err != nil {
		level.Error(a.logger).Log("msg", "Cannot write response", "err", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (a *Admission) mutatePrometheusRules(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	level.Debug(a.logger).Log("msg", "Mutating prometheusrules")

	if ar.Request.Resource != ruleResource {
		err := fmt.Errorf("expected resource to be %v, but received %v", ruleResource, ar.Request.Resource)
		level.Warn(a.logger).Log("err", err)
		return toAdmissionResponseFailure("Unexpected resource kind", []error{err})
	}

	rule := &PrometheusRules{}
	if err := json.Unmarshal(ar.Request.Object.Raw, rule); err != nil {
		level.Info(a.logger).Log("msg", "Cannot unmarshal rules from spec", "err", err)
		return toAdmissionResponseFailure("Cannot unmarshal rules from spec", []error{err})
	}

	reviewResponse := &v1beta1.AdmissionResponse{Allowed: true}
	if len(rule.Annotations) == 0 {
		reviewResponse.Patch = []byte(addFirstAnnotationPatch)
	} else {
		reviewResponse.Patch = []byte(addAdditionalAnnotationPatch)
	}
	pt := v1beta1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	return reviewResponse
}

func (a *Admission) validatePrometheusRules(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	(*a.validationTriggeredCounter).Inc()
	level.Debug(a.logger).Log("msg", "Validating prometheusrules")

	if ar.Request.Resource != ruleResource {
		err := fmt.Errorf("expected resource to be %v, but received %v", ruleResource, ar.Request.Resource)
		level.Warn(a.logger).Log("err", err)
		(*a.validationErrorsCounter).Inc()
		return toAdmissionResponseFailure("Unexpected resource kind", []error{err})
	}

	rules := &PrometheusRules{}
	if err := json.Unmarshal(ar.Request.Object.Raw, rules); err != nil {
		level.Info(a.logger).Log("msg", "Cannot unmarshal rules from spec", "err", err)
		(*a.validationErrorsCounter).Inc()
		return toAdmissionResponseFailure("Cannot unmarshal rules from spec", []error{err})
	}

	_, errors := rulefmt.Parse(rules.Spec.Raw)
	if len(errors) != 0 {
		level.Debug(a.logger).Log("msg", "Invalid rule", "content", rules.Spec.Raw)
		for _, err := range errors {
			level.Info(a.logger).Log("msg", "Invalid rule", "err", err)
		}

		(*a.validationErrorsCounter).Inc()
		return toAdmissionResponseFailure("Rules are not valid", errors)
	}

	return &v1beta1.AdmissionResponse{Allowed: true}
}

func (a *Admission) serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		level.Warn(a.logger).Log("msg", "request has no body")
		http.Error(w, "request has no body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		level.Warn(a.logger).Log("msg", fmt.Sprintf("invalid Content-Type %s, want `application/json`", contentType))
		http.Error(w, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	level.Debug(a.logger).Log("msg", "Received request", "content", string(body))

	requestedAdmissionReview := v1beta1.AdmissionReview{}
	responseAdmissionReview := v1beta1.AdmissionReview{}

	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		level.Warn(a.logger).Log("msg", "Unable to deserialize request", "err", err)
		responseAdmissionReview.Response = toAdmissionResponseFailure("Unable to deserialize request", []error{err})
	} else {
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	respBytes, err := json.Marshal(responseAdmissionReview)

	level.Debug(a.logger).Log("msg", "sending response", "content", string(respBytes))

	if err != nil {
		level.Error(a.logger).Log("msg", "Cannot serialize response", "err", err)
		http.Error(w, fmt.Sprintf("could not serialize response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(respBytes); err != nil {
		level.Error(a.logger).Log("msg", "Cannot write response", "err", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
