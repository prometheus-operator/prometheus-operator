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

package prometheus

import (
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExpectedStatefulSetShardNames(t *testing.T) {
	replicas := int32(2)
	shards := int32(3)
	res := expectedStatefulSetShardNames(&monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Shards:   &shards,
				Replicas: &replicas,
			},
		},
	})

	expected := []string{
		"prometheus-test",
		"prometheus-test-shard-1",
		"prometheus-test-shard-2",
	}

	for i, name := range expected {
		if res[i] != name {
			t.Fatal("Unexpected StatefulSet shard name")
		}
	}
}
