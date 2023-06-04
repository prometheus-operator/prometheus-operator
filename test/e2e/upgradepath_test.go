// Copyright 2020 The prometheus-operator Authors
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
	"context"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func testOperatorUpgrade(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	// Delete cluster wide resources to make sure the environment is clean
	err := framework.DeletePrometheusOperatorClusterResource(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Create Prometheus Operator with previous stable minor version
	_, err = previousVersionFramework.CreateOrUpdatePrometheusOperator(context.Background(), ns, nil, nil, nil, nil, true, true, false)
	if err != nil {
		t.Fatal(err)
	}

	name := "operator-upgrade"

	// Create Alertmanager, Prometheus, Thanosruler services with selector labels promised by Prometheus Operator
	alertmanagerService := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("alertmanager-%s", name),
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9093,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":       "alertmanager",
				"app.kubernetes.io/managed-by": "prometheus-operator",
				"app.kubernetes.io/instance":   name,
				"alertmanager":                 name,
			},
		},
	}

	prometheusService := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s", name),
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":       "prometheus",
				"app.kubernetes.io/managed-by": "prometheus-operator",
				"app.kubernetes.io/instance":   name,
				"prometheus":                   name,
				"operator.prometheus.io/shard": "0",
				"operator.prometheus.io/name":  name,
			},
		},
	}

	thanosRulerService := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("thanos-ruler-%s", name),
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":       "thanos-ruler",
				"app.kubernetes.io/managed-by": "prometheus-operator",
				"app.kubernetes.io/instance":   name,
				"thanos-ruler":                 name,
			},
		},
	}

	alertmanager := previousVersionFramework.MakeBasicAlertmanager(ns, name, 1)
	_, err = previousVersionFramework.CreateAlertmanagerAndWaitUntilReady(context.Background(), alertmanager)
	if err != nil {
		t.Fatal(err)
	}
	_, err = previousVersionFramework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, &alertmanagerService)
	if err != nil {
		t.Fatal(err)
	}

	previousVersionFramework.SetupPrometheusRBAC(context.Background(), t, nil, ns)
	prometheus := previousVersionFramework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = previousVersionFramework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, previousVersionFramework.MakeBasicPrometheus(ns, name, name, 1))
	if err != nil {
		t.Fatal(err)
	}

	_, err = previousVersionFramework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, &prometheusService)
	if err != nil {
		t.Fatal(err)
	}

	thanosRuler := previousVersionFramework.MakeBasicThanosRuler(name, 1, "http://test.example.com")
	_, err = previousVersionFramework.CreateThanosRulerAndWaitUntilReady(context.Background(), ns, thanosRuler)
	if err != nil {
		t.Fatal(err)
	}
	_, err = previousVersionFramework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, &thanosRulerService)
	if err != nil {
		t.Fatal(err)
	}

	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	// Update Prometheus Operator to current version
	finalizers, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), ns, nil, nil, nil, nil, true, true, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range finalizers {
		testCtx.AddFinalizerFn(f)
	}

	// Wait for the updated Prometheus Operator to take effect on Alertmanager, Prometheus, and ThanosRuler.
	time.Sleep(time.Minute)

	err = framework.WaitForAlertmanagerReady(context.Background(), alertmanager)
	if err != nil {
		t.Fatal(err)
	}
	err = framework.WaitForServiceReady(context.Background(), ns, alertmanagerService.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForPrometheusReady(context.Background(), prometheus, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	err = framework.WaitForServiceReady(context.Background(), ns, prometheusService.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForThanosRulerReady(context.Background(), ns, thanosRuler, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	err = framework.WaitForServiceReady(context.Background(), ns, thanosRulerService.Name)
	if err != nil {
		t.Fatal(err)
	}
}
