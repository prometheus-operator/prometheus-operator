package framework

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/yaml"
)

func MakeDeployment(pathToYaml string) (*v1beta1.Deployment, error) {
	manifest, err := PathToOSFile(pathToYaml)
	if err != nil {
		return nil, err
	}
	tectonicPromOp := v1beta1.Deployment{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&tectonicPromOp); err != nil {
		return nil, err
	}

	return &tectonicPromOp, nil
}

func CreateDeployment(kubeClient kubernetes.Interface, namespace string, d *v1beta1.Deployment) error {
	_, err := kubeClient.Extensions().Deployments(namespace).Create(d)
	if err != nil {
		return err
	}
	return nil
}

func DeleteDeployment(kubeClient kubernetes.Interface, namespace, name string) error {
	d, err := kubeClient.Extensions().Deployments(namespace).Get(name)
	if err != nil {
		return err
	}

	zero := int32(0)
	d.Spec.Replicas = &zero

	d, err = kubeClient.Extensions().Deployments(namespace).Update(d)
	if err != nil {
		return err
	}
	return kubeClient.Extensions().Deployments(namespace).Delete(d.Name, &v1.DeleteOptions{})
}
