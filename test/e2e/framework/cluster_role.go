package framework

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	rbacv1alpha1 "k8s.io/client-go/pkg/apis/rbac/v1alpha1"
	"k8s.io/client-go/pkg/util/yaml"
)

func CreateClusterRole(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRole, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return err
	}

	_, err = kubeClient.RbacV1alpha1().ClusterRoles().Get(clusterRole.Name)

	if err == nil {
		// ClusterRole already exists -> Update
		_, err = kubeClient.RbacV1alpha1().ClusterRoles().Update(clusterRole)
		if err != nil {
			return err
		}

	} else {
		// ClusterRole doesn't exists -> Create
		_, err = kubeClient.RbacV1alpha1().ClusterRoles().Create(clusterRole)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteClusterRole(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRole, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return err
	}

	if err := kubeClient.RbacV1alpha1().ClusterRoles().Delete(clusterRole.Name, &v1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func parseClusterRoleYaml(relativePath string) (*rbacv1alpha1.ClusterRole, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	clusterRole := rbacv1alpha1.ClusterRole{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&clusterRole); err != nil {
		return nil, err
	}

	return &clusterRole, nil
}
