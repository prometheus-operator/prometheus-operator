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

package framework

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/pkg/errors"
)

type Framework struct {
	KubeClient        kubernetes.Interface
	MonClientV1       monitoringv1.MonitoringV1Interface
	MonClientV1alpha1 monitoringv1alpha1.MonitoringV1alpha1Interface
	HTTPClient        *http.Client
	MasterHost        string
	Namespace         *v1.Namespace
	OperatorPod       *v1.Pod
	DefaultTimeout    time.Duration
}

// New setups a test framework and returns it.
func New(ns, kubeconfig, opImage string) (*Framework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "build config from flags failed")
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating new kube-client failed")
	}

	httpc := cli.CoreV1().RESTClient().(*rest.RESTClient).Client
	if err != nil {
		return nil, errors.Wrap(err, "creating http-client failed")
	}

	mClientV1, err := monitoringv1.NewForConfig(&monitoringv1.DefaultCrdKinds, monitoringv1.Group, config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1 monitoring client failed")
	}

	mClientV1alpha1, err := monitoringv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1alpha1 monitoring client failed")
	}

	namespace, err := CreateNamespace(cli, ns)
	if err != nil {
		return nil, err
	}

	f := &Framework{
		MasterHost:        config.Host,
		KubeClient:        cli,
		MonClientV1:       mClientV1,
		MonClientV1alpha1: mClientV1alpha1,
		HTTPClient:        httpc,
		Namespace:         namespace,
		DefaultTimeout:    time.Minute,
	}

	err = f.Setup(opImage)
	if err != nil {
		return nil, errors.Wrap(err, "setup test environment failed")
	}

	return f, nil
}

func (f *Framework) Setup(opImage string) error {
	if err := f.setupPrometheusOperator(opImage); err != nil {
		return errors.Wrap(err, "setup prometheus operator failed")
	}

	return nil
}

func (f *Framework) setupPrometheusOperator(opImage string) error {
	if _, err := CreateServiceAccount(f.KubeClient, f.Namespace.Name, "../../example/rbac/prometheus-operator/prometheus-operator-service-account.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create prometheus operator service account")
	}

	if err := CreateClusterRole(f.KubeClient, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create prometheus cluster role")
	}

	if _, err := CreateClusterRoleBinding(f.KubeClient, f.Namespace.Name, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create prometheus cluster role binding")
	}

	if err := CreateClusterRole(f.KubeClient, "../../example/rbac/prometheus/prometheus-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create prometheus cluster role")
	}

	deploy, err := MakeDeployment("../../example/rbac/prometheus-operator/prometheus-operator-deployment.yaml")
	if err != nil {
		return err
	}

	if opImage != "" {
		// Override operator image used, if specified when running tests.
		deploy.Spec.Template.Spec.Containers[0].Image = opImage
		repoAndTag := strings.Split(opImage, ":")
		if len(repoAndTag) != 2 {
			return errors.Errorf(
				"expected operator image '%v' split by colon to result in two substrings but got '%v'",
				opImage,
				repoAndTag,
			)
		}
		// Override Prometheus config reloader image
		for i, arg := range deploy.Spec.Template.Spec.Containers[0].Args {
			if strings.Contains(arg, "--prometheus-config-reloader=") {
				deploy.Spec.Template.Spec.Containers[0].Args[i] = "--prometheus-config-reloader=" +
					"quay.io/coreos/prometheus-config-reloader:" +
					repoAndTag[1]
			}
		}
	}

	deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, "--log-level=all")

	err = CreateDeployment(f.KubeClient, f.Namespace.Name, deploy)
	if err != nil {
		return err
	}

	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(deploy.Spec.Template.ObjectMeta.Labels)).String()}
	err = WaitForPodsReady(f.KubeClient, f.Namespace.Name, f.DefaultTimeout, 1, opts)
	if err != nil {
		return errors.Wrap(err, "failed to wait for prometheus operator to become ready")
	}

	pl, err := f.KubeClient.Core().Pods(f.Namespace.Name).List(opts)
	if err != nil {
		return err
	}
	f.OperatorPod = &pl.Items[0]

	err = k8sutil.WaitForCRDReady(f.MonClientV1.Prometheuses(v1.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "Prometheus CRD not ready: %v\n")
	}

	err = k8sutil.WaitForCRDReady(f.MonClientV1.ServiceMonitors(v1.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "ServiceMonitor CRD not ready: %v\n")
	}

	err = k8sutil.WaitForCRDReady(f.MonClientV1.PrometheusRules(v1.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "PrometheusRule CRD not ready: %v\n")
	}

	err = k8sutil.WaitForCRDReady(f.MonClientV1.Alertmanagers(v1.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "Alertmanager CRD not ready: %v\n")
	}

	return nil
}

func (ctx *TestCtx) SetupPrometheusRBAC(t *testing.T, ns string, kubeClient kubernetes.Interface) {
	if finalizerFn, err := CreateServiceAccount(kubeClient, ns, "../../example/rbac/prometheus/prometheus-service-account.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus service account"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if finalizerFn, err := CreateRoleBinding(kubeClient, ns, "../framework/ressources/prometheus-role-binding.yml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus role binding"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}
}

func (ctx *TestCtx) SetupPrometheusRBACGlobal(t *testing.T, ns string, kubeClient kubernetes.Interface) {
	if finalizerFn, err := CreateServiceAccount(kubeClient, ns, "../../example/rbac/prometheus/prometheus-service-account.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus service account"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if finalizerFn, err := CreateClusterRoleBinding(kubeClient, ns, "../../example/rbac/prometheus/prometheus-cluster-role-binding.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus cluster role binding"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}
}

// Teardown tears down a previously initialized test environment.
func (f *Framework) Teardown() error {
	if err := f.KubeClient.Core().Services(f.Namespace.Name).Delete("prometheus-operated", nil); err != nil && !k8sutil.IsResourceNotFoundError(err) {
		return err
	}

	if err := f.KubeClient.Core().Services(f.Namespace.Name).Delete("alertmanager-operated", nil); err != nil && !k8sutil.IsResourceNotFoundError(err) {
		return err
	}

	if err := f.KubeClient.Extensions().Deployments(f.Namespace.Name).Delete("prometheus-operator", nil); err != nil {
		return err
	}

	return DeleteNamespace(f.KubeClient, f.Namespace.Name)
}
