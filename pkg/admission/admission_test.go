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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/admission/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1beta1"
)

func TestMutateRule(t *testing.T) {
	ts := server(api().servePrometheusRulesMutate)
	defer ts.Close()

	resp := sendAdmissionReview(t, ts, goodRulesWithAnnotations)

	if len(resp.Response.Patch) == 0 {
		t.Errorf("Expected a patch to be applied but found none")
	}
}

func TestMutateRuleNoAnnotations(t *testing.T) {
	ts := server(api().servePrometheusRulesMutate)
	defer ts.Close()

	resp := sendAdmissionReview(t, ts, badRulesNoAnnotations)

	if len(resp.Response.Patch) == 0 {
		t.Errorf("Expected a patch to be applied but found none")
	}
}

func TestAdmitGoodRule(t *testing.T) {
	ts := server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp := sendAdmissionReview(t, ts, goodRulesWithAnnotations)

	if !resp.Response.Allowed {
		t.Errorf("Expected admission to be allowed but it was not")
	}
}

func TestAdmitGoodRuleExternalLabels(t *testing.T) {
	ts := server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp := sendAdmissionReview(t, ts, goodRulesWithExternalLabelsInAnnotations)

	if !resp.Response.Allowed {
		t.Errorf("Expected admission to be allowed but it was not")
	}
}

func TestAdmitBadRule(t *testing.T) {
	ts := server(api().servePrometheusRulesValidate)
	defer ts.Close()

	resp := sendAdmissionReview(t, ts, badRulesNoAnnotations)

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

	resp := sendAdmissionReview(t, ts, badRulesWithBooleanInAnnotations)

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
	resp := sendAdmissionReview(t, ts, request)
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
	ts = server(api().servePrometheusRulesMutate)
	defer ts.Close()

	resp = sendAdmissionReview(t, ts, request)
	if !resp.Response.Allowed {
		t.Errorf("Expected admission to be allowed but it was not")
	}
}

// TestAlertmanagerConfigAdmission tests the admission controller
// validation of the AlertmanagerConfig but does not aim to cover
// all the edge cases of the Validate function in pkg/alertmanager
func TestAlertmanagerConfigAdmission(t *testing.T) {
	ts := server(api().serveAlertmanagerConfigValidate)
	t.Cleanup(ts.Close)

	testCases := []struct {
		name                   string
		apiVersion             string
		spec                   string
		expectAdmissionAllowed bool
	}{
		{
			name:       "Test reject on duplicate receiver",
			apiVersion: "v1alpha1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example"
    },
    {
      "name": "wechat-example"
    }
  ]
}`,
			expectAdmissionAllowed: false,
		},
		{
			name:       "Test reject on duplicate receiver",
			apiVersion: "v1beta1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example"
    },
    {
      "name": "wechat-example"
    }
  ]
}`,
			expectAdmissionAllowed: false,
		},
		{
			name:       "Test reject on invalid receiver",
			apiVersion: "v1alpha1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "https://%<>wechatserver:8080/",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ]
}`,
			expectAdmissionAllowed: false,
		},
		{
			name:       "Test reject on invalid receiver",
			apiVersion: "v1beta1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "https://%<>wechatserver:8080/",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ]
}`,
			expectAdmissionAllowed: false,
		},
		{
			name:       "Test reject on invalid mute time intervals",
			apiVersion: "v1alpha1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "https://wechatserver:8080",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ],
  "muteTimeIntervals": [
    {
      "name": "out-of-business-hours",
      "timeIntervals": [
        {
          "weekdays": [
            "Xaturday",
            "Sunday"
          ]
        },
        {
          "times": [
            {
              "startTime": "50:00",
              "endTime": "08:00"
            },
            {
              "startTime": "18:00",
              "endTime": "24:00"
            }
          ]
        }
      ]
    }
  ]
}`,
			expectAdmissionAllowed: false,
		},
		{
			name:       "Test reject on invalid time intervals",
			apiVersion: "v1beta1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "https://wechatserver:8080",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ],
  "timeIntervals": [
    {
      "name": "out-of-business-hours",
      "timeIntervals": [
        {
          "weekdays": [
            "Xaturday",
            "Sunday"
          ]
        },
        {
          "times": [
            {
              "startTime": "50:00",
              "endTime": "08:00"
            },
            {
              "startTime": "18:00",
              "endTime": "24:00"
            }
          ]
        }
      ]
    }
  ]
}`,
			expectAdmissionAllowed: false,
		},
		{
			name:       "Test happy path",
			apiVersion: "v1alpha1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "http://wechatserver:8080/",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ],
  "muteTimeIntervals": [
    {
      "name": "out-of-business-hours",
      "timeIntervals": [
        {
          "weekdays": [
            "Saturday",
            "Sunday"
          ]
        },
        {
          "times": [
            {
              "startTime": "00:00",
              "endTime": "08:00"
            },
            {
              "startTime": "18:00",
              "endTime": "24:00"
            }
          ]
        }
      ]
    }
  ]
}`,
			expectAdmissionAllowed: true,
		},
		{
			name:       "Test happy path",
			apiVersion: "v1beta1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "http://wechatserver:8080/",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ],
  "timeIntervals": [
    {
      "name": "out-of-business-hours",
      "timeIntervals": [
        {
          "weekdays": [
            "Saturday",
            "Sunday"
          ]
        },
        {
          "times": [
            {
              "startTime": "00:00",
              "endTime": "08:00"
            },
            {
              "startTime": "18:00",
              "endTime": "24:00"
            }
          ]
        }
      ]
    }
  ]
}`,
			expectAdmissionAllowed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name+","+tc.apiVersion, func(t *testing.T) {
			resp := sendAdmissionReview(t, ts, buildAdmissionReviewFromAlertmanagerConfigSpec(t, tc.apiVersion, tc.spec))
			if resp.Response.Allowed != tc.expectAdmissionAllowed {
				t.Errorf(
					"Unexpected admission result, wanted %v but got %v - (warnings=%v) - (details=%v)",
					tc.expectAdmissionAllowed, resp.Response.Allowed, resp.Response.Warnings, resp.Response.Result.Details)
			}
		})
	}
}

func TestAlertmanagerConfigConversion(t *testing.T) {
	ts := server(api().serveConvert)
	t.Cleanup(ts.Close)

	for _, tc := range []struct {
		name string
		from string
		to   string
		spec string

		checkFn func(converted []byte) error
	}{
		{
			name: "happy path",
			from: "v1alpha1",
			to:   "v1beta1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "http://wechatserver:8080/",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ],
  "muteTimeIntervals": [
    {
      "name": "out-of-business-hours",
      "timeIntervals": [
        {
          "weekdays": [
            "Saturday",
            "Sunday"
          ]
        },
        {
          "times": [
            {
              "startTime": "00:00",
              "endTime": "08:00"
            },
            {
              "startTime": "18:00",
              "endTime": "24:00"
            }
          ]
        }
      ]
    }
  ]
}`,

			checkFn: func(converted []byte) error {
				o := v1beta1.AlertmanagerConfig{}

				err := json.Unmarshal(converted, &o)
				if err != nil {
					return err
				}

				if len(o.Spec.TimeIntervals) != 1 {
					return fmt.Errorf("expecting 1 item in spec.timeIntervals, got %d", len(o.Spec.TimeIntervals))
				}

				if o.Spec.TimeIntervals[0].Name != "out-of-business-hours" {
					return fmt.Errorf("expecting spec.timeIntervals[0].name to be %q, got %q", "out-of-business-hours", o.Spec.TimeIntervals[0].Name)
				}

				return nil
			},
		},
		{
			name: "happy path",
			from: "v1beta1",
			to:   "v1alpha1",
			spec: `{
  "route": {
    "groupBy": [
      "job"
    ],
    "groupWait": "30s",
    "groupInterval": "5m",
    "repeatInterval": "12h",
    "receiver": "wechat-example"
  },
  "receivers": [
    {
      "name": "wechat-example",
      "wechatConfigs": [
        {
          "apiURL": "http://wechatserver:8080/",
          "corpID": "wechat-corpid",
          "apiSecret": {
            "name": "wechat-config",
            "key": "apiSecret"
          }
        }
      ]
    }
  ],
  "timeIntervals": [
    {
      "name": "out-of-business-hours",
      "timeIntervals": [
        {
          "weekdays": [
            "Saturday",
            "Sunday"
          ]
        },
        {
          "times": [
            {
              "startTime": "00:00",
              "endTime": "08:00"
            },
            {
              "startTime": "18:00",
              "endTime": "24:00"
            }
          ]
        }
      ]
    }
  ]
}`,

			checkFn: func(converted []byte) error {
				o := v1alpha1.AlertmanagerConfig{}

				err := json.Unmarshal(converted, &o)
				if err != nil {
					return err
				}

				if len(o.Spec.MuteTimeIntervals) != 1 {
					return fmt.Errorf("expecting 1 item in spec.muteTimeIntervals, got %d", len(o.Spec.MuteTimeIntervals))
				}

				if o.Spec.MuteTimeIntervals[0].Name != "out-of-business-hours" {
					return fmt.Errorf("expecting spec.muteTimeIntervals[0].name to be %q, got %q", "out-of-business-hours", o.Spec.MuteTimeIntervals[0].Name)
				}

				return nil
			},
		},
	} {
		t.Run(tc.name+","+tc.from+">"+tc.to, func(t *testing.T) {
			resp := sendConversionReview(t, ts, buildConversionReviewFromAlertmanagerConfigSpec(t, tc.from, tc.to, tc.spec))
			if resp.Response.Result.Status != "Success" {
				t.Fatalf(
					"Unexpected conversion result, wanted 'Success' but got %v - (result=%v)",
					resp.Response.Result.Status,
					resp.Response.Result)
			}

			if len(resp.Response.ConvertedObjects) != 1 {
				t.Fatalf("expected 1 converted object, got %d", len(resp.Response.ConvertedObjects))
			}

			if tc.checkFn == nil {
				return
			}

			err := tc.checkFn(resp.Response.ConvertedObjects[0].Raw)
			if err != nil {
				t.Fatalf("unexpected error while checking converted object: %v", err)
			}
		})
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
	alertManagerConfigValidationTriggered := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_alertmanager_config_validation_triggered_total",
		Help: "Number of times an alertmanagerconfig object triggered validation",
	})

	alertManagerConfigValidationError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_alertmanager_config_validation_errors_total",
		Help: "Number of errors that occurred while validating a alertmanagerconfig object",
	})

	a := New(level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), level.AllowNone()))
	a.RegisterMetrics(validationTriggered, validationErrors, alertManagerConfigValidationTriggered, alertManagerConfigValidationError)

	return a
}

func server(h http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(h)
}

func sendAdmissionReview(t *testing.T, ts *httptest.Server, b []byte) *v1.AdmissionReview {
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST request returned an error: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll(resp.Body) returned an error: %s", err)
	}

	rev := &v1.AdmissionReview{}
	if err := json.Unmarshal(body, rev); err != nil {
		t.Fatalf("unable to parse webhook response: %s", err)
	}

	return rev
}

func sendConversionReview(t *testing.T, ts *httptest.Server, b []byte) *apiextensionsv1.ConversionReview {
	t.Helper()
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST request returned an error: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll(resp.Body) returned an error: %s", err)
	}

	rev := &apiextensionsv1.ConversionReview{}
	if err := json.Unmarshal(body, rev); err != nil {
		t.Fatalf("unable to parse webhook response: %s (%q)", err, string(body))
	}

	return rev
}

var goodRulesWithAnnotations = []byte(`
{
  "kind": "AdmissionReview",
  "apiVersion": "admission.k8s.io/v1",
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
            "partial_response_strategy": "abort",
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
  "apiVersion": "admission.k8s.io/v1",
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
  "apiVersion": "admission.k8s.io/v1",
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
  "apiVersion": "admission.k8s.io/v1",
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
  "apiVersion": "admission.k8s.io/v1",
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

func buildAdmissionReviewFromAlertmanagerConfigSpec(t *testing.T, version, spec string) []byte {
	t.Helper()
	tmpl := fmt.Sprintf(`
{
  "kind": "AdmissionReview",
  "apiVersion": "admission.k8s.io/v1",
  "request": {
    "uid": "87c5df7f-5090-11e9-b9b4-02425473f309",
    "kind": {
      "group": "%s",
      "version": "%s",
      "kind": "%s"
    },
    "resource": {
      "group": "monitoring.coreos.com",
      "version": "%s",
      "resource": "%s"
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
      "apiVersion": "monitoring.coreos.com/%s",
      "kind": "%s",
      "metadata": {
        "creationTimestamp": "2019-03-27T13:02:09Z",
        "generation": 1,
        "name": "test",
        "namespace": "monitoring",
        "uid": "87c5d31d-5090-11e9-b9b4-02425473f309"
      },
    "spec": %s,
    "oldObject": null,
    "dryRun": false
  }
 }
}
`,
		group,
		version,
		alertManagerConfigKind,
		version,
		alertManagerConfigResource,
		version,
		alertManagerConfigKind,
		spec)
	return []byte(tmpl)
}

func buildConversionReviewFromAlertmanagerConfigSpec(t *testing.T, from, to, spec string) []byte {
	t.Helper()
	tmpl := fmt.Sprintf(`
{
  "kind": "ConversionReview",
  "apiVersion": "apiextensions.k8s.io/v1",
  "request": {
    "uid": "87c5df7f-5090-11e9-b9b4-02425473f309",
    "desiredAPIVersion": "monitoring.coreos.com/%s",
    "objects": [{
      "apiVersion": "monitoring.coreos.com/%s",
      "kind": "%s",
      "metadata": {
        "creationTimestamp": "2019-03-27T13:02:09Z",
        "generation": 1,
        "name": "test",
        "namespace": "monitoring",
        "uid": "87c5d31d-5090-11e9-b9b4-02425473f309"
      },
      "spec": %s
    }]
  }
}
`,
		to,
		from,
		alertManagerConfigKind,
		spec)
	return []byte(tmpl)
}
