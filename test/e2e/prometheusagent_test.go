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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	pa "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/agent"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

func testCreatePrometheusAgent(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)

	_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD)
	require.NoError(t, err)

	err = framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)

}

func testCreatePrometheusAgentDaemonSet(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ctx := context.Background()

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []string{"PrometheusAgentDaemonSet"},
		},
	)
	require.NoError(t, err)

	name := "test"
	prometheusAgentDSCRD := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p, err := framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, prometheusAgentDSCRD)
	require.NoError(t, err)

	err = framework.DeletePrometheusAgentDSAndWaitUntilGone(ctx, p, ns, name)
	require.NoError(t, err)
}

func testAgentAndServerNameColision(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err := framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, prometheusAgentCRD)
	require.NoError(t, err)
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
	require.NoError(t, err)

	err = framework.DeletePrometheusAgentAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)

}

func testAgentCheckStorageClass(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	name := "test"

	prometheusAgentCRD := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	prometheusAgentCRD, err := framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, prometheusAgentCRD)
	require.NoError(t, err)

	// Invalid storageclass e2e test
	_, err = framework.PatchPrometheusAgent(
		context.Background(),
		prometheusAgentCRD.Name,
		ns,
		monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
						Spec: v1.PersistentVolumeClaimSpec{
							StorageClassName: ptr.To("unknown-storage-class"),
							Resources: v1.VolumeResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceStorage: resource.MustParse("200Mi"),
								},
							},
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	var loopError error
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, framework.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		current, err := framework.MonClientV1alpha1.PrometheusAgents(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			loopError = fmt.Errorf("failed to get object: %w", err)
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Reconciled, monitoringv1.ConditionFalse); err == nil {
			return true, nil
		}

		return false, nil
	})
	require.NoError(t, err, "%v: %v", err, loopError)
}

func testPrometheusAgentStatusScale(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	name := "test"

	pAgent := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	pAgent.Spec.CommonPrometheusFields.Shards = proto.Int32(1)

	pAgent, err := framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, pAgent)
	require.NoError(t, err)

	require.Equal(t, int32(1), pAgent.Status.Shards)

	pAgent, err = framework.ScalePrometheusAgentAndWaitUntilReady(ctx, name, ns, 2)
	require.NoError(t, err)

	require.Equal(t, int32(2), pAgent.Status.Shards)
}

func testPromAgentDaemonSetResourceUpdate(t *testing.T) {
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []string{"PrometheusAgentDaemonSet"},
		},
	)
	require.NoError(t, err)

	name := "test"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}

	p, err = framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	dmsName := fmt.Sprintf("prom-agent-%s", p.Name)
	dms, err := framework.KubeClient.AppsV1().DaemonSets(ns).Get(ctx, dmsName, metav1.GetOptions{})
	require.NoError(t, err)

	res := dms.Spec.Template.Spec.Containers[0].Resources
	require.Equal(t, res, p.Spec.Resources)

	p, err = framework.PatchPrometheusAgentAndWaitUntilReady(
		context.Background(),
		p.Name,
		ns,
		monitoringv1alpha1.PrometheusAgentSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("200Mi"),
					},
				},
			},
		},
	)
	require.NoError(t, err)

	var pollErr error
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		dms, err = framework.KubeClient.AppsV1().DaemonSets(ns).Get(ctx, dmsName, metav1.GetOptions{})
		if err != nil {
			pollErr = fmt.Errorf("failed to get Prometheus Agent DaemonSet: %w", err)
			return false, nil
		}

		res = dms.Spec.Template.Spec.Containers[0].Resources
		if !reflect.DeepEqual(res, p.Spec.Resources) {
			pollErr = fmt.Errorf("resources don't match. Has %#+v, want %#+v", res, p.Spec.Resources)
			return false, nil
		}

		return true, nil
	})

	require.NoError(t, pollErr)
	require.NoError(t, err)
}

func testPromAgentReconcileDaemonSetResourceUpdate(t *testing.T) {
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []string{"PrometheusAgentDaemonSet"},
		},
	)
	require.NoError(t, err)

	name := "test"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}

	p, err = framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	dmsName := fmt.Sprintf("prom-agent-%s", p.Name)
	dms, err := framework.KubeClient.AppsV1().DaemonSets(ns).Get(ctx, dmsName, metav1.GetOptions{})
	require.NoError(t, err)

	res := dms.Spec.Template.Spec.Containers[0].Resources
	require.Equal(t, res, p.Spec.Resources)

	dms.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("200Mi"),
		},
	}
	framework.KubeClient.AppsV1().DaemonSets(ns).Update(ctx, dms, metav1.UpdateOptions{})

	var pollErr error
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		dms, err = framework.KubeClient.AppsV1().DaemonSets(ns).Get(ctx, dmsName, metav1.GetOptions{})
		if err != nil {
			pollErr = fmt.Errorf("failed to get Prometheus Agent DaemonSet: %w", err)
			return false, nil
		}

		res = dms.Spec.Template.Spec.Containers[0].Resources
		if !reflect.DeepEqual(res, p.Spec.Resources) {
			pollErr = fmt.Errorf("resources don't match. Has %#+v, want %#+v", res, p.Spec.Resources)
			return false, nil
		}

		return true, nil
	})

	require.NoError(t, pollErr)
	require.NoError(t, err)
}

func testPromAgentReconcileDaemonSetResourceDelete(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ctx := context.Background()

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []string{"PrometheusAgentDaemonSet"},
		},
	)
	require.NoError(t, err)

	name := "test"
	prometheusAgentDSCRD := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p, err := framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, prometheusAgentDSCRD)
	require.NoError(t, err)

	dmsName := fmt.Sprintf("prom-agent-%s", p.Name)
	framework.KubeClient.AppsV1().DaemonSets(ns).Delete(ctx, dmsName, metav1.DeleteOptions{})

	err = framework.WaitForPrometheusAgentDSReady(ctx, ns, prometheusAgentDSCRD)
	require.NoError(t, err)
}

func testPrometheusAgentDaemonSetSelectPodMonitor(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ctx := context.Background()
	name := "test"

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []string{"PrometheusAgentDaemonSet"},
		},
	)
	require.NoError(t, err)

	app, err := testFramework.MakeDeployment("../../test/framework/resources/basic-app-for-daemonset-test.yaml")
	require.NoError(t, err)

	err = framework.CreateDeployment(ctx, ns, app)
	require.NoError(t, err)

	pm := framework.MakeBasicPodMonitor(name)
	_, err = framework.MonClientV1.PodMonitors(ns).Create(ctx, pm, metav1.CreateOptions{})
	require.NoError(t, err)

	prometheusAgentDS := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)
	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, prometheusAgentDS)
	require.NoError(t, err)

	var pollErr error
	var paPods *v1.PodList
	var firstTargetIP string
	var secondTargetIP string

	appPodsNodes := make([]string, 0, 2)
	appPodsIPs := make([]string, 0, 2)
	paPodsNodes := make([]string, 0, 2)

	cfg := framework.RestConfig
	httpClient := http.Client{}

	err = wait.PollUntilContextTimeout(context.Background(), 15*time.Second, 15*time.Minute, false, func(_ context.Context) (bool, error) {
		ctx := context.Background()

		appPods, err := framework.KubeClient.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
			LabelSelector: "group=test",
		})
		if err != nil {
			pollErr = fmt.Errorf("can't list app pods: %w", err)
			return false, nil
		}

		for _, pod := range appPods.Items {
			appPodsNodes = append(appPodsNodes, pod.Spec.NodeName)
			appPodsIPs = append(appPodsIPs, pod.Status.PodIP)
		}

		paPods, err = framework.KubeClient.CoreV1().Pods(ns).List(ctx, pa.ListOptions(name))
		if err != nil {
			pollErr = fmt.Errorf("can't list prometheus agent pods: %w", err)
			return false, nil
		}

		for _, pod := range paPods.Items {
			paPodsNodes = append(paPodsNodes, pod.Spec.NodeName)
		}

		if len(appPodsNodes) != len(paPodsNodes) {
			pollErr = fmt.Errorf("got %d application pods and %d prometheus-agent pods", len(appPodsNodes), len(paPodsNodes))
			return false, nil
		}
		for _, n := range appPodsNodes {
			if !slices.Contains(paPodsNodes, n) {
				pollErr = fmt.Errorf("no prometheus-agent pod found on node %s", n)
				return false, nil
			}
		}

		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		closer, err := testFramework.StartPortForward(ctx, cfg, "https", paPods.Items[0].Name, ns, "9090")
		if err != nil {
			pollErr = fmt.Errorf("can't start port forward to first prometheus agent pod: %w", err)
			return false, nil
		}
		defer closer()

		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9090/api/v1/targets", nil)
		if err != nil {
			pollErr = fmt.Errorf("can't create http request to first prometheus server: %w", err)
			return false, nil
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			pollErr = fmt.Errorf("can't send http request to first prometheus server: %w", err)
			return false, nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			pollErr = fmt.Errorf("can't read http response from first prometheus server: %w", err)
			return false, nil
		}

		var targetsResponse TargetsResponse
		err = json.Unmarshal(body, &targetsResponse)
		if err != nil {
			pollErr = fmt.Errorf("can't unmarshall target's http response from first prometheus server: %w", err)
			return false, nil
		}
		if len(targetsResponse.Data.ActiveTargets) != 1 {
			pollErr = fmt.Errorf("expect 1 target from first prometheus agent. Actual target's response: %#+v", targetsResponse)
			return false, nil
		}

		target := targetsResponse.Data.ActiveTargets[0]
		instance := target.Labels.Instance
		host := strings.Split(instance, ":")[0]
		ips, err := net.LookupHost(host)
		if err != nil {
			pollErr = fmt.Errorf("can't find IPs from first target's host: %w", err)
			return false, nil
		}

		found := false
		for _, ip := range ips {
			if slices.Contains(appPodsIPs, ip) {
				found = true
				firstTargetIP = ip
			}
		}
		if found == false {
			pollErr = fmt.Errorf("first target IP not found in app's list of pod IPs. Target's IP: %#+v, app's pod IPs: %#+v", ips, appPodsIPs)
			return false, nil
		}

		return true, nil
	})
	require.NoError(t, pollErr)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), 15*time.Second, 15*time.Minute, false, func(_ context.Context) (bool, error) {
		ctx := context.Background()

		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		closer, err := testFramework.StartPortForward(ctx, cfg, "https", paPods.Items[1].Name, ns, "9090")
		if err != nil {
			pollErr = fmt.Errorf("can't start port forward to second prometheus agent pod: %w", err)
			return false, nil
		}
		defer closer()

		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9090/api/v1/targets", nil)
		if err != nil {
			pollErr = fmt.Errorf("can't create http request to second prometheus server: %w", err)
			return false, nil
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			pollErr = fmt.Errorf("can't send http request to second prometheus server: %w", err)
			return false, nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			pollErr = fmt.Errorf("can't read http response from second prometheus server: %w", err)
			return false, nil
		}

		var targetsResponse TargetsResponse
		err = json.Unmarshal(body, &targetsResponse)
		if err != nil {
			pollErr = fmt.Errorf("can't unmarshall target's http response from second prometheus server: %w", err)
			return false, nil
		}
		if len(targetsResponse.Data.ActiveTargets) != 1 {
			pollErr = fmt.Errorf("expect 1 target from second prometheus agent. Actual target's response: %#+v", targetsResponse)
			return false, nil
		}

		target := targetsResponse.Data.ActiveTargets[0]
		instance := target.Labels.Instance
		host := strings.Split(instance, ":")[0]
		ips, err := net.LookupHost(host)
		if err != nil {
			pollErr = fmt.Errorf("can't find IPs from second target's host: %w", err)
			return false, nil
		}

		found := false
		for _, ip := range ips {
			if slices.Contains(appPodsIPs, ip) {
				found = true
				secondTargetIP = ip
			}
		}
		if found == false {
			pollErr = fmt.Errorf("second target IP not found in app's list of pod IPs. Target's IP: %#+v, app's pod IPs: %#+v", ips, appPodsIPs)
			return false, nil
		}

		return true, nil
	})

	require.NoError(t, pollErr)
	require.NoError(t, err)

	require.NotEqual(t, firstTargetIP, secondTargetIP)
}

type Target struct {
	Labels struct {
		Instance string `json:"instance"`
	} `json:"labels"`
}

type TargetsResponse struct {
	Status string `json:"status"`
	Data   struct {
		ActiveTargets []Target `json:"activeTargets"`
	} `json:"data"`
}
