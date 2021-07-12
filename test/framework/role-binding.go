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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) CreateRoleBinding(ns string, relativePath string) (FinalizerFn, error) {
	finalizerFn := func() error { return f.DeleteRoleBinding(ns, relativePath) }
	roleBinding, err := f.parseRoleBindingYaml(relativePath)
	if err != nil {
		return finalizerFn, err
	}

	_, err = f.KubeClient.RbacV1().RoleBindings(ns).Create(f.Ctx, roleBinding, metav1.CreateOptions{})
	return finalizerFn, err
}

func (f *Framework) CreateRoleBindingForSubjectNamespace(ns, subjectNs string, relativePath string) (FinalizerFn, error) {
	finalizerFn := func() error { return f.DeleteRoleBinding(ns, relativePath) }
	roleBinding, err := f.parseRoleBindingYaml(relativePath)

	for i := range roleBinding.Subjects {
		roleBinding.Subjects[i].Namespace = subjectNs
	}

	if err != nil {
		return finalizerFn, err
	}

	_, err = f.KubeClient.RbacV1().RoleBindings(ns).Create(f.Ctx, roleBinding, metav1.CreateOptions{})
	return finalizerFn, err
}

func (f *Framework) DeleteRoleBinding(ns string, relativePath string) error {
	roleBinding, err := f.parseRoleBindingYaml(relativePath)
	if err != nil {
		return err
	}

	return f.KubeClient.RbacV1().RoleBindings(ns).Delete(f.Ctx, roleBinding.Name, metav1.DeleteOptions{})
}

func (f *Framework) parseRoleBindingYaml(relativePath string) (*rbacv1.RoleBinding, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	roleBinding := rbacv1.RoleBinding{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&roleBinding); err != nil {
		return nil, err
	}

	return &roleBinding, nil
}
