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
	"k8s.io/client-go/kubernetes"
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

func CreateClusterRole(kubeClient kubernetes.Interface, relativePath string) (*rbacv1.ClusterRole, error) {
	clusterRole, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return nil, err
	}

	_, err = kubeClient.RbacV1().ClusterRoles().Get(context.TODO(), clusterRole.Name, metav1.GetOptions{})

	if err == nil {
		// ClusterRole already exists -> Update
		clusterRole, err = kubeClient.RbacV1().ClusterRoles().Update(context.TODO(), clusterRole, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}

	} else {
		// ClusterRole doesn't exists -> Create
		clusterRole, err = kubeClient.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return clusterRole, nil
}

func DeleteClusterRole(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRole, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return err
	}

	return kubeClient.RbacV1().ClusterRoles().Delete(context.TODO(), clusterRole.Name, metav1.DeleteOptions{})
}

func UpdateClusterRole(kubeClient kubernetes.Interface, clusterRole *rbacv1.ClusterRole) error {
	_, err := kubeClient.RbacV1().ClusterRoles().Update(context.TODO(), clusterRole, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func parseClusterRoleYaml(relativePath string) (*rbacv1.ClusterRole, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	clusterRole := rbacv1.ClusterRole{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&clusterRole); err != nil {
		return nil, err
	}

	return &clusterRole, nil
}
