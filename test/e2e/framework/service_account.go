package framework

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/yaml"
)

func CreateServiceAccount(kubeClient kubernetes.Interface, namespace string, relativPath string) error {
	manifest, err := PathToOSFile(relativPath)
	if err != nil {
		return err
	}

	serviceAccount := v1.ServiceAccount{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&serviceAccount); err != nil {
		return err
	}

	_, err = kubeClient.CoreV1().ServiceAccounts(namespace).Create(&serviceAccount)
	if err != nil {
		return err
	}

	return nil
}
