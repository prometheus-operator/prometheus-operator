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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
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
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
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
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
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
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
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
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, false, func(_ context.Context) (bool, error) {
		ctx := context.Background()
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
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
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
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
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

		paPods, err = framework.KubeClient.CoreV1().Pods(ns).List(
			ctx,
			metav1.ListOptions{
				LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
					operator.ApplicationNameLabelKey:     "prometheus-agent",
					operator.ApplicationInstanceLabelKey: name,
				})).String(),
			},
		)
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

func testPrometheusAgentSSetServiceName(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	name := "test-agent-servicename"

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-service", name),
			Namespace: ns,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Name: "web",
					Port: 9090,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":       "prometheus-agent",
				"app.kubernetes.io/instance":   name,
				"app.kubernetes.io/managed-by": "prometheus-operator",
			},
		},
	}

	_, err := framework.KubeClient.CoreV1().Services(ns).Create(context.Background(), svc, metav1.CreateOptions{})
	require.NoError(t, err)

	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	p := framework.MakeBasicPrometheusAgent(ns, name, name, 1)
	p.Spec.ServiceName = &svc.Name

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	targets, err := framework.GetActiveTargets(context.Background(), ns, svc.Name)
	require.NoError(t, err)
	require.Empty(t, targets)

	// Ensure that the default governing service was not created by the operator.
	svcList, err := framework.KubeClient.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, svcList.Items, 1)
	require.Equal(t, svcList.Items[0].Name, svc.Name)

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

func testPrometheusAgentDaemonSetCELValidations(t *testing.T) {
	t.Run("DaemonSetInvalidReplicas", testDaemonSetInvalidReplicas)
	t.Run("DaemonSetInvalidStorage", testDaemonSetInvalidStorage)
	t.Run("DaemonSetInvalidShards", testDaemonSetInvalidShards)
	t.Run("DaemonSetInvalidPVCRetentionPolicy", testDaemonSetInvalidPVCRetentionPolicy)
	t.Run("DaemonSetInvalidScrapeConfigSelector", testDaemonSetInvalidScrapeConfigSelector)
	t.Run("DaemonSetInvalidProbeSelector", testDaemonSetInvalidProbeSelector)
	t.Run("DaemonSetInvalidScrapeConfigNamespaceSelector", testDaemonSetInvalidScrapeConfigNamespaceSelector)
	t.Run("DaemonSetInvalidProbeNamespaceSelector", testDaemonSetInvalidProbeNamespaceSelector)
	t.Run("DaemonSetInvalidServiceMonitorSelector", testDaemonSetInvalidServiceMonitorSelector)
	t.Run("DaemonSetInvalidServiceMonitorNamespaceSelector", testDaemonSetInvalidServiceMonitorNamespaceSelector)
	t.Run("DaemonSetInvalidAdditionalScrapeConfigs", testDaemonSetInvalidAdditionalScrapeConfigs)
}

func testDaemonSetInvalidReplicas(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-replicas"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	// no replicas should be set in Daemonsets
	p.Spec.Replicas = ptr.To(int32(3))

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "replicas cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-storage"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	// storage should not be set in Daemonsets
	p.Spec.CommonPrometheusFields.Storage = &monitoringv1.StorageSpec{
		VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: ptr.To("standard"),
				Resources: v1.VolumeResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("200Mi"),
					},
				},
			},
		},
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "storage cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidShards(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-shards"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	// shards cannot be greater than 1 in DaemonSets
	p.Spec.Shards = ptr.To(int32(2))

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "shards cannot be greater than 1 when mode is DaemonSet")
}

func testDaemonSetInvalidPVCRetentionPolicy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-pvc-retention"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	// persistentVolumeClaimRetentionPolicy cannot be set in DaemonSets
	p.Spec.PersistentVolumeClaimRetentionPolicy = &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
		WhenDeleted: appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
		WhenScaled:  appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "persistentVolumeClaimRetentionPolicy cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidScrapeConfigSelector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-scrape-config-selector"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	// scrapeConfigSelector cannot be set in DaemonSets
	p.Spec.ScrapeConfigSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "prometheus-agent",
		},
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scrapeConfigSelector cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidProbeSelector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-probe-selector"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.ProbeSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "test",
		},
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "probeSelector cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidScrapeConfigNamespaceSelector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-scrape-config-namespace-selector"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.ScrapeConfigNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "test",
		},
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scrapeConfigNamespaceSelector cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidProbeNamespaceSelector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-probe-namespace-selector"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.ProbeNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "test",
		},
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "probeNamespaceSelector cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidServiceMonitorSelector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-service-monitor-selector"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.ServiceMonitorSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "test",
		},
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "serviceMonitorSelector cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidServiceMonitorNamespaceSelector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-service-monitor-namespace-selector"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "test",
		},
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "serviceMonitorNamespaceSelector cannot be set when mode is DaemonSet")
}

func testDaemonSetInvalidAdditionalScrapeConfigs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	_, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx, testFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			AllowedNamespaces:   []string{ns},
			EnabledFeatureGates: []operator.FeatureGateName{operator.PrometheusAgentDaemonSetFeature},
		},
	)
	require.NoError(t, err)

	name := "test-invalid-additional-scrape-configs"
	p := framework.MakeBasicPrometheusAgentDaemonSet(ns, name)

	p.Spec.AdditionalScrapeConfigs = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "test-secret",
		},
		Key: "key",
	}

	_, err = framework.CreatePrometheusAgentAndWaitUntilReady(ctx, ns, p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "additionalScrapeConfigs cannot be set when mode is DaemonSet")
}
