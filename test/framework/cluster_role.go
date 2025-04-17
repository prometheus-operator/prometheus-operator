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

	"github.com/cespare/xxhash/v2"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	CRDCreateRule = rbacv1.PolicyRule{
		APIGroups: []string{"apiextensions.k8s.io"},
		Resources: []string{"customresourcedefinitions"},
		Verbs:     []string{"create"},
	}

	CRDMonitoringRule = rbacv1.PolicyRule{
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

func (f *Framework) CreateOrUpdateClusterRole(ctx context.Context, cr *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	_, err := f.KubeClient.RbacV1().ClusterRoles().Get(ctx, cr.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		if apierrors.IsNotFound(err) {
			// ClusterRole doesn't exists -> Create
			return f.KubeClient.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
		}
	}

	// ClusterRole already exists -> Update
	return f.KubeClient.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
}

func (f *Framework) DeleteClusterRole(ctx context.Context, name string) error {
	return f.KubeClient.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
}

func (f *Framework) UpdateClusterRole(ctx context.Context, clusterRole *rbacv1.ClusterRole) error {
	_, err := f.KubeClient.RbacV1().ClusterRoles().Update(ctx, clusterRole, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func clusterRoleFromYaml(suffix, source string) (*rbacv1.ClusterRole, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}

	clusterRole := rbacv1.ClusterRole{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&clusterRole); err != nil {
		return nil, err
	}

	// Use a unique cluster role name to avoid parallel tests doing concurrent
	// updates to the same resource.
	if suffix != "" {
		xxh := xxhash.New()
		if _, err := xxh.Write([]byte(suffix)); err != nil {
			// Write() never returns nil.
			panic(fmt.Errorf("failed to write hash: %w", err))
		}

		clusterRole.Name = fmt.Sprintf("%s-%x", clusterRole.Name, xxh.Sum64())
	}

	return &clusterRole, nil
}
