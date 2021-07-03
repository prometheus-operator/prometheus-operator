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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) createServiceAccount(namespace string, relativePath string) (FinalizerFn, error) {
	finalizerFn := func() error { return f.DeleteServiceAccount(namespace, relativePath) }

	serviceAccount, err := parseServiceAccountYaml(relativePath)
	if err != nil {
		return finalizerFn, err
	}
	serviceAccount.Namespace = namespace
	_, err = f.KubeClient.CoreV1().ServiceAccounts(namespace).Create(f.Ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return finalizerFn, err
	}

	return finalizerFn, nil
}

func parseServiceAccountYaml(relativePath string) (*v1.ServiceAccount, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	serviceAccount := v1.ServiceAccount{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&serviceAccount); err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}

func (f *Framework) DeleteServiceAccount(namespace string, relativePath string) error {
	serviceAccount, err := parseServiceAccountYaml(relativePath)
	if err != nil {
		return err
	}

	return f.KubeClient.CoreV1().ServiceAccounts(namespace).Delete(f.Ctx, serviceAccount.Name, metav1.DeleteOptions{})
}
