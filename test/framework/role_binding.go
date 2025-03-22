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

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) createOrUpdateRoleBinding(ctx context.Context, ns string, cr *rbacv1.ClusterRole, relativePath string) (FinalizerFn, error) {
	return f.createOrUpdateRoleBindingForSubjectNamespace(ctx, ns, "", cr, relativePath)
}

func (f *Framework) createOrUpdateRoleBindingForSubjectNamespace(ctx context.Context, ns, subjectNs string, cr *rbacv1.ClusterRole, source string) (FinalizerFn, error) {
	roleBinding, err := f.parseRoleBindingYaml(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse role binding manifest: %w", err)
	}

	if subjectNs != "" {
		for i := range roleBinding.Subjects {
			roleBinding.Subjects[i].Namespace = subjectNs
		}
	}
	roleBinding.RoleRef.Name = cr.Name

	finalizerFn := func() error { return f.deleteRoleBinding(ctx, ns, roleBinding.Name) }

	_, err = f.KubeClient.RbacV1().RoleBindings(ns).Get(ctx, roleBinding.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err = f.KubeClient.RbacV1().RoleBindings(ns).Create(ctx, roleBinding, metav1.CreateOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to create role binding: %w", err)
			}

			return finalizerFn, nil
		}

		return nil, fmt.Errorf("failed to get role binding: %w", err)
	}

	_, err = f.KubeClient.RbacV1().RoleBindings(ns).Update(ctx, roleBinding, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update role binding: %w", err)
	}

	return finalizerFn, nil
}

func (f *Framework) deleteRoleBinding(ctx context.Context, ns, name string) error {
	err := f.KubeClient.RbacV1().RoleBindings(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

func (f *Framework) parseRoleBindingYaml(source string) (*rbacv1.RoleBinding, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}

	roleBinding := rbacv1.RoleBinding{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&roleBinding); err != nil {
		return nil, err
	}

	return &roleBinding, nil
}
