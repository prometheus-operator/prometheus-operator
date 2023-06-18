// Copyright 2021 The prometheus-operator Authors
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
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestMakeHostAliases(t *testing.T) {
	cases := []struct {
		input    []monitoringv1.HostAlias
		expected []v1.HostAlias
	}{
		{
			input:    nil,
			expected: nil,
		},
		{
			input:    []monitoringv1.HostAlias{},
			expected: nil,
		},
		{
			input: []monitoringv1.HostAlias{
				{
					IP:        "1.1.1.1",
					Hostnames: []string{"foo.com"},
				},
			},
			expected: []v1.HostAlias{
				{
					IP:        "1.1.1.1",
					Hostnames: []string{"foo.com"},
				},
			},
		},
	}

	for i, c := range cases {
		result := MakeHostAliases(c.input)
		if !reflect.DeepEqual(result, c.expected) {
			t.Errorf("expected test case %d to be %s but got %s", i, c.expected, result)
		}
	}
}
