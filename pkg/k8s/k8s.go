// Copyright The prometheus-operator Authors
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

package k8s

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"net/url"
	"os"
	"strings"

	promversion "github.com/prometheus/common/version"
	appsv1 "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedauthv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringv1beta1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1beta1"
)

const (
	// KubeConfigEnv (optionally) specify the location of kubeconfig file.
	KubeConfigEnv = "KUBECONFIG"

	// PrometheusOperatorFieldManager is the field manager name used by the
	// operator.
	PrometheusOperatorFieldManager = "PrometheusOperator"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(monitoringv1.SchemeBuilder.AddToScheme(scheme))
	utilruntime.Must(monitoringv1alpha1.SchemeBuilder.AddToScheme(scheme))
	utilruntime.Must(monitoringv1beta1.SchemeBuilder.AddToScheme(scheme))
}

type ClusterConfig struct {
	Host           string
	TLSConfig      rest.TLSClientConfig
	AsUser         string
	KubeconfigPath string
}

func NewClusterConfig(config ClusterConfig) (*rest.Config, error) {
	var cfg *rest.Config
	var err error

	if config.KubeconfigPath == "" {
		config.KubeconfigPath = os.Getenv(KubeConfigEnv)
	}

	if config.KubeconfigPath != "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating config from %s: %w", config.KubeconfigPath, err)
		}
	} else {
		if len(config.Host) == 0 {
			if cfg, err = rest.InClusterConfig(); err != nil {
				return nil, err
			}
		} else {
			cfg = &rest.Config{
				Host: config.Host,
			}
			hostURL, err := url.Parse(config.Host)
			if err != nil {
				return nil, fmt.Errorf("error parsing host url %s: %w", config.Host, err)
			}
			if hostURL.Scheme == "https" {
				cfg.TLSClientConfig = config.TLSConfig
			}
		}
	}

	cfg.QPS = 100
	cfg.Burst = 100

	cfg.UserAgent = fmt.Sprintf("PrometheusOperator/%s", promversion.Version)
	cfg.Impersonate.UserName = config.AsUser

	return cfg, nil
}

// ResourceAttribute represents authorization attributes to check on a given resource.
type ResourceAttribute struct {
	Resource string
	Name     string
	Group    string
	Version  string
	Verbs    []string
}

// IsAllowed returns whether the user (e.g. the operator's service account) has
// been granted the required RBAC attributes.
// It returns true when the conditions are met for the namespaces (an empty
// namespace value means "all").
// The second return value returns the list of permissions that are missing if
// the requirements aren't met.
func IsAllowed(
	ctx context.Context,
	ssarClient typedauthv1.SelfSubjectAccessReviewInterface,
	namespaces []string,
	attributes ...ResourceAttribute,
) (bool, []error, error) {
	if len(attributes) == 0 {
		return false, nil, fmt.Errorf("resource attributes must not be empty")
	}

	if len(namespaces) == 0 {
		namespaces = []string{corev1.NamespaceAll}
	}

	var missingPermissions []error
	for _, ns := range namespaces {
		for _, ra := range attributes {
			for _, verb := range ra.Verbs {
				resourceAttributes := authv1.ResourceAttributes{
					Verb:     verb,
					Group:    ra.Group,
					Version:  ra.Version,
					Resource: ra.Resource,
					// An empty name value means "all" resources.
					Name: ra.Name,
					// An empty namespace value means "all" for namespace-scoped resources.
					Namespace: ns,
				}

				// Special case for SAR on namespaces resources: Namespace and
				// Name need to be equal.
				if resourceAttributes.Group == "" && resourceAttributes.Resource == "namespaces" && resourceAttributes.Name != "" && resourceAttributes.Namespace == "" {
					resourceAttributes.Namespace = resourceAttributes.Name
				}

				ssar := &authv1.SelfSubjectAccessReview{
					Spec: authv1.SelfSubjectAccessReviewSpec{
						ResourceAttributes: &resourceAttributes,
					},
				}

				// FIXME(simonpasquier): retry in case of server-side errors.
				ssarResponse, err := ssarClient.Create(ctx, ssar, metav1.CreateOptions{})
				if err != nil {
					return false, nil, err
				}

				if !ssarResponse.Status.Allowed {
					var (
						reason   error
						resource = ra.Resource
					)
					if ra.Name != "" {
						resource += "/" + ra.Name
					}

					switch ns {
					case corev1.NamespaceAll:
						reason = fmt.Errorf("missing %q permission on resource %q (group: %q) for all namespaces", verb, resource, ra.Group)
					default:
						reason = fmt.Errorf("missing %q permission on resource %q (group: %q) for namespace %q", verb, resource, ra.Group, ns)
					}

					missingPermissions = append(missingPermissions, reason)
				}
			}
		}
	}

	return len(missingPermissions) == 0, missingPermissions, nil
}

// UpdateDaemonSet merges metadata of existing DaemonSet with new one and updates it.
func UpdateDaemonSet(ctx context.Context, dmsClient clientappsv1.DaemonSetInterface, dset *appsv1.DaemonSet) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existingDset, err := dmsClient.Get(ctx, dset.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		mergeMetadata(&dset.ObjectMeta, existingDset.ObjectMeta)
		// Propagate annotations set by kubectl on spec.template.annotations. e.g performing a rolling restart.
		copyKubectlAnnotations(&dset.Spec.Template.ObjectMeta, existingDset.Spec.Template.Annotations)

		_, err = dmsClient.Update(ctx, dset, metav1.UpdateOptions{})
		return err
	})
}

// CreateOrUpdateSecret merges metadata of existing Secret with new one and updates it.
func CreateOrUpdateSecret(ctx context.Context, secretClient typedcorev1.SecretInterface, desired *corev1.Secret) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existingSecret, err := secretClient.Get(ctx, desired.Name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}

			_, err = secretClient.Create(ctx, desired, metav1.CreateOptions{})
			return err
		}

		mutated := existingSecret.DeepCopyObject().(*corev1.Secret)
		mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
		if apiequality.Semantic.DeepEqual(existingSecret, desired) {
			return nil
		}
		_, err = secretClient.Update(ctx, desired, metav1.UpdateOptions{})
		return err
	})
}

// CreateOrUpdateConfigMap merges metadata of existing ConfigMap with new one and updates it.
func CreateOrUpdateConfigMap(ctx context.Context, cmClient typedcorev1.ConfigMapInterface, desired *corev1.ConfigMap) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existingCM, err := cmClient.Get(ctx, desired.Name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}

			_, err = cmClient.Create(ctx, desired, metav1.CreateOptions{})
			return err
		}

		mutated := existingCM.DeepCopyObject().(*corev1.ConfigMap)
		mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
		if apiequality.Semantic.DeepEqual(existingCM, desired) {
			return nil
		}
		_, err = cmClient.Update(ctx, desired, metav1.UpdateOptions{})
		return err
	})
}

// IsAPIGroupVersionResourceSupported checks if given groupVersion and resource is supported by the cluster.
func IsAPIGroupVersionResourceSupported(discoveryCli discovery.DiscoveryInterface, groupVersion schema.GroupVersion, resource string) (bool, error) {
	apiResourceList, err := discoveryCli.ServerResourcesForGroupVersion(groupVersion.String())
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	for _, apiResource := range apiResourceList.APIResources {
		if resource == apiResource.Name {
			return true, nil
		}
	}

	return false, nil
}

// AddTypeInformationToObject adds TypeMeta information to a runtime.Object based upon the loaded scheme.Scheme
// See https://github.com/kubernetes/client-go/issues/308#issuecomment-700099260
func AddTypeInformationToObject(obj runtime.Object) error {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}

	for _, gvk := range gvks {
		if len(gvk.Kind) == 0 {
			continue
		}
		if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}

	return nil
}

// mergeMetadata takes labels and annotations from the old resource and merges
// them into the new resource. If a key is present in both resources, the new
// resource wins. All keys starting with the "operator.prometheus.io/" prefix
// in the old resource are dropped before merging.
// It also copies the ResourceVersion from the old resource to the new resource
// to prevent update conflicts.
func mergeMetadata(newObj *metav1.ObjectMeta, oldObj metav1.ObjectMeta) {
	newObj.ResourceVersion = oldObj.ResourceVersion

	newObj.SetLabels(mergeMap(maps.Collect(excludeOperatorPrefixSeq(oldObj.Labels)), maps.All(newObj.Labels)))
	newObj.SetAnnotations(mergeMap(maps.Collect(excludeOperatorPrefixSeq(oldObj.Annotations)), maps.All(newObj.Annotations)))
}

// copyKubectlAnnotations copies the kubectl's annotations into the object
// metadata.
func copyKubectlAnnotations(objMeta *metav1.ObjectMeta, annotations map[string]string) {
	objMeta.SetAnnotations(mergeMap(objMeta.Annotations, filterByPrefixSeq(annotations, "kubectl.kubernetes.io/")))
}

// excludeOperatorPrefixSeq returns a iterator over m excluding all keys
// which start by the reserved operator's prefix.
func excludeOperatorPrefixSeq(m map[string]string) iter.Seq2[string, string] {
	return func(yield func(k, v string) bool) {
		for k, v := range m {
			if strings.HasPrefix(k, "operator.prometheus.io/") {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// filterByPrefixSeq returns a iterator over m for all keys matching the
// prefix.
func filterByPrefixSeq(m map[string]string, prefix string) iter.Seq2[string, string] {
	return func(yield func(k, v string) bool) {
		for k, v := range m {
			if strings.HasPrefix(k, prefix) {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// mergeMap returns a clone of m for which key-value pairs from seq have been added.
func mergeMap(m map[string]string, seq iter.Seq2[string, string]) map[string]string {
	// Don't mutate the input maps.
	m = maps.Clone(m)
	if m == nil {
		m = make(map[string]string)
	}

	maps.Insert(m, seq)
	return m
}
