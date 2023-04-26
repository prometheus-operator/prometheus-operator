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
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func MakeService(source string) (*v1.Service, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}
	resource := v1.Service{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&resource); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to decode file %s", source))
	}

	return &resource, nil
}

func (f *Framework) CreateOrUpdateServiceAndWaitUntilReady(ctx context.Context, namespace string, service *v1.Service) (FinalizerFn, error) {
	finalizerFn := func() error { return f.DeleteServiceAndWaitUntilGone(ctx, namespace, service.Name) }

	s, err := f.KubeClient.CoreV1().Services(namespace).Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return finalizerFn, errors.Wrap(err, fmt.Sprintf("getting service %v failed", service.Name))
	}

	if apierrors.IsNotFound(err) {
		// Service doesn't exists -> Create
		if _, err := f.KubeClient.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
			return finalizerFn, errors.Wrap(err, fmt.Sprintf("creating service %v failed", service.Name))
		}
	} else {
		// must set these immutable fields from the existing service to prevent update fail
		service.ObjectMeta.ResourceVersion = s.ObjectMeta.ResourceVersion
		service.Spec.ClusterIP = s.Spec.ClusterIP
		service.Spec.ClusterIPs = s.Spec.ClusterIPs
		service.Spec.IPFamilies = s.Spec.IPFamilies

		// Service already exists -> Update
		if _, err := f.KubeClient.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{}); err != nil {
			return finalizerFn, errors.Wrap(err, fmt.Sprintf("updating service %v failed", service.Name))
		}
	}

	if err := f.WaitForServiceReady(ctx, namespace, service.Name); err != nil {
		return finalizerFn, errors.Wrap(err, fmt.Sprintf("waiting for service %v to become ready timed out", service.Name))
	}
	return finalizerFn, nil
}

func (f *Framework) WaitForServiceReady(ctx context.Context, namespace string, serviceName string) error {
	err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute*5, false, func(ctx context.Context) (bool, error) {
		endpoints, err := f.getEndpoints(ctx, namespace, serviceName)
		if err != nil {
			return false, err
		}
		if len(endpoints.Subsets) != 0 && len(endpoints.Subsets[0].Addresses) > 0 {
			return true, nil
		}
		return false, nil
	})
	return err
}

func (f *Framework) DeleteServiceAndWaitUntilGone(ctx context.Context, namespace string, serviceName string) error {
	if err := f.KubeClient.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("deleting service %v failed", serviceName))
	}

	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, time.Minute, false, func(ctx context.Context) (bool, error) {
		_, err := f.getEndpoints(ctx, namespace, serviceName)
		if err != nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "waiting for service to go away failed")
	}

	return nil
}

func (f *Framework) getEndpoints(ctx context.Context, namespace, serviceName string) (*v1.Endpoints, error) {
	endpoints, err := f.KubeClient.CoreV1().Endpoints(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("requesting endpoints for service %v failed", serviceName))
	}
	return endpoints, nil
}
