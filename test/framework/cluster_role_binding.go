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

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) createOrUpdateClusterRoleBinding(ctx context.Context, ns string, source string) (FinalizerFn, error) {
	finalizerFn := func() error { return f.DeleteClusterRoleBinding(ctx, ns, source) }
	clusterRoleBinding, err := parseClusterRoleBindingYaml(source)
	if err != nil {
		return finalizerFn, err
	}

	// Make sure to create a new cluster role binding for each namespace to
	// prevent concurrent tests to delete each others bindings.
	clusterRoleBinding.Name = ns + "-" + clusterRoleBinding.Name

	clusterRoleBinding.Subjects[0].Namespace = ns

	_, err = f.KubeClient.RbacV1().ClusterRoleBindings().Get(ctx, clusterRoleBinding.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return finalizerFn, err
	}

	if apierrors.IsNotFound(err) {
		// ClusterRoleBinding doesn't exists -> Create
		_, err = f.KubeClient.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
		if err != nil {
			return finalizerFn, err
		}
	} else {
		// ClusterRoleBinding already exists -> Update
		_, err = f.KubeClient.RbacV1().ClusterRoleBindings().Update(ctx, clusterRoleBinding, metav1.UpdateOptions{})
		if err != nil {
			return finalizerFn, err
		}
	}

	return finalizerFn, err
}

func (f *Framework) DeleteClusterRoleBinding(ctx context.Context, ns string, source string) error {
	clusterRoleBinding, err := parseClusterRoleYaml(source)
	if err != nil {
		return err
	}

	// Make sure to delete the specific cluster role binding for the namespace
	// it was created preventing concurrent tests to delete each others bindings.
	clusterRoleBinding.Name = ns + "-" + clusterRoleBinding.Name

	return f.KubeClient.RbacV1().ClusterRoleBindings().Delete(ctx, clusterRoleBinding.Name, metav1.DeleteOptions{})
}

func parseClusterRoleBindingYaml(source string) (*rbacv1.ClusterRoleBinding, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}

	clusterRoleBinding := rbacv1.ClusterRoleBinding{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&clusterRoleBinding); err != nil {
		return nil, err
	}

	return &clusterRoleBinding, nil
}
