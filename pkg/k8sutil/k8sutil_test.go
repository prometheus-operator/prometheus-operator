// Copyright 2016 The prometheus-operator Authors
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

package k8sutil

import (
	"strings"
	"testing"

	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"k8s.io/apimachinery/pkg/util/validation"
)

func Test_SanitizeVolumeName(t *testing.T) {
	cases := []struct {
		name     string
		expected string
	}{
		{
			name:     "@$!!@$%!#$%!#$%!#$!#$%%$#@!#",
			expected: "",
		},
		{
			name:     "NAME",
			expected: "name",
		},
		{
			name:     "foo--",
			expected: "foo",
		},
		{
			name:     "foo^%#$bar",
			expected: "foo-bar",
		},
		{
			name:     "fOo^%#$bar",
			expected: "foo-bar",
		},
		{
			name:     strings.Repeat("a", validation.DNS1123LabelMaxLength*2),
			expected: strings.Repeat("a", validation.DNS1123LabelMaxLength),
		},
	}

	for i, c := range cases {
		out := SanitizeVolumeName(c.name)
		if c.expected != out {
			t.Errorf("expected test case %d to be %q but got %q", i, c.expected, out)
		}
	}
}

// Test_CRDUnmarshalling checks that known CRD kinds can be unmarshalled successfully
// from the bin data.
func Test_CRDUnmarshalling(t *testing.T) {

	givenGroup := monitoring.GroupName
	givenLabels := map[string]string{
		"foo": "bar",
	}
	givenValidation := true
	givenVersion := "v1"

	cases := []struct {
		givenKind    monitoringv1.CrdKind
		expectedName string
	}{
		{
			givenKind:    monitoringv1.DefaultCrdKinds.Alertmanager,
			expectedName: "alertmanagers." + givenGroup,
		},
		{
			givenKind:    monitoringv1.DefaultCrdKinds.PodMonitor,
			expectedName: "podmonitors." + givenGroup,
		},
		{
			givenKind:    monitoringv1.DefaultCrdKinds.Prometheus,
			expectedName: "prometheuses." + givenGroup,
		},
		{
			givenKind:    monitoringv1.DefaultCrdKinds.PrometheusRule,
			expectedName: "prometheusrules." + givenGroup,
		},
		{
			givenKind:    monitoringv1.DefaultCrdKinds.ServiceMonitor,
			expectedName: "servicemonitors." + givenGroup,
		},
	}

	for _, c := range cases {
		crd := NewCustomResourceDefinition(c.givenKind, givenGroup, givenLabels, givenValidation)
		if c.givenKind.Kind != crd.Spec.Names.Kind {
			t.Errorf("incorrect kind, want: %v, got: %v", c.givenKind.Kind, crd.Spec.Names.Kind)
		}
		if givenVersion != crd.Spec.Version {
			t.Errorf("incorrect version, want: %v, got %v", givenVersion, crd.Spec.Version)
		}
		if c.expectedName != crd.ObjectMeta.Name {
			t.Errorf("want crd name: %v, got: %v", c.expectedName, crd.ObjectMeta.Name)
		}
	}

}
