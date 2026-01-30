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
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/cespare/xxhash/v2"
	"github.com/gogo/protobuf/proto"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	v1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	v1alpha1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1alpha1"
	v1beta1monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1beta1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
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
		return nil, fmt.Errorf("build config from flags failed: %w", err)
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating new kube-client failed: %w", err)
	}

	apiCli, err := apiclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating new kube-client failed: %w", err)
	}

	httpc := cli.CoreV1().RESTClient().(*rest.RESTClient).Client

	mClientV1, err := v1monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating v1 monitoring client failed: %w", err)
	}

	mClientv1alpha1, err := v1alpha1monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating v1alpha1 monitoring client failed: %w", err)
	}

	mClientv1beta1, err := v1beta1monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating v1beta1 monitoring client failed: %w", err)
	}

	nodes, err := cli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) < 1 {
		return nil, errors.New("no nodes returned")
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

type PrometheusOperatorOpts struct {
	Namespace              string
	AllowedNamespaces      []string
	DeniedNamespaces       []string
	PrometheusNamespaces   []string
	AlertmanagerNamespaces []string
	EnableAdmissionWebhook bool
	ClusterRoleBindings    bool
	EnableScrapeConfigs    bool
	AdditionalArgs         []string
	EnabledFeatureGates    []operator.FeatureGateName
}

func (f *Framework) CreateOrUpdatePrometheusOperator(
	ctx context.Context,
	namespace string,
	namespaceAllowlist,
	namespaceDenylist,
	prometheusInstanceNamespaces,
	alertmanagerInstanceNamespaces []string,
	createResourceAdmissionHooks,
	createClusterRoleBindings,
	createScrapeConfigCrd bool,
	enabledFeatureGates ...operator.FeatureGateName,
) ([]FinalizerFn, error) {
	return f.CreateOrUpdatePrometheusOperatorWithOpts(
		ctx,
		PrometheusOperatorOpts{
			Namespace:              namespace,
			AllowedNamespaces:      namespaceAllowlist,
			DeniedNamespaces:       namespaceDenylist,
			PrometheusNamespaces:   prometheusInstanceNamespaces,
			AlertmanagerNamespaces: alertmanagerInstanceNamespaces,
			EnableAdmissionWebhook: createResourceAdmissionHooks,
			ClusterRoleBindings:    createClusterRoleBindings,
			EnableScrapeConfigs:    createScrapeConfigCrd,
			EnabledFeatureGates:    enabledFeatureGates,
		},
	)
}

// CreateOrUpdatePrometheusOperatorWithOpts creates or updates a Prometheus
// Operator Kubernetes Deployment inside the specified namespace using the
// specified operator image. Semver is used to control the installation for
// different versions of Prometheus Operator. In addition one can specify the
// namespaces to watch, which defaults to all namespaces.  It returns a slice
// of functions to tear down the deployment.
func (f *Framework) CreateOrUpdatePrometheusOperatorWithOpts(
	ctx context.Context,
	opts PrometheusOperatorOpts,
) ([]FinalizerFn, error) {

	var finalizers []FinalizerFn

	_, err := f.createOrUpdateServiceAccount(
		ctx,
		opts.Namespace,
		fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-service-account.yaml", f.exampleDir),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create or update prometheus operator service account: %w", err)
	}

	clusterRole, err := clusterRoleFromYaml(opts.Namespace, f.exampleDir+"/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load prometheus-operator cluster role: %w", err)
	}

	// Use a unique cluster role name to avoid parallel tests doing concurrent
	// updates to the same resource.
	xxh := xxhash.New()
	if _, err := xxh.Write([]byte(opts.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to write hash: %w", err)
	}
	clusterRole.Name = fmt.Sprintf("%s-%x", clusterRole.Name, xxh.Sum64())

	clusterRole.Rules = append(clusterRole.Rules, CRDCreateRule, CRDMonitoringRule)
	if slices.Contains(opts.EnabledFeatureGates, operator.PrometheusAgentDaemonSetFeature) {
		daemonsetRule := rbacv1.PolicyRule{
			APIGroups: []string{"apps"},
			Resources: []string{"daemonsets"},
			Verbs:     []string{"*"},
		}
		clusterRole.Rules = append(clusterRole.Rules, daemonsetRule)
	}

	clusterRole, err = f.CreateOrUpdateClusterRole(ctx, clusterRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update prometheus cluster role: %w", err)
	}
	finalizers = append(finalizers, func() error {
		return f.DeleteClusterRole(ctx, clusterRole.Name)
	})

	if opts.ClusterRoleBindings {
		// Grant permissions on all namespaces.
		fn, err := f.createOrUpdateClusterRoleBinding(ctx, opts.Namespace, clusterRole, f.exampleDir+"/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to create or update prometheus cluster role binding: %w", err)
		}
		finalizers = append(finalizers, fn)
	} else {
		// Grant permissions on specific namespaces.
		var namespaces []string
		namespaces = append(namespaces, opts.AllowedNamespaces...)
		namespaces = append(namespaces, opts.PrometheusNamespaces...)
		namespaces = append(namespaces, opts.AlertmanagerNamespaces...)

		for _, n := range namespaces {
			if _, err := f.createOrUpdateRoleBindingForSubjectNamespace(ctx, n, opts.Namespace, clusterRole, fmt.Sprintf("%s/prometheus-operator-role-binding.yaml", f.resourcesDir)); err != nil {
				return nil, fmt.Errorf("failed to create or update prometheus operator role binding: %w", err)
			}
		}
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.AlertmanagerName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Alertmanagers(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize Alertmanager CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.PodMonitorName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.PodMonitors(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize PodMonitor CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.ProbeName, func(opts metav1.ListOptions) (object runtime.Object, err error) {
		return f.MonClientV1.Probes(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize Probe CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.PrometheusName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Prometheuses(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize Prometheus CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.PrometheusRuleName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.PrometheusRules(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize PrometheusRule CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.ServiceMonitorName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.ServiceMonitors(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize ServiceMonitor CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1.ThanosRulerName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.ThanosRulers(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize ThanosRuler CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1alpha1.AlertmanagerConfigName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1alpha1.AlertmanagerConfigs(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize AlertmanagerConfig v1alpha1 CRD: %w", err)
	}

	err = WaitForCRDReady(func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1beta1.AlertmanagerConfigs(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("wait for AlertmanagerConfig v1beta1 CRD: %w", err)
	}

	err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1alpha1.PrometheusAgentName, func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1alpha1.PrometheusAgents(v1.NamespaceAll).List(ctx, opts)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize PrometheusAgent v1alpha1 CRD: %w", err)
	}

	if opts.EnableScrapeConfigs {
		err = f.CreateOrUpdateCRDAndWaitUntilReady(ctx, monitoringv1alpha1.ScrapeConfigName, func(opts metav1.ListOptions) (runtime.Object, error) {
			return f.MonClientV1alpha1.ScrapeConfigs(v1.NamespaceAll).List(ctx, opts)
		})
		if err != nil {
			return nil, fmt.Errorf("initialize ScrapeConfig v1alpha1 CRD: %w", err)
		}
	}

	certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(fmt.Sprintf("%s.%s.svc", prometheusOperatorServiceDeploymentName, opts.Namespace), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate and key: %w", err)
	}

	if err := f.CreateOrUpdateSecretWithCert(ctx, certBytes, keyBytes, opts.Namespace, prometheusOperatorCertsSecretName); err != nil {
		return nil, fmt.Errorf("failed to create or update prometheus-operator TLS secret: %w", err)
	}

	deploy, err := MakeDeployment(f.exampleDir + "/rbac/prometheus-operator/prometheus-operator-deployment.yaml")
	if err != nil {
		return nil, err
	}

	// Make sure only that only one instance of the Prometheus operator is running during update.
	deploy.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType

	deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, "--log-level=debug")
	var featureGates string
	if len(opts.EnabledFeatureGates) > 0 {
		featureGates = "-feature-gates="
	}
	for _, fGate := range opts.EnabledFeatureGates {
		featureGates += fmt.Sprintf("%s=true,", fGate)
	}
	if featureGates != "" {
		// Remove the trailing comma
		deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, featureGates[:len(featureGates)-1])
	}

	var webhookServerImage string
	if f.opImage != "" {
		// Override operator image used, if specified when running tests.
		deploy.Spec.Template.Spec.Containers[0].Image = f.opImage
		repoAndTag := strings.Split(f.opImage, ":")
		if len(repoAndTag) != 2 {
			return nil, fmt.Errorf(
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

	deploy.Name = prometheusOperatorServiceDeploymentName

	for _, ns := range opts.AllowedNamespaces {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--namespaces=%v", ns),
		)
	}

	for _, ns := range opts.DeniedNamespaces {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--deny-namespaces=%v", ns),
		)
	}

	for _, ns := range opts.PrometheusNamespaces {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			fmt.Sprintf("--prometheus-instance-namespaces=%v", ns),
		)
	}

	for _, ns := range opts.AlertmanagerNamespaces {
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
	if opts.EnableAdmissionWebhook {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			"--web.enable-tls=true",
			fmt.Sprintf("--web.listen-address=%v", ":8443"),
		)
	}

	deploy.Spec.Template.Spec.Containers[0].Args = append(
		deploy.Spec.Template.Spec.Containers[0].Args,
		opts.AdditionalArgs...,
	)

	err = f.CreateOrUpdateDeploymentAndWaitUntilReady(ctx, opts.Namespace, deploy)
	if err != nil {
		return nil, err
	}

	service, err := MakeService(fmt.Sprintf("%s/rbac/prometheus-operator/prometheus-operator-service.yaml", f.exampleDir))
	if err != nil {
		return finalizers, fmt.Errorf("cannot parse service file: %w", err)
	}

	service.Namespace = opts.Namespace
	service.Spec.ClusterIP = ""
	service.Spec.Ports = []v1.ServicePort{{Name: "https", Port: 443, TargetPort: intstr.FromInt(8443)}}

	if _, err := f.CreateOrUpdateServiceAndWaitUntilReady(ctx, opts.Namespace, service); err != nil {
		return finalizers, fmt.Errorf("failed to create or update prometheus operator service: %w", err)
	}

	if opts.EnableAdmissionWebhook {
		webhookService, b, err := f.CreateOrUpdateAdmissionWebhookServer(ctx, opts.Namespace, webhookServerImage)
		if err != nil {
			return nil, fmt.Errorf("failed to create webhook server: %w", err)
		}

		finalizer, err := f.createOrUpdateMutatingHook(ctx, b, opts.Namespace, fmt.Sprintf("%s/prometheus-operator-mutatingwebhook.yaml", f.resourcesDir))
		if err != nil {
			return nil, fmt.Errorf("failed to create or update mutating webhook for PrometheusRule objects: %w", err)
		}
		finalizers = append(finalizers, finalizer)

		finalizer, err = f.createOrUpdateValidatingHook(ctx, b, opts.Namespace, fmt.Sprintf("%s/prometheus-operator-validatingwebhook.yaml", f.resourcesDir))
		if err != nil {
			return nil, fmt.Errorf("failed to create or update validating webhook for PrometheusRule objects: %w", err)
		}
		finalizers = append(finalizers, finalizer)

		finalizer, err = f.createOrUpdateValidatingHook(ctx, b, opts.Namespace, fmt.Sprintf("%s/alertmanager-config-validating-webhook.yaml", f.resourcesDir))
		if err != nil {
			return nil, fmt.Errorf("failed to create or update validating webhook for AlertManagerConfig objects: %w", err)
		}
		finalizers = append(finalizers, finalizer)

		finalizer, err = f.configureAlertmanagerConfigConversion(ctx, webhookService, b)
		if err != nil {
			return nil, fmt.Errorf("failed to configure conversion webhook for AlertManagerConfig objects: %w", err)
		}
		finalizers = append(finalizers, finalizer)
	}

	return finalizers, nil
}

// DeletePrometheusOperatorClusterResource delete Prometheus Operator cluster wide resources.
func (f *Framework) DeletePrometheusOperatorClusterResource(ctx context.Context) error {
	group := monitoring.GroupName

	alertmanagerCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.AlertmanagerName))
	if err != nil {
		return fmt.Errorf("failed to make alertmanager CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", alertmanagerCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete alertmanager CRD: %w", err)
	}

	podMonitorCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.PodMonitorName))
	if err != nil {
		return fmt.Errorf("failed to make podMonitor CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", podMonitorCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete podMonitor CRD: %w", err)
	}

	probeCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.ProbeName))
	if err != nil {
		return fmt.Errorf("failed to make probe CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", probeCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete probe CRD: %w", err)
	}

	prometheusCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.PrometheusName))
	if err != nil {
		return fmt.Errorf("failed to make prometheus CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", prometheusCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete prometheus CRD: %w", err)
	}

	prometheusRuleCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.PrometheusRuleName))
	if err != nil {
		return fmt.Errorf("failed to make prometheusRule CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", prometheusRuleCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete prometheusRule CRD: %w", err)
	}

	serviceMonitorCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.ServiceMonitorName))
	if err != nil {
		return fmt.Errorf("failed to make serviceMonitor CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", serviceMonitorCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete serviceMonitor CRD: %w", err)
	}

	thanosRulerCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1.ThanosRulerName))
	if err != nil {
		return fmt.Errorf("failed to make thanosRuler CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", thanosRulerCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete thanosRuler CRD: %w", err)
	}

	alertmanagerConfigCRD, err := f.MakeCRD(fmt.Sprintf("%s/prometheus-operator-crd/%s_%s.yaml", f.exampleDir, group, monitoringv1alpha1.AlertmanagerConfigName))
	if err != nil {
		return fmt.Errorf("failed to make alertmanagerConfig CRD: %w", err)
	}
	err = f.DeleteCRD(ctx, fmt.Sprintf("%s.%s", alertmanagerConfigCRD.Name, group))
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete alertmanagerConfig CRD: %w", err)
	}

	operatorMutatingHook, err := parseMutatingHookYaml(fmt.Sprintf("%s/prometheus-operator-mutatingwebhook.yaml", f.resourcesDir))
	if err != nil {
		return fmt.Errorf("failed to parse operator mutatingwebhook: %w", err)
	}
	err = f.deleteMutatingWebhook(ctx, operatorMutatingHook.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete operator mutatingwebhook: %w", err)
	}

	operatorValidatingHook, err := parseValidatingHookYaml(fmt.Sprintf("%s/prometheus-operator-validatingwebhook.yaml", f.resourcesDir))
	if err != nil {
		return fmt.Errorf("failed to parse operator validatingwebhook: %w", err)
	}
	err = f.deleteValidatingWebhook(ctx, operatorValidatingHook.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete operator mutatingwebhook: %w", err)
	}

	AlertmanagerConfigValidatingHook, err := parseValidatingHookYaml(fmt.Sprintf("%s/alertmanager-config-validating-webhook.yaml", f.resourcesDir))
	if err != nil {
		return fmt.Errorf("failed to parse alertmanager config mutatingwebhook: %w", err)
	}
	err = f.deleteValidatingWebhook(ctx, AlertmanagerConfigValidatingHook.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete alertmanager config mutatingwebhook: %w", err)
	}

	return nil
}

func (f *Framework) SetupPrometheusRBAC(ctx context.Context, t *testing.T, testCtx *TestCtx, ns string) {
	t.Helper()

	clusterRole, err := clusterRoleFromYaml(ns, f.exampleDir+"/rbac/prometheus/prometheus-cluster-role.yaml")
	if err != nil {
		t.Fatalf("failed to load prometheus cluster role: %v", err)
	}

	cr, err := f.CreateOrUpdateClusterRole(ctx, clusterRole)
	if err != nil {
		t.Fatalf("failed to create or update prometheus cluster role: %v", err)
	}

	finalizerFn, err := f.createOrUpdateServiceAccount(ctx, ns, f.exampleDir+"/rbac/prometheus/prometheus-service-account.yaml")
	if err != nil {
		t.Fatalf("failed to create or update prometheus service account: %v", err)
	}
	testCtx.AddFinalizerFn(finalizerFn)

	finalizerFn, err = f.createOrUpdateRoleBinding(ctx, ns, cr, f.resourcesDir+"/prometheus-role-binding.yml")
	if err != nil {
		t.Fatalf("failed to create prometheus role binding: %v", err)
	}
	testCtx.AddFinalizerFn(finalizerFn)
}

func (f *Framework) SetupPrometheusRBACGlobal(ctx context.Context, t *testing.T, testCtx *TestCtx, ns string) {
	t.Helper()

	clusterRole, err := clusterRoleFromYaml(ns, f.exampleDir+"/rbac/prometheus/prometheus-cluster-role.yaml")
	if err != nil {
		t.Fatalf("failed to load prometheus cluster role: %v", err)
	}

	if _, err := f.CreateOrUpdateClusterRole(ctx, clusterRole); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create or update prometheus cluster role: %v", err)
	}

	finalizerFn, err := f.createOrUpdateServiceAccount(ctx, ns, f.exampleDir+"/rbac/prometheus/prometheus-service-account.yaml")
	if err != nil {
		t.Fatalf("failed to create or update prometheus service account: %v", err)
	}
	testCtx.AddFinalizerFn(finalizerFn)

	finalizerFn, err = f.createOrUpdateClusterRoleBinding(ctx, ns, clusterRole, f.exampleDir+"/rbac/prometheus/prometheus-cluster-role-binding.yaml")
	if err != nil {
		t.Fatalf("failed to create or update prometheus cluster role binding: %v", err)
	}
	testCtx.AddFinalizerFn(finalizerFn)
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
		return nil, nil, fmt.Errorf("failed to generate certificate and key: %w", err)
	}

	if err := f.CreateOrUpdateSecretWithCert(ctx, certBytes, keyBytes, namespace, standaloneAdmissionHookSecretName); err != nil {
		return nil, nil, fmt.Errorf("failed to create or update admission webhook secret: %w", err)
	}

	deploy, err := MakeDeployment(fmt.Sprintf("%s/admission-webhook/deployment.yaml", f.exampleDir))
	if err != nil {
		return nil, nil, err
	}

	// Adjust replica count in case of single-node clusters because the
	// deployment manifest has anti-affinity rules.
	nodes, err := f.Nodes(ctx)
	if err != nil {
		return nil, nil, err
	}
	if len(nodes) == 1 {
		deploy.Spec.Replicas = ptr.To(int32(1))
		deploy.Spec.Template.Spec.Affinity = nil
		deploy.Spec.Strategy = appsv1.DeploymentStrategy{}
	}

	deploy.Spec.Template.Spec.Containers[0].Args = append(deploy.Spec.Template.Spec.Containers[0].Args, "--log-level=debug")

	if image != "" {
		// Override operator image used, if specified when running tests.
		deploy.Spec.Template.Spec.Containers[0].Image = image
		repoAndTag := strings.Split(image, ":")
		if len(repoAndTag) != 2 {
			return nil, nil, fmt.Errorf(
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
		return nil, nil, fmt.Errorf("cannot parse service file: %w", err)
	}

	service.Namespace = namespace
	if _, err := f.CreateOrUpdateServiceAndWaitUntilReady(ctx, namespace, service); err != nil {
		return nil, nil, fmt.Errorf("failed to create or update admission webhook server service: %w", err)
	}

	return service, certBytes, nil
}

func removeLabelsPatch(labels ...string) ([]byte, error) {
	type patch struct {
		Op   string `json:"op"`
		Path string `json:"path"`
	}

	var patches []patch
	encoder := strings.NewReplacer("/", "~1", "~", "~0")
	for _, label := range labels {
		patches = append(patches, patch{Op: "remove", Path: "/metadata/labels/" + encoder.Replace(label)})
	}

	return json.Marshal(patches)
}
