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

package e2e

import (
	"testing"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/prometheus"
	"github.com/coreos/prometheus-operator/pkg/spec"
)

func TestCreateCluster(t *testing.T) {
	spec := &spec.Prometheus{
		ObjectMeta: v1.ObjectMeta{
			Name: "prometheus-test",
		},
		Spec: spec.PrometheusSpec{
			Replicas: 1,
		},
	}

	_, err := framework.CreatePrometheus(spec)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := framework.DeletePrometheus("prometheus-test"); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := framework.WaitForPodsReady(time.Minute*2, 1, prometheus.ListOptions("prometheus-test")); err != nil {
		t.Fatalf("failed to create 1 Prometheus instances: %v", err)
	}
}
