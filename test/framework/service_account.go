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

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) createOrUpdateServiceAccount(ctx context.Context, namespace string, source string) (FinalizerFn, error) {
	finalizerFn := func() error { return f.DeleteServiceAccount(ctx, namespace, source) }

	serviceAccount, err := parseServiceAccountYaml(source)
	if err != nil {
		return finalizerFn, err
	}
	serviceAccount.Namespace = namespace
	_, err = f.KubeClient.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccount.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return finalizerFn, err
	}

	if apierrors.IsNotFound(err) {
		// ServiceAccount doesn't exists -> Create
		_, err = f.KubeClient.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
		if err != nil {
			return finalizerFn, err
		}
	} else {
		// ServiceAccount already exists -> Update
		_, err = f.KubeClient.CoreV1().ServiceAccounts(namespace).Update(ctx, serviceAccount, metav1.UpdateOptions{})
		if err != nil {
			return finalizerFn, err
		}
	}

	return finalizerFn, nil
}

func parseServiceAccountYaml(source string) (*v1.ServiceAccount, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}

	serviceAccount := v1.ServiceAccount{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&serviceAccount); err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}

func (f *Framework) DeleteServiceAccount(ctx context.Context, namespace string, source string) error {
	serviceAccount, err := parseServiceAccountYaml(source)
	if err != nil {
		return err
	}

	return f.KubeClient.CoreV1().ServiceAccounts(namespace).Delete(ctx, serviceAccount.Name, metav1.DeleteOptions{})
}
