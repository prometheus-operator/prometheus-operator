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
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMutateRule(t *testing.T) {
	ts := server(api().servePrometheusRulesMutate)
	defer ts.Close()

	resp := send(t, ts, goodRulesWithAnnotations)

	if len(resp.Response.Patch) == 0 {
		t.Errorf("Expected a patch to be applied but found none")
	}
}

func TestMutateRuleNoAnnotations(t *testing.T) {
	ts := server(api().servePrometheusRulesMutate)
	defer ts.Close()

	resp := send(t, ts, badRulesNoAnnotations)

	if len(resp.Response.Patch) == 0 {
		t.Errorf("Expected a patch to be applied but found none")
	}
}

func TestAdmitGoodRule(t *testing.T) {
	ts := server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp := send(t, ts, goodRulesWithAnnotations)

	if !resp.Response.Allowed {
		t.Errorf("Expected admission to be allowed but it was not")
	}
}

func TestAdmitBadRule(t *testing.T) {
	ts := server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp := send(t, ts, badRulesNoAnnotations)

	if resp.Response.Allowed {
		t.Errorf("Expected admission to not be allowed but it was")
	}

	if resp.Response.Result.Details.Causes[0].Message !=
		`group "test.rules", rule 0, "Test": could not parse expression: parse error at char 10: could not parse remaining input ")"...` {
		t.Error("Expected error about inability to parse query")
	}

	if resp.Response.Result.Details.Causes[1].Message !=
		`group "test.rules", rule 0, "Test": msg=template: __alert_Test:1: unrecognized character in action: U+201C '“'` {
		t.Error("Expected error about invalid template")
	}
}

func api() *Admission {
	validationTriggered := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_rule_validation_triggered_total",
		Help: "Number of times a prometheusRule object triggered validation",
	})

	validationErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_rule_validation_errors_total",
		Help: "Number of errors that occurred while validating a prometheusRules object",
	})
	a := &Admission{
		logger:                     log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)),
		validationErrorsCounter:    &validationErrors,
		validationTriggeredCounter: &validationTriggered}
	a.logger = level.NewFilter(a.logger, level.AllowNone())
	return a
}

type serveFunc func(w http.ResponseWriter, r *http.Request)

func server(s serveFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(s))
}

func send(t *testing.T, ts *httptest.Server, rules string) *v1beta1.AdmissionReview {
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(rules))
	if err != nil {
		t.Errorf("Publish() returned an error: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("ioutil.ReadAll(resp.Body) returned an error: %s", err)
	}

	rev := &v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, rev); err != nil {
		t.Errorf("unable to parse webhook response: %s", err)
	}

	return rev
}

var goodRulesWithAnnotations = `
{
  "kind": "AdmissionReview",
  "apiVersion": "admission.k8s.io/v1beta1",
  "request": {
    "uid": "87c5df7f-5090-11e9-b9b4-02425473f309",
    "kind": {
      "group": "monitoring.coreos.com",
      "version": "v1",
      "kind": "PrometheusRule"
    },
    "resource": {
      "group": "monitoring.coreos.com",
      "version": "v1",
      "resource": "prometheusrules"
    },
    "namespace": "monitoring",
    "operation": "CREATE",
    "userInfo": {
      "username": "kubernetes-admin",
      "groups": [
        "system:masters",
        "system:authenticated"
      ]
    },
    "object": {
      "apiVersion": "monitoring.coreos.com/v1",
      "kind": "PrometheusRule",
      "metadata": {
        "annotations": {
          "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"monitoring.coreos.com/v1\",\"kind\":\"PrometheusRule\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"monitoring\"},\"spec\":{\"groups\":[{\"name\":\"test.rules\",\"rules\":[{\"alert\":\"Test\",\"annotations\":{\"message\":\"Test rule\"},\"expr\":\"vector(1))\",\"for\":\"5m\",\"labels\":{\"severity\":\"critical\"}}]}]}}\n"
        },
        "creationTimestamp": "2019-03-27T13:02:09Z",
        "generation": 1,
        "name": "test",
        "namespace": "monitoring",
        "uid": "87c5d31d-5090-11e9-b9b4-02425473f309"
      },
      "spec": {
        "groups": [
          {
            "name": "test.rules",
            "rules": [
              {
                "alert": "Test",
                "annotations": {
                  "message": "Test rule"
                },
                "expr": "vector(1)",
                "for": "5m",
                "labels": {
                  "severity": "critical"
                }
              }
            ]
          }
        ]
      }
    },
    "oldObject": null,
    "dryRun": false
  }
}
`

var badRulesNoAnnotations = `
{
  "kind": "AdmissionReview",
  "apiVersion": "admission.k8s.io/v1beta1",
  "request": {
    "uid": "87c5df7f-5090-11e9-b9b4-02425473f309",
    "kind": {
      "group": "monitoring.coreos.com",
      "version": "v1",
      "kind": "PrometheusRule"
    },
    "resource": {
      "group": "monitoring.coreos.com",
      "version": "v1",
      "resource": "prometheusrules"
    },
    "namespace": "monitoring",
    "operation": "CREATE",
    "userInfo": {
      "username": "kubernetes-admin",
      "groups": [
        "system:masters",
        "system:authenticated"
      ]
    },
    "object": {
      "apiVersion": "monitoring.coreos.com/v1",
      "kind": "PrometheusRule",
      "metadata": {
        "creationTimestamp": "2019-03-27T13:02:09Z",
        "generation": 1,
        "name": "test",
        "namespace": "monitoring",
        "uid": "87c5d31d-5090-11e9-b9b4-02425473f309"
      },
      "spec": {
        "groups": [
          {
            "name": "test.rules",
            "rules": [
              {
                "alert": "Test",
                "annotations": {
                  "message": "Test rule",
                  "val": "{{ print “%f“ $value }}"
                },
                "expr": "vector(1))",
                "for": "5m",
                "labels": {
                  "severity": "critical"
                }
              }
            ]
          }
        ]
      }
    },
    "oldObject": null,
    "dryRun": false
  }
}
`
