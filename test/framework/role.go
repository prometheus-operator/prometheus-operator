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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	roleCRDCreateRule = rbacv1.PolicyRule{
		APIGroups: []string{"apiextensions.k8s.io"},
		Resources: []string{"customresourcedefinitions"},
		Verbs:     []string{"create"},
	}

	roleCRDMonitoringRule = rbacv1.PolicyRule{
		APIGroups: []string{"apiextensions.k8s.io"},
		Resources: []string{"customresourcedefinitions"},
		ResourceNames: []string{
			"alertmanagers.monitoring.coreos.com",
			"podmonitors.monitoring.coreos.com",
			"probes.monitoring.coreos.com",
			"prometheuses.monitoring.coreos.com",
			"prometheusrules.monitoring.coreos.com",
			"servicemonitors.monitoring.coreos.com",
			"thanosrulers.monitoring.coreos.com",
		},
		Verbs: []string{"get", "update"},
	}
)

func (f *Framework) CreateRole(ctx context.Context, ns, relativePath string) (*rbacv1.Role, error) {
	role, err := parseRoleYaml(relativePath)
	if err != nil {
		return nil, err
	}

	_, err = f.KubeClient.RbacV1().Roles(ns).Get(ctx, role.Name, metav1.GetOptions{})

	if err == nil {
		// ClusterRole already exists -> Update
		role, err = f.KubeClient.RbacV1().Roles(ns).Update(ctx, role, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}

	} else {
		// ClusterRole doesn't exists -> Create
		role, err = f.KubeClient.RbacV1().Roles(ns).Create(ctx, role, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return role, nil
}

func (f *Framework) DeleteRole(ctx context.Context, ns, relativePath string) error {
	role, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return err
	}

	return f.KubeClient.RbacV1().Roles(ns).Delete(ctx, role.Name, metav1.DeleteOptions{})
}

func (f *Framework) UpdateRole(ctx context.Context, ns string, role *rbacv1.Role) error {
	_, err := f.KubeClient.RbacV1().Roles(ns).Update(ctx, role, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func parseRoleYaml(relativePath string) (*rbacv1.Role, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	role := rbacv1.Role{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&role); err != nil {
		return nil, err
	}

	return &role, nil
}
