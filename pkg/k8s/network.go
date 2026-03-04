// Copyright 2026 The prometheus-operator Authors
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

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	clientdiscoveryv1 "k8s.io/client-go/kubernetes/typed/discovery/v1"
	"k8s.io/client-go/util/retry"
)

// CreateOrUpdateService creates or updates a Service resource.
func CreateOrUpdateService(ctx context.Context, sclient clientv1.ServiceInterface, svc *v1.Service) (*v1.Service, error) {
	var ret *v1.Service

	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		service, err := sclient.Get(ctx, svc.Name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}

			ret, err = sclient.Create(ctx, svc, metav1.CreateOptions{})
			return err
		}

		// Apply immutable fields from the existing service.
		svc.Spec.IPFamilies = service.Spec.IPFamilies
		svc.Spec.IPFamilyPolicy = service.Spec.IPFamilyPolicy
		svc.Spec.ClusterIP = service.Spec.ClusterIP
		svc.Spec.ClusterIPs = service.Spec.ClusterIPs

		svc.SetOwnerReferences(mergeOwnerReferences(service.GetOwnerReferences(), svc.GetOwnerReferences()))
		mergeMetadata(&svc.ObjectMeta, service.ObjectMeta)

		ret, err = sclient.Update(ctx, svc, metav1.UpdateOptions{})
		return err
	})

	return ret, err
}

func mergeOwnerReferences(oldObj []metav1.OwnerReference, newObj []metav1.OwnerReference) []metav1.OwnerReference {
	existing := make(map[metav1.OwnerReference]bool)
	for _, ownerRef := range oldObj {
		existing[ownerRef] = true
	}
	for _, ownerRef := range newObj {
		if _, ok := existing[ownerRef]; !ok {
			oldObj = append(oldObj, ownerRef)
		}
	}
	return oldObj
}

// CreateOrUpdateEndpoints creates or updates an Endpoints resource.
//
//nolint:staticcheck // Ignore SA1019 Endpoints is marked as deprecated.
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

// CreateOrUpdateEndpointSlice creates or updates an EndpointSlice resource.
func CreateOrUpdateEndpointSlice(ctx context.Context, c clientdiscoveryv1.EndpointSliceInterface, eps *discoveryv1.EndpointSlice) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if eps.Name == "" {
			_, err := c.Create(ctx, eps, metav1.CreateOptions{})
			return err
		}

		endpoints, err := c.Get(ctx, eps.Name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}

			_, err = c.Create(ctx, eps, metav1.CreateOptions{})
			return err
		}

		mergeMetadata(&eps.ObjectMeta, endpoints.ObjectMeta)

		_, err = c.Update(ctx, eps, metav1.UpdateOptions{})
		return err
	})
}

// EnsureCustomGoverningService is responsible for the following:
//
// Verify that the service exists in the resource's namespace
// If it does not exist, fail the reconciliation.
//
// If the ServiceName is specified and a service with the same name exists in the same namespace as the
// resource, ensure that the custom governing service's selector matches the
// labels.
// If it is not selected, fail the reconciliation
// Warning: the function will panic if the resource's ServiceName is nil..
func EnsureCustomGoverningService(ctx context.Context, namespace string, serviceName string, svcClient clientv1.ServiceInterface, selectorLabels map[string]string) error {
	// Check if the custom governing service is defined in the same namespace and selects the Prometheus pod.
	svc, err := svcClient.Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get custom governing service %s/%s: %w", namespace, serviceName, err)
	}

	svcSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: svc.Spec.Selector})
	if err != nil {
		return fmt.Errorf("failed to parse the selector labels for custom governing service %s/%s: %w", namespace, serviceName, err)
	}

	if !svcSelector.Matches(labels.Set(selectorLabels)) {
		return fmt.Errorf("custom governing service %s/%s with selector %q does not select pods with labels %q",
			namespace, serviceName, svcSelector.String(), labels.Set(selectorLabels).String())
	}
	return nil
}
