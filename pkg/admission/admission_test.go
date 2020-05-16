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
	"bytes"
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/api/admission/v1beta1"
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

func TestAdmitGoodRuleExternalLabels(t *testing.T) {
	ts := server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp := send(t, ts, goodRulesWithExternalLabelsInAnnotations)

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
	{
		exp := 2
		act := len(resp.Response.Result.Details.Causes)
		if act != exp {
			t.Errorf("Expected %d errors but got %d\n", exp, act)
		}
	}
	{
		exp := `unexpected right parenthesis ')'`
		act := resp.Response.Result.Details.Causes[0].Message
		if !strings.Contains(act, exp) {
			t.Error("Expected error about inability to parse query")
		}

		exp = `unrecognized character in action: U+201C`
		act = resp.Response.Result.Details.Causes[1].Message
		if !strings.Contains(act, exp) {
			t.Error("Expected error about invalid character")
		}
	}
}

func TestAdmitBadRuleWithBooleanInAnnotations(t *testing.T) {
	ts := server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp := send(t, ts, badRulesWithBooleanInAnnotations)

	if resp.Response.Allowed {
		t.Errorf("Expected admission to not be allowed but it was")
		return
	}

	if resp.Response.Result.Details.Causes[0].Message !=
		`json: cannot unmarshal bool into Go struct field Rule.spec.groups.rules.annotations of type string` {
		t.Error("Expected error about inability to parse query")
	}
}

func TestMutateNonStringsToStrings(t *testing.T) {
	request := nonStringsInLabelsAnnotations
	ts := server(api().servePrometheusRulesMutate)
	resp := send(t, ts, request)
	if len(resp.Response.Patch) == 0 {
		t.Errorf("Expected a patch to be applied but found none")
	}

	// Apply patch to original request
	patchObj, err := jsonpatch.DecodePatch(resp.Response.Patch)
	if err != nil {
		t.Fatal(err, "Expected a valid patch")
	}
	rev := v1.AdmissionReview{}
	deserializer.Decode(nonStringsInLabelsAnnotations, nil, &rev)
	rev.Request.Object.Raw, err = patchObj.Apply(rev.Request.Object.Raw)
	if err != nil {
		fmt.Println(string(resp.Response.Patch))
		t.Fatal(err, "Expected to successfully apply patch")
	}
	request, _ = json.Marshal(rev)

	// Sent patched request to validation endpoint
	ts.Close()
	ts = server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp = send(t, ts, request)
	if !resp.Response.Allowed {
		t.Errorf("Expected admission to be allowed but it was not")
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

func send(t *testing.T, ts *httptest.Server, rules []byte) *v1beta1.AdmissionReview {
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(rules))
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

var goodRulesWithAnnotations = []byte(`
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
                  "message": "Test rule",
                  "humanizePercentage": "Should work {{ $value | humanizePercentage }}"
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
`)

var goodRulesWithExternalLabelsInAnnotations = []byte(`
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
                  "message": "Test externalLabels {{ $externalLabels.cluster }}"
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
`)

var badRulesNoAnnotations = []byte(`
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
`)

var badRulesWithBooleanInAnnotations = []byte(`
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
                  "badBoolean": false,
                  "message": "Test rule",
                  "humanizePercentage": "Should work {{ $value | humanizePercentage }}"
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
`)

var nonStringsInLabelsAnnotations = []byte(`
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
                  "annBool": false,
                  "message": "Test rule",
                  "annNil": null,
                  "humanizePercentage": "Should work {{ $value | humanizePercentage }}",
                  "annEmpty": "",
                  "annInteger": 1
                },
                "expr": "vector(1)",
                "for": "5m",
                "labels": {
                  "severity": "critical",
                  "labNil": null,
                  "labInt": 1,
                  "labEmpty": "",
                  "labBool": true
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
}`)
