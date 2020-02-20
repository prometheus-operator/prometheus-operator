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

func GetDeployment(kubeCilent kubernetes.Interface, ns, name string) (*appsv1.Deployment, error) {
	return kubeCilent.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{})
}

func UpdateDeployment(kubeCilent kubernetes.Interface, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return kubeCilent.AppsV1().Deployments(deployment.Namespace).Update(deployment)
}

func MakeDeployment(pathToYaml string) (*appsv1.Deployment, error) {
	manifest, err := PathToOSFile(pathToYaml)
	if err != nil {
		return nil, err
	}
	deployment := appsv1.Deployment{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&deployment); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to decode file %s", pathToYaml))
	}

	return &deployment, nil
}

func CreateDeployment(kubeClient kubernetes.Interface, namespace string, d *appsv1.Deployment) error {
	d.Namespace = namespace
	_, err := kubeClient.AppsV1().Deployments(namespace).Create(d)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create deployment %s", d.Name))
	}
	return nil
}

func DeleteDeployment(kubeClient kubernetes.Interface, namespace, name string) error {
	d, err := kubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	zero := int32(0)
	d.Spec.Replicas = &zero

	d, err = kubeClient.AppsV1().Deployments(namespace).Update(d)
	if err != nil {
		return err
	}
	return kubeClient.AppsV1beta2().Deployments(namespace).Delete(d.Name, &metav1.DeleteOptions{})
}

func WaitUntilDeploymentGone(kubeClient kubernetes.Interface, namespace, name string, timeout time.Duration) error {
	return wait.Poll(time.Second, timeout, func() (bool, error) {
		_, err := kubeClient.
			AppsV1beta2().Deployments(namespace).
			Get(name, metav1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}
