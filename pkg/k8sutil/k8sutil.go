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

package k8sutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/cespare/xxhash/v2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringv1beta1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1beta1"
	promversion "github.com/prometheus/common/version"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/discovery"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

// KubeConfigEnv (optionally) specify the location of kubeconfig file
const KubeConfigEnv = "KUBECONFIG"

var invalidDNS1123Characters = regexp.MustCompile("[^-a-z0-9]+")

var scheme = runtime.NewScheme()

func init() {
	_ = monitoringv1.SchemeBuilder.AddToScheme(scheme)
	_ = monitoringv1alpha1.SchemeBuilder.AddToScheme(scheme)
	_ = monitoringv1beta1.SchemeBuilder.AddToScheme(scheme)
}

// PodRunningAndReady returns whether a pod is running and each container has
// passed it's ready state.
func PodRunningAndReady(pod v1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case v1.PodFailed, v1.PodSucceeded:
		return false, fmt.Errorf("pod completed")
	case v1.PodRunning:
		for _, cond := range pod.Status.Conditions {
			if cond.Type != v1.PodReady {
				continue
			}
			return cond.Status == v1.ConditionTrue, nil
		}
		return false, fmt.Errorf("pod ready condition not found")
	}
	return false, nil
}

func NewClusterConfig(host string, tlsInsecure bool, tlsConfig *rest.TLSClientConfig) (*rest.Config, error) {
	var cfg *rest.Config
	var err error

	kubeconfigPath := os.Getenv(KubeConfigEnv)
	if kubeconfigPath != "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating config from %s: %w", kubeconfigPath, err)
		}
	} else {
		if len(host) == 0 {
			if cfg, err = rest.InClusterConfig(); err != nil {
				return nil, err
			}
		} else {
			cfg = &rest.Config{
				Host: host,
			}
			hostURL, err := url.Parse(host)
			if err != nil {
				return nil, fmt.Errorf("error parsing host url %s: %w", host, err)
			}
			if hostURL.Scheme == "https" {
				cfg.TLSClientConfig = *tlsConfig
				cfg.Insecure = tlsInsecure
			}
		}
	}

	cfg.QPS = 100
	cfg.Burst = 100

	cfg.UserAgent = fmt.Sprintf("PrometheusOperator/%s", promversion.Version)

	return cfg, nil
}

func IsResourceNotFoundError(err error) bool {
	se, ok := err.(*apierrors.StatusError)
	if !ok {
		return false
	}
	if se.Status().Code == http.StatusNotFound && se.Status().Reason == metav1.StatusReasonNotFound {
		return true
	}
	return false
}

func CreateOrUpdateService(ctx context.Context, sclient clientv1.ServiceInterface, svc *v1.Service) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		service, err := sclient.Get(ctx, svc.Name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}

			_, err = sclient.Create(ctx, svc, metav1.CreateOptions{})
			return err
		}

		// Apply immutable fields from the existing service.
		svc.Spec.IPFamilies = service.Spec.IPFamilies
		svc.Spec.IPFamilyPolicy = service.Spec.IPFamilyPolicy
		svc.Spec.ClusterIP = service.Spec.ClusterIP
		svc.Spec.ClusterIPs = service.Spec.ClusterIPs

		svc.SetOwnerReferences(mergeOwnerReferences(service.GetOwnerReferences(), svc.GetOwnerReferences()))
		mergeMetadata(&svc.ObjectMeta, service.ObjectMeta)

		_, err = sclient.Update(ctx, svc, metav1.UpdateOptions{})
		return err
	})
}

func CreateOrUpdateEndpoints(ctx context.Context, eclient clientv1.EndpointsInterface, eps *v1.Endpoints) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		endpoints, err := eclient.Get(ctx, eps.Name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}

			_, err = eclient.Create(ctx, eps, metav1.CreateOptions{})
			return err
		}

		mergeMetadata(&eps.ObjectMeta, endpoints.ObjectMeta)

		_, err = eclient.Update(ctx, eps, metav1.UpdateOptions{})
		return err
	})
}

// UpdateStatefulSet merges metadata of existing StatefulSet with new one and updates it.
func UpdateStatefulSet(ctx context.Context, sstClient clientappsv1.StatefulSetInterface, sset *appsv1.StatefulSet) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existingSset, err := sstClient.Get(ctx, sset.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		mergeMetadata(&sset.ObjectMeta, existingSset.ObjectMeta)
		// Propagate annotations set by kubectl on spec.template.annotations. e.g performing a rolling restart.
		mergeKubectlAnnotations(&existingSset.Spec.Template.ObjectMeta, sset.Spec.Template.ObjectMeta)

		_, err = sstClient.Update(ctx, sset, metav1.UpdateOptions{})
		return err
	})
}

// CreateOrUpdateSecret merges metadata of existing Secret with new one and updates it.
func CreateOrUpdateSecret(ctx context.Context, secretClient clientv1.SecretInterface, desired *v1.Secret) error {
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

		mutated := existingSecret.DeepCopyObject().(*v1.Secret)
		mergeMetadata(&desired.ObjectMeta, mutated.ObjectMeta)
		if apiequality.Semantic.DeepEqual(existingSecret, desired) {
			return nil
		}
		_, err = secretClient.Update(ctx, desired, metav1.UpdateOptions{})
		return err
	})
}

// IsAPIGroupVersionResourceSupported checks if given groupVersion and resource is supported by the cluster.
//
// you can exec `kubectl api-resources` to find groupVersion and resource.
func IsAPIGroupVersionResourceSupported(discoveryCli discovery.DiscoveryInterface, groupversion string, resource string) (bool, error) {
	apiResourceList, err := discoveryCli.ServerResourcesForGroupVersion(groupversion)
	if err != nil {
		if IsResourceNotFoundError(err) {
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

// ResourceNamer knows how to generate valid names for various Kubernetes resources.
type ResourceNamer struct {
	prefix string
}

// NewResourceNamerWithPrefix returns a ResourceNamer that adds a prefix
// followed by an hyphen character to all resource names.
func NewResourceNamerWithPrefix(p string) ResourceNamer {
	return ResourceNamer{prefix: p}
}

func (rn ResourceNamer) sanitizedVolumeName(name string) string {
	if rn.prefix != "" {
		name = strings.TrimRight(rn.prefix, "-") + "-" + name
	}

	name = strings.ToLower(name)
	name = invalidDNS1123Characters.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	return name
}

func isValidDNS1123Label(name string) error {
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return errors.New(strings.Join(errs, ","))
	}

	return nil
}

// UniqueVolumeName returns a volume name that is a valid DNS-1123 label.
// The returned name has a hash-based suffix to ensure uniqueness in case the
// input name exceeds the 63-chars limit.
func (rn ResourceNamer) UniqueVolumeName(name string) (string, error) {
	// Hash the name and append the 8 first characters of the hash
	// value to the resulting name to ensure that 2 names longer than
	// DNS1123LabelMaxLength return unique volume names.
	// E.g. long-63-chars-abc, long-63-chars-XYZ may be added to
	// volume name since they are trimmed at long-63-chars, there will be 2
	// volume entries with the same name.
	// In practice, the hash is computed for the full name then trimmed to
	// the first 8 chars and added to the end:
	// * long-63-chars-abc -> first-54-chars-deadbeef
	// * long-63-chars-XYZ -> first-54-chars-d3adb33f
	xxh := xxhash.New()
	if _, err := xxh.Write([]byte(name)); err != nil {
		return "", err
	}

	h := fmt.Sprintf("-%x", xxh.Sum64())
	h = h[:9]

	name = rn.sanitizedVolumeName(name)

	if len(name) > validation.DNS1123LabelMaxLength-9 {
		name = name[:validation.DNS1123LabelMaxLength-9]
	}

	name = name + h
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return "", errors.New(strings.Join(errs, ","))
	}

	return name, isValidDNS1123Label(name)
}

// VolumeName returns a volume name that is a valid DNS-1123 label.
func (rn ResourceNamer) VolumeName(name string) (string, error) {
	name = rn.sanitizedVolumeName(name)

	if len(name) > validation.DNS1123LabelMaxLength {
		name = name[:validation.DNS1123LabelMaxLength]
	}

	return name, isValidDNS1123Label(name)
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

func mergeOwnerReferences(old []metav1.OwnerReference, new []metav1.OwnerReference) []metav1.OwnerReference {
	existing := make(map[metav1.OwnerReference]bool)
	for _, ownerRef := range old {
		existing[ownerRef] = true
	}
	for _, ownerRef := range new {
		if _, ok := existing[ownerRef]; !ok {
			old = append(old, ownerRef)
		}
	}
	return old
}

// mergeMetadata takes labels and annotations from the old resource and merges
// them into the new resource. If a key is present in both resources, the new
// resource wins. It also copies the ResourceVersion from the old resource to
// the new resource to prevent update conflicts.
func mergeMetadata(new *metav1.ObjectMeta, old metav1.ObjectMeta) {
	new.ResourceVersion = old.ResourceVersion

	new.SetLabels(mergeMaps(new.Labels, old.Labels))
	new.SetAnnotations(mergeMaps(new.Annotations, old.Annotations))
}

func mergeMaps(new map[string]string, old map[string]string) map[string]string {
	return mergeMapsByPrefix(new, old, "")
}

func mergeKubectlAnnotations(from *metav1.ObjectMeta, to metav1.ObjectMeta) {
	from.SetAnnotations(mergeMapsByPrefix(from.Annotations, to.Annotations, "kubectl.kubernetes.io/"))
}

func mergeMapsByPrefix(from map[string]string, to map[string]string, prefix string) map[string]string {
	if to == nil {
		to = make(map[string]string)
	}

	if from == nil {
		from = make(map[string]string)
	}

	for k, v := range from {
		if strings.HasPrefix(k, prefix) {
			to[k] = v
		}
	}

	return to
}
