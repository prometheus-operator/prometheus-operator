package framework

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	rbacv1alpha1 "k8s.io/client-go/pkg/apis/rbac/v1alpha1"
	"k8s.io/client-go/pkg/util/yaml"
)

func CreateClusterRoleBinding(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRoleBinding, err := parseClusterRoleBindingYaml(relativePath)
	if err != nil {
		return err
	}

	_, err = kubeClient.RbacV1alpha1().ClusterRoleBindings().Get(clusterRoleBinding.Name)

	if err == nil {
		// ClusterRoleBinding already exists -> Update
		_, err = kubeClient.RbacV1alpha1().ClusterRoleBindings().Update(clusterRoleBinding)
		if err != nil {
			return err
		}
	} else {
		// ClusterRoleBinding doesn't exists -> Create
		_, err = kubeClient.RbacV1alpha1().ClusterRoleBindings().Create(clusterRoleBinding)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteClusterRoleBinding(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRoleBinding, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return err
	}

	if err := kubeClient.RbacV1alpha1().ClusterRoleBindings().Delete(clusterRoleBinding.Name, &v1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func parseClusterRoleBindingYaml(relativePath string) (*rbacv1alpha1.ClusterRoleBinding, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	clusterRoleBinding := rbacv1alpha1.ClusterRoleBinding{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&clusterRoleBinding); err != nil {
		return nil, err
	}

	return &clusterRoleBinding, nil
}
