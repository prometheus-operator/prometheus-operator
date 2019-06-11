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
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"

	monitoringclient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"

	"github.com/pkg/errors"
)

const (
	admissionHookSecretName                 = "prometheus-operator-admission"
	prometheusOperatorServiceDeploymentName = "prometheus-operator"
)

type Framework struct {
	KubeClient     kubernetes.Interface
	MonClientV1    monitoringclient.MonitoringV1Interface
	HTTPClient     *http.Client
	MasterHost     string
	DefaultTimeout time.Duration
}

// New setups a test framework and returns it.
func New(kubeconfig, opImage string) (*Framework, error) {
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

	mClientV1, err := monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1 monitoring client failed")
	}

	f := &Framework{
		MasterHost:     config.Host,
		KubeClient:     cli,
		MonClientV1:    mClientV1,
		HTTPClient:     httpc,
		DefaultTimeout: time.Minute,
	}

	return f, nil
}

// CreatePrometheusOperator creates a Prometheus Operator Kubernetes Deployment
// inside the specified namespace using the specified operator image. In addition
// one can specify the namespaces to watch, which defaults to all namespaces.
// Returns the CA, which can bs used to access the operator over TLS
func (f *Framework) CreatePrometheusOperator(ns, opImage string, namespacesToWatch []string, createRuleAdmissionHooks bool) ([]finalizerFn, error) {
	tru := true
	fals := false
	var finalizers []finalizerFn

	_, err := CreateServiceAccount(
		f.KubeClient,
		ns,
		"../../example/rbac/prometheus-operator/prometheus-operator-service-account.yaml",
	)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "failed to create prometheus operator service account")
	}

	if err := CreateClusterRole(f.KubeClient, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "failed to create prometheus cluster role")
	}

	if _, err := CreateClusterRoleBinding(f.KubeClient, ns, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "failed to create prometheus cluster role binding")
	}

	certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(fmt.Sprintf("%s.%s.svc", prometheusOperatorServiceDeploymentName, ns), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate and key")
	}

	if err := CreateSecretWithCert(f.KubeClient, certBytes, keyBytes, ns, admissionHookSecretName); err != nil {
		return nil, errors.Wrap(err, "failed to create admission webhook secret")
	}

	deploy, err := MakeDeployment("../../example/rbac/prometheus-operator/prometheus-operator-deployment.yaml")
	if err != nil {
		return nil, err
	}

	if opImage != "" {
		// Override operator image used, if specified when running tests.
		deploy.Spec.Template.Spec.Containers[0].Image = opImage
		repoAndTag := strings.Split(opImage, ":")
		if len(repoAndTag) != 2 {
			return nil, errors.Errorf(
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
	deploy.Name = prometheusOperatorServiceDeploymentName

	for _, ns := range namespacesToWatch {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--namespaces=%v", ns),
		)
	}

	// Load the certificate and key from the created secret into the operator
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes,
		v1.Volume{
			Name:         "cert",
			VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: admissionHookSecretName}}})

	// Use ghostunnel to provide TLS
	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers,
		v1.Container{
			Name: "ghostunnel",
			Args: []string{
				"server",
				"--listen=:8443",
				"--target=127.0.0.1:8080",
				"--key=cert/key",
				"--cert=cert/cert",
				"--disable-authentication"},
			Image:           "squareup/ghostunnel:v1.4.1",
			Ports:           []v1.ContainerPort{{Name: "https", ContainerPort: 8443}},
			VolumeMounts:    []v1.VolumeMount{{Name: "cert", MountPath: "/cert", ReadOnly: true}},
			SecurityContext: &v1.SecurityContext{AllowPrivilegeEscalation: &fals, ReadOnlyRootFilesystem: &tru}})

	err = CreateDeployment(f.KubeClient, ns, deploy)
	if err != nil {
		return nil, err
	}

	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(deploy.Spec.Template.ObjectMeta.Labels)).String()}
	err = WaitForPodsReady(f.KubeClient, ns, f.DefaultTimeout, 1, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for prometheus operator to become ready")
	}

	err = k8sutil.WaitForCRDReady(func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Prometheuses(v1.NamespaceAll).List(opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "Prometheus CRD not ready: %v\n")
	}

	err = k8sutil.WaitForCRDReady(func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.ServiceMonitors(v1.NamespaceAll).List(opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "ServiceMonitor CRD not ready: %v\n")
	}

	err = k8sutil.WaitForCRDReady(func(opts metav1.ListOptions) (object runtime.Object, e error) {
		return  f.MonClientV1.PodMonitors(v1.NamespaceAll).List(opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "PodMonitor CRD not ready: %v\n")
	}

	err = k8sutil.WaitForCRDReady(func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.PrometheusRules(v1.NamespaceAll).List(opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "PrometheusRule CRD not ready: %v\n")
	}

	err = k8sutil.WaitForCRDReady(func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Alertmanagers(v1.NamespaceAll).List(opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "Alertmanager CRD not ready: %v\n")
	}

	service, err := MakeService("../../example/rbac/prometheus-operator/prometheus-operator-service.yaml")
	if err != nil {
		return finalizers, errors.Wrap(err, "cannot parse service file")
	}

	service.Namespace = ns
	service.Spec.ClusterIP = ""
	service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{Name: "https", Port: 443, TargetPort: intstr.FromInt(8443)})

	if _, err := CreateServiceAndWaitUntilReady(f.KubeClient, ns, service); err != nil {
		return finalizers, errors.Wrap(err, "failed to create prometheus operator service")
	}

	if createRuleAdmissionHooks {
		if finalizer, err := CreateMutatingHook(f.KubeClient, certBytes, ns, "../../test/framework/ressources/prometheus-operator-mutatingwebhook.yaml"); err != nil {
			return nil, errors.Wrap(err, "failed to create mutating webhook")
		} else {
			finalizers = append(finalizers, finalizer)
		}

		if finalizer, err := CreateValidatingHook(f.KubeClient, certBytes, ns, "../../test/framework/ressources/prometheus-operator-validatingwebhook.yaml"); err != nil {
			return nil, errors.Wrap(err, "failed to create validating webhook")
		} else {
			finalizers = append(finalizers, finalizer)
		}
	}

	return finalizers, nil
}

func (ctx *TestCtx) SetupPrometheusRBAC(t *testing.T, ns string, kubeClient kubernetes.Interface) {
	if err := CreateClusterRole(kubeClient, "../../example/rbac/prometheus/prometheus-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create prometheus cluster role: %v", err)
	}
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
	if err := CreateClusterRole(kubeClient, "../../example/rbac/prometheus/prometheus-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create prometheus cluster role: %v", err)
	}
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
