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
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/pkg/errors"
)

type Framework struct {
	KubeClient     kubernetes.Interface
	MonClient      *v1alpha1.MonitoringV1alpha1Client
	HTTPClient     *http.Client
	MasterHost     string
	Namespace      *v1.Namespace
	OperatorPod    *v1.Pod
	DefaultTimeout time.Duration
}

// Setup setups a test framework and returns it.
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

	mclient, err := v1alpha1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating monitoring client failed")
	}

	namespace, err := CreateNamespace(cli, ns)
	if err != nil {
		return nil, err
	}

	f := &Framework{
		MasterHost:     config.Host,
		KubeClient:     cli,
		MonClient:      mclient,
		HTTPClient:     httpc,
		Namespace:      namespace,
		DefaultTimeout: time.Minute,
	}

	err = f.setup(opImage)
	if err != nil {
		return nil, errors.Wrap(err, "setup test environment failed")
	}

	return f, nil
}

func (f *Framework) setup(opImage string) error {
	if err := f.setupPrometheusOperator(opImage); err != nil {
		return errors.Wrap(err, "setup prometheus operator failed")
	}
	return nil
}

func (f *Framework) setupPrometheusOperator(opImage string) error {
	if _, err := CreateServiceAccount(f.KubeClient, f.Namespace.Name, "../../example/rbac/prometheus-operator/prometheus-operator-service-account.yaml"); err != nil {
		return errors.Wrap(err, "failed to create prometheus operator service account")
	}

	if err := CreateClusterRole(f.KubeClient, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml"); err != nil {
		return errors.Wrap(err, "failed to create prometheus cluster role")
	}

	if _, err := CreateClusterRoleBinding(f.KubeClient, f.Namespace.Name, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml"); err != nil {
		return errors.Wrap(err, "failed to create prometheus cluster role binding")
	}

	if err := CreateClusterRole(f.KubeClient, "../../example/rbac/prometheus/prometheus-cluster-role.yaml"); err != nil {
		return errors.Wrap(err, "failed to create prometheus cluster role")
	}

	deploy, err := MakeDeployment("../../example/rbac/prometheus-operator/prometheus-operator.yaml")
	if err != nil {
		return err
	}

	if opImage != "" {
		// Override operator image used, if specified when running tests.
		deploy.Spec.Template.Spec.Containers[0].Image = opImage
	}

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

	err = k8sutil.WaitForTPRReady(f.KubeClient.Core().RESTClient(), v1alpha1.TPRGroup, v1alpha1.TPRVersion, v1alpha1.TPRPrometheusName)
	if err != nil {
		return err
	}

	err = k8sutil.WaitForTPRReady(f.KubeClient.Core().RESTClient(), v1alpha1.TPRGroup, v1alpha1.TPRVersion, v1alpha1.TPRServiceMonitorName)
	if err != nil {
		return err
	}

	return k8sutil.WaitForTPRReady(f.KubeClient.Core().RESTClient(), v1alpha1.TPRGroup, v1alpha1.TPRVersion, v1alpha1.TPRAlertmanagerName)
}

func (ctx *TestCtx) SetupPrometheusRBAC(t *testing.T, ns string, kubeClient kubernetes.Interface) {
	if finalizerFn, err := CreateServiceAccount(kubeClient, ns, "../../example/rbac/prometheus/prometheus-service-account.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus service account"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if finalizerFn, err := CreateRoleBinding(kubeClient, ns, "framework/ressources/prometheus-role-binding.yml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus role binding"))
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
	if err := DeleteNamespace(f.KubeClient, f.Namespace.Name); err != nil {
		return err
	}

	return nil
}
