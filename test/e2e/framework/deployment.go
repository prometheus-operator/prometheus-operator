package framework

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
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
	d, err := kubeClient.Extensions().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	zero := int32(0)
	d.Spec.Replicas = &zero

	d, err = kubeClient.Extensions().Deployments(namespace).Update(d)
	if err != nil {
		return err
	}
	return kubeClient.Extensions().Deployments(namespace).Delete(d.Name, &metav1.DeleteOptions{})
}
