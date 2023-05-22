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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"

	"github.com/blang/semver/v4"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	v1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	v1alpha1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1alpha1"
	v1beta1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1beta1"
)

const (
	prometheusOperatorServiceDeploymentName = "prometheus-operator"
	prometheusOperatorCertsSecretName       = "prometheus-operator-certs"

	admissionWebhookServiceName       = "prometheus-operator-admission-webhook"
	standaloneAdmissionHookSecretName = "admission-webhook-certs"

	operatorTLSDir = "/etc/tls/private"
)

type Framework struct {
	KubeClient        kubernetes.Interface
	MonClientV1       v1monitoringclient.MonitoringV1Interface
	MonClientV1alpha1 v1alpha1monitoringclient.MonitoringV1alpha1Interface
	MonClientV1beta1  v1beta1monitoringclient.MonitoringV1beta1Interface
	APIServerClient   apiclient.Interface
	HTTPClient        *http.Client
	MasterHost        string
	DefaultTimeout    time.Duration
	RestConfig        *rest.Config

	operatorVersion semver.Version
	opImage         string
	exampleDir      string
	resourcesDir    string
}

// New setups a test framework and returns it.
func New(kubeconfig, opImage, exampleDir, resourcesDir string, operatorVersion semver.Version) (*Framework, error) {
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

	mClientv1alpha1, err := v1alpha1monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1alpha1 monitoring client failed")
	}

	mClientv1beta1, err := v1beta1monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1beta1 monitoring client failed")
	}

	f := &Framework{
		RestConfig:        config,
		MasterHost:        config.Host,
		KubeClient:        cli,
		MonClientV1:       mClientV1,
		MonClientV1alpha1: mClientv1alpha1,
		MonClientV1beta1:  mClientv1beta1,
		APIServerClient:   apiCli,
		HTTPClient:        httpc,
		DefaultTimeout:    time.Minute,
		operatorVersion:   operatorVersion,
		opImage:           opImage,
		exampleDir:        exampleDir,
		resourcesDir:      resourcesDir,
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

// CreateOrUpdatePrometheusOperator creates or updates a Prometheus Operator Kubernetes Deployment
// inside the specified namespace using the specified operator image. Semver is used
// to control the installation for different version of Prometheus Operator. In addition
// one can specify the namespaces to watch, which defaults to all namespaces.
// Returns the CA, which can bs used to access the operator over TLS
func (f *Framework) CreateOrUpdatePrometheusOperator(ctx context.Context, ns string, namespaceAllowlist,
	namespaceDenylist, prometheusInstanceNamespaces, alertmanagerInstanceNamespaces []string,
	createResourceAdmissionHooks, createClusterRoleBindings, createScrapeConfigCrd bool) ([]FinalizerFn, error) {

	var finalizers []FinalizerFn

	_, err := f.createOrUpdateServiceAccount(
		ctx,
		ns,
		fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-service-account.yaml", f.exampleDir),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create or update prometheus operator service account")
	}

	clusterRole, err := f.CreateOrUpdateClusterRole(ctx, fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml", f.exampleDir))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create or update prometheus cluster role")
	}

	// Add CRD rbac rules
	clusterRole.Rules = append(clusterRole.Rules, CRDCreateRule, CRDMonitoringRule)
	if err := f.UpdateClusterRole(ctx, clusterRole); err != nil {
		return nil, errors.Wrap(err, "failed to update prometheus cluster role")
	}

	if createClusterRoleBindings {
		if _, err := f.createOrUpdateClusterRoleBinding(ctx, ns, fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml", f.exampleDir)); err != nil {
			return nil, errors.Wrap(err, "failed to create or update prometheus cluster role binding")
		}
	} else {
		namespaces := namespaceAllowlist
		namespaces = append(namespaces, prometheusInstanceNamespaces...)
		namespaces = append(namespaces, alertmanagerInstanceNamespaces...)

		for _, n := range namespaces {
			if _, err := f.CreateOrUpdateRoleBindingForSubjectNamespace(ctx, n, ns, fmt.Sprintf("%s/prometheus-operator-role-binding.yaml", f.resourcesDir)); err != nil {
				return nil, errors.Wrap(err, "failed to create or update prometheus operator role binding")
			}
		}
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.AlertmanagerName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Alertmanagers(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize Alertmanager CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.PodMonitorName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.PodMonitors(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize PodMonitor CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.ProbeName, func(opts metav1.ListOptions) (object runtime.Object, err error) {
		return f.MonClientV1.Probes(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize Probe CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.PrometheusName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Prometheuses(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize Prometheus CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.PrometheusRuleName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.PrometheusRules(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize PrometheusRule CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.ServiceMonitorName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.ServiceMonitors(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize ServiceMonitor CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.ThanosRulerName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.ThanosRulers(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize ThanosRuler CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1alpha1.AlertmanagerConfigName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1alpha1.AlertmanagerConfigs(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize AlertmanagerConfig v1alpha1 CRD")
	}

	err = WaitForCRDReady(func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1beta1.AlertmanagerConfigs(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "wait for AlertmanagerConfig v1beta1 CRD")
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1alpha1.PrometheusAgentName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1alpha1.PrometheusAgents(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, errors.Wrap(err, "initialize PrometheusAgent v1alpha1 CRD")
	}

	// TODO(xiu): The OperatorUpgrade tests won't pass because the operator v0.64.1 doesn't have the scrapeconfig CRD.
	// This check can be removed after v0.66.0 is released.
	if createScrapeConfigCrd {
		err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1alpha1.ScrapeConfigName, func(opts metav1.ListOptions) (runtime.Object, error) {
			return f.MonClientV1alpha1.PrometheusAgents(v1.NamespaceAll).List(ctx, opts)
		})
		if err != nil {
			return nil, errors.Wrap(err, "initialize ScrapeConfig v1alpha1 CRD")
		}
	}

	certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(fmt.Sprintf("%s.%s.svc", prometheusOperatorServiceDeploymentName, ns), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate and key")
	}

	if err := f.CreateOrUpdateSecretWithCert(ctx, certBytes, keyBytes, ns, prometheusOperatorCertsSecretName); err != nil {
		return nil, errors.Wrap(err, "failed to create or update prometheus-operator TLS secret")
	}

	deploy, err := MakeDeployment(fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-deployment.yaml", f.exampleDir))
	if err != nil {
		return nil, err
	}

	// Make sure only one version of prometheus operator when update
	deploy.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType

	deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, "--log-level=debug")

	var webhookServerImage string
	if f.opImage != "" {
		// Override operator image used, if specified when running tests.
		deploy.Spec.Template.Spec.Containers[0].Image = f.opImage
		repoAndTag := strings.Split(f.opImage, ":")
		if len(repoAndTag) != 2 {
			return nil, errors.Errorf(
				"expected operator image '%v' split by colon to result in two substrings but got '%v'",
				f.opImage,
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
		webhookServerImage = "quay.io/prometheus-operator/admission-webhook:" + repoAndTag[1]
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
			VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: prometheusOperatorCertsSecretName}}})

	deploy.Spec.Template.Spec.Containers[0].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[0].VolumeMounts,
		v1.VolumeMount{Name: "cert", MountPath: operatorTLSDir, ReadOnly: true})

	// The addition of rule admission webhooks requires TLS, so enable it and
	// switch to a more common https port
	if createResourceAdmissionHooks {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			"--web.enable-tls=true",
			fmt.Sprintf("--web.listen-address=%v", ":8443"),
		)
	}

	err = f.CreateOrUpdateDeploymentAndWaitUntilReady(ctx, ns, deploy)
	if err != nil {
		return nil, err
	}

	service, err := MakeService(fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-service.yaml", f.exampleDir))
	if err != nil {
		return finalizers, errors.Wrap(err, "cannot parse service file")
	}

	service.Namespace = ns
	service.Spec.ClusterIP = ""
	service.Spec.Ports = []v1.ServicePort{{Name: "https", Port: 443, TargetPort: intstr.FromInt(8443)}}

	if _, err := f.CreateOrUpdateServiceAndWaitUntilReady(ctx, ns, service); err != nil {
		return finalizers, errors.Wrap(err, "failed to create or update prometheus operator service")
	}

	if createResourceAdmissionHooks {
		webhookService, b, err := f.CreateOrUpdateAdmissionWebhookServer(ctx, ns, webhookServerImage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create webhook server")
		}

		finalizer, err := f.createOrUpdateMutatingHook(ctx, b, ns, fmt.Sprintf("%s/prometheus-operator-mutatingwebhook.yaml", f.resourcesDir))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create or update mutating webhook for PrometheusRule objects")
		}
		finalizers = append(finalizers, finalizer)

		finalizer, err = f.createOrUpdateValidatingHook(ctx, b, ns, fmt.Sprintf("%s/prometheus-operator-validatingwebhook.yaml", f.resourcesDir))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create or update validating webhook for PrometheusRule objects")
		}
		finalizers = append(finalizers, finalizer)

		finalizer, err = f.createOrUpdateValidatingHook(ctx, b, ns, fmt.Sprintf("%s/alertmanager-config-validating-webhook.yaml", f.resourcesDir))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create or update validating webhook for AlertManagerConfig objects")
		}
		finalizers = append(finalizers, finalizer)

		finalizer, err = f.configureAlertmanagerConfigConversion(ctx, webhookService, b)
		if err != nil {
			return nil, errors.Wrap(err, "failed to configure conversion webhook for AlertManagerConfig objects")
		}
		finalizers = append(finalizers, finalizer)
	}

	return finalizers, nil
}

// DeletePrometheusOperatorClusterResource delete Prometheus Operator cluster wide resources
// if the resource is found.
func (f *Framework) DeletePrometheusOperatorClusterResource(ctx context.Context) error {
	err := f.DeleteClusterRole(ctx, fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml", f.exampleDir))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete prometheus cluster role")
	}

	group := monitoring.GroupName

	alertmanagerCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.AlertmanagerName))
	if err != nil {
		return errors.Wrap(err, "failed to make alertmanager CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", alertmanagerCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete alertmanager CRD")
	}

	podMonitorCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.PodMonitorName))
	if err != nil {
		return errors.Wrap(err, "failed to make podMonitor CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", podMonitorCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete podMonitor CRD")
	}

	probeCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.ProbeName))
	if err != nil {
		return errors.Wrap(err, "failed to make probe CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", probeCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete probe CRD")
	}

	prometheusCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.PrometheusName))
	if err != nil {
		return errors.Wrap(err, "failed to make prometheus CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", prometheusCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete prometheus CRD")
	}

	prometheusRuleCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.PrometheusRuleName))
	if err != nil {
		return errors.Wrap(err, "failed to make prometheusRule CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", prometheusRuleCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete prometheusRule CRD")
	}

	serviceMonitorCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.ServiceMonitorName))
	if err != nil {
		return errors.Wrap(err, "failed to make serviceMonitor CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", serviceMonitorCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete serviceMonitor CRD")
	}

	thanosRulerCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.ThanosRulerName))
	if err != nil {
		return errors.Wrap(err, "failed to make thanosRuler CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", thanosRulerCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete thanosRuler CRD")
	}

	alertmanagerConfigCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1alpha1.AlertmanagerConfigName))
	if err != nil {
		return errors.Wrap(err, "failed to make alertmanagerConfig CRD")
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", alertmanagerConfigCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete alertmanagerConfig CRD")
	}

	operatorMutatingHook, err := parseMutatingHookYaml(fmt.Sprintf("%s/prometheus-operator-mutatingwebhook.yaml", f.resourcesDir))
	if err != nil {
		return errors.Wrap(err, "failed to parse operator mutatingwebhook")
	}
	err = f.deleteMutatingWebhook(ctx, operatorMutatingHook.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete operator mutatingwebhook")
	}

	operatorValidatingHook, err := parseValidatingHookYaml(fmt.Sprintf("%s/prometheus-operator-validatingwebhook.yaml", f.resourcesDir))
	if err != nil {
		return errors.Wrap(err, "failed to parse operator validatingwebhook")
	}
	err = f.deleteValidatingWebhook(ctx, operatorValidatingHook.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete operator mutatingwebhook")
	}

	AlertmanagerConfigValidatingHook, err := parseValidatingHookYaml(fmt.Sprintf("%s/alertmanager-config-validating-webhook.yaml", f.resourcesDir))
	if err != nil {
		return errors.Wrap(err, "failed to parse alertmanager config mutatingwebhook")
	}
	err = f.deleteValidatingWebhook(ctx, AlertmanagerConfigValidatingHook.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete alertmanager config mutatingwebhook")
	}

	return nil
}

func (f *Framework) SetupPrometheusRBAC(ctx context.Context, t *testing.T, testCtx *TestCtx, ns string) {
	if _, err := f.CreateOrUpdateClusterRole(ctx, fmt.Sprintf("%s/rbac/prometheus/prometheus-cluster-role.yaml", f.exampleDir)); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create or update prometheus cluster role: %v", err)
	}
	if finalizerFn, err := f.createOrUpdateServiceAccount(ctx, ns, fmt.Sprintf("%s/rbac/prometheus/prometheus-service-account.yaml", f.exampleDir)); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create or update prometheus service account"))
	} else {
		if testCtx != nil {
			testCtx.AddFinalizerFn(finalizerFn)
		}

	}

	if finalizerFn, err := f.CreateOrUpdateRoleBinding(ctx, ns, fmt.Sprintf("%s/prometheus-role-binding.yml", f.resourcesDir)); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create prometheus role binding"))
	} else {
		if testCtx != nil {
			testCtx.AddFinalizerFn(finalizerFn)
		}
	}
}

func (f *Framework) SetupPrometheusRBACGlobal(ctx context.Context, t *testing.T, testCtx *TestCtx, ns string) {
	if _, err := f.CreateOrUpdateClusterRole(ctx, "../../example/rbac/prometheus/prometheus-cluster-role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create or update prometheus cluster role: %v", err)
	}
	if finalizerFn, err := f.createOrUpdateServiceAccount(ctx, ns, "../../example/rbac/prometheus/prometheus-service-account.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create or update prometheus service account"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	if finalizerFn, err := f.createOrUpdateClusterRoleBinding(ctx, ns, "../../example/rbac/prometheus/prometheus-cluster-role-binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatal(errors.Wrap(err, "failed to create or update prometheus cluster role binding"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}
}

func (f *Framework) configureAlertmanagerConfigConversion(ctx context.Context, svc *v1.Service, cert []byte) (FinalizerFn, error) {
	patch, err := f.MakeCRD(fmt.Sprintf("%s/alertmanager-crd-conversion/patch.json", f.exampleDir))
	if err != nil {
		return nil, err
	}

	crd, err := f.GetCRD(ctx, patch.Name)
	if err != nil {
		return nil, err
	}

	originalBytes, err := json.Marshal(crd)
	if err != nil {
		return nil, err
	}

	patch.Spec.Conversion.Webhook.ClientConfig.Service.Name = svc.Name
	patch.Spec.Conversion.Webhook.ClientConfig.Service.Namespace = svc.Namespace
	patch.Spec.Conversion.Webhook.ClientConfig.Service.Port = &svc.Spec.Ports[0].Port
	patch.Spec.Conversion.Webhook.ClientConfig.CABundle = cert

	crd.Spec.Conversion = patch.Spec.Conversion

	patchBytes, err := json.Marshal(crd)
	if err != nil {
		return nil, err
	}

	jsonResult, err := strategicpatch.StrategicMergePatch(originalBytes, patchBytes, apiextensionsv1.CustomResourceDefinition{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate merge patch: %w", err)
	}

	crd, err = f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Patch(
		ctx,
		crd.Name,
		types.StrategicMergePatchType,
		jsonResult,
		metav1.PatchOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to patch CustomResourceDefinition object: %w", err)
	}

	if crd.Spec.Conversion.Strategy != apiextensionsv1.WebhookConverter {
		return nil, fmt.Errorf("expected conversion strategy to be %s, got %s", apiextensionsv1.WebhookConverter, crd.Spec.Conversion.Strategy)
	}

	finalizerFn := func() error {
		crd, err := f.GetCRD(ctx, patch.Name)
		if err != nil {
			return err
		}

		crd.Spec.Conversion = nil
		_, err = f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to reset conversion configuration of AlertmanagerConfig CRD: %w", err)
		}

		return err
	}

	return finalizerFn, nil
}

// CreateOrUpdateAdmissionWebhookServer deploys an HTTPS server which acts as a
// validating and mutating webhook server for PrometheusRule and
// AlertManagerConfig. It is also able to convert AlertmanagerConfig objects
// from v1alpha1 to v1beta1.
// Returns the service and the certificate authority which can be used to trust the TLS certificate of the server.
func (f *Framework) CreateOrUpdateAdmissionWebhookServer(
	ctx context.Context,
	namespace string,
	image string,
) (*v1.Service, []byte, error) {

	certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(
		fmt.Sprintf("%s.%s.svc", admissionWebhookServiceName, namespace),
		nil,
		nil,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate certificate and key")
	}

	if err := f.CreateOrUpdateSecretWithCert(ctx, certBytes, keyBytes, namespace, standaloneAdmissionHookSecretName); err != nil {
		return nil, nil, errors.Wrap(err, "failed to create or update admission webhook secret")
	}

	deploy, err := MakeDeployment(fmt.Sprintf("%s/admission-webhook/deployment.yaml", f.exampleDir))
	if err != nil {
		return nil, nil, err
	}

	// Deploy only 1 replica because the end-to-end environment (single node
	// cluster) can't satisfy the anti-affinity rules.
	deploy.Spec.Replicas = func(i int32) *int32 { return &i }(1)
	deploy.Spec.Template.Spec.Affinity = nil
	deploy.Spec.Strategy = appsv1.DeploymentStrategy{}

	deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, "--log-level=debug")

	if image != "" {
		// Override operator image used, if specified when running tests.
		deploy.Spec.Template.Spec.Containers[0].Image = image
		repoAndTag := strings.Split(image, ":")
		if len(repoAndTag) != 2 {
			return nil, nil, errors.Errorf(
				"expected image '%v' split by colon to result in two substrings but got '%v'",
				image,
				repoAndTag,
			)
		}
	}

	_, err = f.createOrUpdateServiceAccount(ctx, namespace, fmt.Sprintf("%s/admission-webhook/service-account.yaml", f.exampleDir))
	if err != nil {
		return nil, nil, err
	}

	err = f.CreateOrUpdateDeploymentAndWaitUntilReady(ctx, namespace, deploy)
	if err != nil {
		return nil, nil, err
	}

	service, err := MakeService(fmt.Sprintf("%s/admission-webhook/service.yaml", f.exampleDir))
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot parse service file")
	}

	service.Namespace = namespace
	if _, err := f.CreateOrUpdateServiceAndWaitUntilReady(ctx, namespace, service); err != nil {
		return nil, nil, errors.Wrap(err, "failed to create or update admission webhook server service")
	}

	return service, certBytes, nil
}
