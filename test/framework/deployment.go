// Copyright 2017 The prometheus-operator Authors
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
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

func (f *Framework) GetDeployment(ctx context.Context, ns, name string) (*appsv1.Deployment, error) {
	return f.KubeClient.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
}

func (f *Framework) UpdateDeployment(ctx context.Context, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return f.KubeClient.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

func MakeDeployment(source string) (*appsv1.Deployment, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}
	deployment := appsv1.Deployment{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&deployment); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to decode file %s", source))
	}

	return &deployment, nil
}

func (f *Framework) CreateDeployment(ctx context.Context, namespace string, d *appsv1.Deployment) error {
	d.Namespace = namespace
	_, err := f.KubeClient.AppsV1().Deployments(namespace).Create(ctx, d, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create deployment %s", d.Name))
	}
	return nil
}

func (f *Framework) CreateOrUpdateDeploymentAndWaitUntilReady(ctx context.Context, namespace string, deployment *appsv1.Deployment) error {
	deployment.Namespace = namespace
	d, err := f.KubeClient.AppsV1().Deployments(namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, fmt.Sprintf("failed to get deployment %s", deployment.Name))
	}

	if apierrors.IsNotFound(err) {
		// Deployment doesn't exists -> Create
		_, err = f.KubeClient.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create deployment %s", deployment.Name))
		}

		err = f.WaitForDeploymentReady(ctx, namespace, deployment.Name, 1)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("after create, waiting for deployment %v to become ready timed out", deployment.Name))
		}
	} else {
		// Deployment already exists -> Update
		_, err = f.KubeClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update deployment %s", deployment.Name))
		}

		err = f.WaitForDeploymentReady(ctx, namespace, deployment.Name, d.Status.ObservedGeneration+1)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("after update, waiting for deployment %v to become ready timed out", deployment.Name))
		}
	}

	return nil
}

func (f *Framework) WaitForDeploymentReady(ctx context.Context, namespace, deploymentName string, expectedGeneration int64) error {
	err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute*5, false, func(ctx context.Context) (bool, error) {
		d, err := f.KubeClient.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if d.Status.ObservedGeneration == expectedGeneration && *d.Spec.Replicas == d.Status.AvailableReplicas {
			return true, nil
		}
		return false, nil
	})
	return err
}

func (f *Framework) DeleteDeployment(ctx context.Context, namespace, name string) error {
	d, err := f.KubeClient.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	zero := int32(0)
	d.Spec.Replicas = &zero

	d, err = f.KubeClient.AppsV1().Deployments(namespace).Update(ctx, d, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return f.KubeClient.AppsV1beta2().Deployments(namespace).Delete(ctx, d.Name, metav1.DeleteOptions{})
}

func (f *Framework) WaitUntilDeploymentGone(ctx context.Context, kubeClient kubernetes.Interface, namespace, name string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		_, err := f.KubeClient.
			AppsV1beta2().Deployments(namespace).
			Get(ctx, name, metav1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}
