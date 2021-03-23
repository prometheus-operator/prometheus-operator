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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

func createServiceAccount(kubeClient kubernetes.Interface, namespace string, relativPath string) (FinalizerFn, error) {
	finalizerFn := func() error { return DeleteServiceAccount(kubeClient, namespace, relativPath) }

	serviceAccount, err := parseServiceAccountYaml(relativPath)
	if err != nil {
		return finalizerFn, err
	}
	serviceAccount.Namespace = namespace
	_, err = kubeClient.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return finalizerFn, err
	}

	return finalizerFn, nil
}

func parseServiceAccountYaml(relativPath string) (*v1.ServiceAccount, error) {
	manifest, err := PathToOSFile(relativPath)
	if err != nil {
		return nil, err
	}

	serviceAccount := v1.ServiceAccount{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&serviceAccount); err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}

func DeleteServiceAccount(kubeClient kubernetes.Interface, namespace string, relativPath string) error {
	serviceAccount, err := parseServiceAccountYaml(relativPath)
	if err != nil {
		return err
	}

	return kubeClient.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), serviceAccount.Name, metav1.DeleteOptions{})
}
