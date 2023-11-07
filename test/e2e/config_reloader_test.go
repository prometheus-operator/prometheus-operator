// Copyright 2023 The prometheus-operator Authors
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
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

func testConfigReloaderResources(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, ns)

	// Start Prometheus operator with the default resource requirements for the
	// config reloader.
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		context.Background(),
		operatorFramework.PrometheusOperatorOpts{
			Namespace:         ns,
			AllowedNamespaces: []string{ns},
		},
	)
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, "instance", "instance", 1)
	p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	// Update the Prometheus operator deployment with new resource requirements
	// for the config reloader.
	var (
		cpuRequest = resource.MustParse("14m")
		cpuLimit   = resource.MustParse("14m")
		memRequest = resource.MustParse("61Mi")
		memLimit   = resource.MustParse("62Mi")
	)
	_, err = framework.CreateOrUpdatePrometheusOperatorWithOpts(
		context.Background(),
		operatorFramework.PrometheusOperatorOpts{
			Namespace:         ns,
			AllowedNamespaces: []string{ns},
			AdditionalArgs: []string{
				"--config-reloader-cpu-limit=" + cpuLimit.String(),
				"--config-reloader-cpu-request=" + cpuRequest.String(),
				"--config-reloader-memory-limit=" + memLimit.String(),
				"--config-reloader-memory-request=" + memRequest.String(),
			},
		},
	)
	require.NoError(t, err)

	sts, err := framework.KubeClient.AppsV1().StatefulSets(p.Namespace).Get(context.Background(), "prometheus-"+p.Name, metav1.GetOptions{})
	require.NoError(t, err)

	var resources []corev1.ResourceRequirements
	var containers []corev1.Container
	containers = append(containers, sts.Spec.Template.Spec.Containers...)
	containers = append(containers, sts.Spec.Template.Spec.InitContainers...)
	for _, c := range containers {
		if c.Name == "config-reloader" || c.Name == "init-config-reloader" {
			resources = append(resources, c.Resources)
		}
	}
	require.Equal(t, 2, len(resources))

	for _, r := range resources {
		require.True(t, cpuLimit.Equal(r.Limits[corev1.ResourceCPU]), "expected %s, got %s", cpuLimit, r.Limits[corev1.ResourceCPU])
		require.True(t, cpuRequest.Equal(r.Requests[corev1.ResourceCPU]), "expected %s, got %s", cpuRequest, r.Requests[corev1.ResourceCPU])
		require.True(t, memLimit.Equal(r.Limits[corev1.ResourceMemory]), "expected %s, got %s", memLimit, r.Limits[corev1.ResourceMemory])
		require.True(t, memRequest.Equal(r.Requests[corev1.ResourceMemory]), "expected %s, got %s", memRequest, r.Requests[corev1.ResourceMemory])
	}
}
