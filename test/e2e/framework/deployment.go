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
