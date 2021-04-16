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
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	v1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	v1alpha1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1alpha1"
)

const (
	admissionHookSecretName                 = "prometheus-operator-admission"
	prometheusOperatorServiceDeploymentName = "prometheus-operator"
	operatorTLSDir                          = "/etc/tls/private"
)

type Framework struct {
	KubeClient        kubernetes.Interface
	MonClientV1       v1monitoringclient.MonitoringV1Interface
	MonClientV1alpha1 v1alpha1monitoringclient.MonitoringV1alpha1Interface
	APIServerClient   apiclient.Interface
	HTTPClient        *http.Client
	MasterHost        string
	DefaultTimeout    time.Duration
	RestConfig        *rest.Config
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

	apiCli, err := apiclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating new kube-client failed")
	}

	httpc := cli.CoreV1().RESTClient().(*rest.RESTClient).Client
	if err != nil {
		return nil, errors.Wrap(err, "creating http-client failed")
	}

	mClientV1, err := v1monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1 monitoring client failed")
	}

	mClientV1alpha1, err := v1alpha1monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1alpha1 monitoring client failed")
	}

	f := &Framework{
		RestConfig:        config,
		MasterHost:        config.Host,
		KubeClient:        cli,
		MonClientV1:       mClientV1,
		MonClientV1alpha1: mClientV1alpha1,
		APIServerClient:   apiCli,
		HTTPClient:        httpc,
		DefaultTimeout:    time.Minute,
	}

	return f, nil
}

func (f *Framework) MakeEchoService(name, group string, serviceType v1.ServiceType) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("echo-%s", name),
			Labels: map[string]string{
				"group": group,
			},
		},
		Spec: v1.ServiceSpec{
			Type: serviceType,
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"echo": name,
			},
		},
	}
	return service
}

func (f *Framework) MakeEchoDeployment(group string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "echoserver",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: proto.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"echo": group,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"echo": group,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "echoserver",
							Image: "k8s.gcr.io/echoserver:1.10",
							Ports: []v1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 8443,
								},
							},
						},
					},
				},
			},
		},
	}
}

// CreatePrometheusOperator creates a Prometheus Operator Kubernetes Deployment
// inside the specified namespace using the specified operator image. In addition
// one can specify the namespaces to watch, which defaults to all namespaces.
// Returns the CA, which can bs used to access the operator over TLS
func (f *Framework) CreatePrometheusOperator(ns, opImage string, namespaceAllowlist,
	namespaceDenylist, prometheusInstanceNamespaces, alertmanagerInstanceNamespaces []string,
	createRuleAdmissionHooks, createClusterRoleBindings bool) ([]FinalizerFn, error) {

	var finalizers []FinalizerFn

	_, err := createServiceAccount(
		f.KubeClient,
		ns,
		"../../example/rbac/prometheus-operator/prometheus-operator-service-account.yaml",
	)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "failed to create prometheus operator service account")
	}

	clusterRole, err := CreateClusterRole(f.KubeClient, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml")
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, errors.Wrap(err, "failed to create prometheus cluster role")
	}

	// Add CRD rbac rules
	clusterRole.Rules = append(clusterRole.Rules, CRDCreateRule, CRDMonitoringRule)
	if err := UpdateClusterRole(f.KubeClient, clusterRole); err != nil {
		return nil, errors.Wrap(err, "failed to update prometheus cluster role")
	}

	if createClusterRoleBindings {
		if _, err := createClusterRoleBinding(f.KubeClient, ns, "../../example/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
			return nil, errors.Wrap(err, "failed to create prometheus cluster role binding")
		}
	} else {
		namespaces := namespaceAllowlist
		namespaces = append(namespaces, prometheusInstanceNamespaces...)
		namespaces = append(namespaces, alertmanagerInstanceNamespaces...)

		for _, n := range namespaces {
			if _, err := CreateRoleBindingForSubjectNamespace(f.KubeClient, n, ns, "../framework/resources/prometheus-operator-role-binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
				return nil, errors.Wrap(err, "failed to create prometheus operator role binding")
			}
		}
	}

	certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(fmt.Sprintf("%s.%s.svc", prometheusOperatorServiceDeploymentName, ns), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate and key")
	}

	if err := CreateSecretWithCert(f.KubeClient, certBytes, keyBytes, ns, admissionHookSecretName); err != nil {
		return nil, errors.Wrap(err, "failed to create admission webhook secret")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1.AlertmanagerName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Alertmanagers(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize Alertmanager CRD")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1.PodMonitorName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.PodMonitors(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize PodMonitor CRD")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1.ProbeName, func(opts metav1.ListOptions) (object runtime.Object, err error) {
		return f.MonClientV1.Probes(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize Probe CRD")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1.PrometheusName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Prometheuses(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize Prometheus CRD")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1.PrometheusRuleName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.PrometheusRules(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize PrometheusRule CRD")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1.ServiceMonitorName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.ServiceMonitors(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize ServiceMonitor CRD")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1.ThanosRulerName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.ThanosRulers(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize ThanosRuler CRD")
	}

	err = f.CreateCRDAndWaitUntilReady(monitoringv1alpha1.AlertmanagerConfigName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1alpha1.AlertmanagerConfigs(v1.NamespaceAll).List(context.TODO(), opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize AlertmanagerConfig CRD")
	}

	deploy, err := MakeDeployment("../../example/rbac/prometheus-operator/prometheus-operator-deployment.yaml")
	if err != nil {
		return nil, err
	}

	deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, "--log-level=debug")

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
					"quay.io/prometheus-operator/prometheus-config-reloader:" +
					repoAndTag[1]
			}
		}
	}

	deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, "--log-level=all")
	deploy.Name = prometheusOperatorServiceDeploymentName

	for _, ns := range namespaceAllowlist {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--namespaces=%v", ns),
		)
	}

	for _, ns := range namespaceDenylist {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--deny-namespaces=%v", ns),
		)
	}

	for _, ns := range prometheusInstanceNamespaces {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--prometheus-instance-namespaces=%v", ns),
		)
	}

	for _, ns := range alertmanagerInstanceNamespaces {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--alertmanager-instance-namespaces=%v", ns),
		)
	}

	// Load the certificate and key from the created secret into the operator
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes,
		v1.Volume{
			Name:         "cert",
			VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: admissionHookSecretName}}})

	deploy.Spec.Template.Spec.Containers[0].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[0].VolumeMounts,
		v1.VolumeMount{Name: "cert", MountPath: operatorTLSDir, ReadOnly: true})

	// The addition of rule admission webhooks requires TLS, so enable it and
	// switch to a more common https port
	if createRuleAdmissionHooks {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			"--web.enable-tls=true",
			fmt.Sprintf("--web.listen-address=%v", ":8443"),
		)
	}

	err = CreateDeployment(f.KubeClient, ns, deploy)
	if err != nil {
		return nil, err
	}

	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(deploy.Spec.Template.ObjectMeta.Labels)).String()}
	err = WaitForPodsReady(f.KubeClient, ns, f.DefaultTimeout, 1, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for prometheus operator to become ready")
	}

	service, err := MakeService("../../example/rbac/prometheus-operator/prometheus-operator-service.yaml")
	if err != nil {
		return finalizers, errors.Wrap(err, "cannot parse service file")
	}

	service.Namespace = ns
	service.Spec.ClusterIP = ""
	service.Spec.Ports = []v1.ServicePort{{Name: "https", Port: 443, TargetPort: intstr.FromInt(8443)}}

	if _, err := CreateServiceAndWaitUntilReady(f.KubeClient, ns, service); err != nil {
		return finalizers, errors.Wrap(err, "failed to create prometheus operator service")
	}

	if createRuleAdmissionHooks {
		finalizer, err := createMutatingHook(f.KubeClient, certBytes, ns, "../../test/framework/resources/prometheus-operator-mutatingwebhook.yaml")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create mutating webhook")
		}
		finalizers = append(finalizers, finalizer)

		finalizer, err = createValidatingHook(f.KubeClient, certBytes, ns, "../../test/framework/resources/prometheus-operator-validatingwebhook.yaml")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create validating webhook")
		}
		finalizers = append(finalizers, finalizer)
	}

	return finalizers, nil
}

func (ctx *TestCtx) SetupPrometheusRBAC(t *testing.T, ns string, kubeClient kubernetes.Interface) {
	if _, err := CreateClusterRole(kubeClient, "../../example/rbac/prometheus/prometheus-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create prometheus cluster role: %v", err)
	}
	if finalizerFn, err := createServiceAccount(kubeClient, ns, "../../example/rbac/prometheus/prometheus-service-account.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus service account"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if finalizerFn, err := CreateRoleBinding(kubeClient, ns, "../framework/resources/prometheus-role-binding.yml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus role binding"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}
}

func (ctx *TestCtx) SetupPrometheusRBACGlobal(t *testing.T, ns string, kubeClient kubernetes.Interface) {
	if _, err := CreateClusterRole(kubeClient, "../../example/rbac/prometheus/prometheus-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create prometheus cluster role: %v", err)
	}
	if finalizerFn, err := createServiceAccount(kubeClient, ns, "../../example/rbac/prometheus/prometheus-service-account.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus service account"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if finalizerFn, err := createClusterRoleBinding(kubeClient, ns, "../../example/rbac/prometheus/prometheus-cluster-role-binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatal(errors.Wrap(err, "failed to create prometheus cluster role binding"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}
}
