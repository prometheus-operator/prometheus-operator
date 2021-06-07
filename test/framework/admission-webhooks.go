// Copyright 2019 The prometheus-operator Authors
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
	"github.com/pkg/errors"
	"k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

func createMutatingHook(kubeClient kubernetes.Interface, certBytes []byte, namespace, yamlPath string) (FinalizerFn, error) {
	h, err := parseMutatingHookYaml(yamlPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed parsing mutating webhook")
	}

	h.Webhooks[0].ClientConfig.Service.Namespace = namespace
	h.Webhooks[0].ClientConfig.CABundle = certBytes

	_, err = kubeClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(context.TODO(), h, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create mutating webhook %s", h.Name)
	}

	finalizerFn := func() error { return deleteMutatingWebhook(kubeClient, h.Name) }

	return finalizerFn, nil
}

func createValidatingHook(kubeClient kubernetes.Interface, certBytes []byte, namespace, yamlPath string) (FinalizerFn, error) {
	h, err := parseValidatingHookYaml(yamlPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed parsing mutating webhook")
	}

	h.Webhooks[0].ClientConfig.Service.Namespace = namespace
	h.Webhooks[0].ClientConfig.CABundle = certBytes

	_, err = kubeClient.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(context.TODO(), h, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create validating webhook %s", h.Name)
	}

	finalizerFn := func() error { return deleteValidatingWebhook(kubeClient, h.Name) }

	return finalizerFn, nil
}

func deleteMutatingWebhook(kubeClient kubernetes.Interface, name string) error {
	return kubeClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func deleteValidatingWebhook(kubeClient kubernetes.Interface, name string) error {
	return kubeClient.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func parseValidatingHookYaml(pathToYaml string) (*v1beta1.ValidatingWebhookConfiguration, error) {
	manifest, err := PathToOSFile(pathToYaml)
	if err != nil {
		return nil, err
	}

	resource := v1beta1.ValidatingWebhookConfiguration{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&resource); err != nil {
		return nil, errors.Wrapf(err, "failed to decode file %s", pathToYaml)
	}

	return &resource, nil
}

func parseMutatingHookYaml(pathToYaml string) (*v1beta1.MutatingWebhookConfiguration, error) {
	manifest, err := PathToOSFile(pathToYaml)
	if err != nil {
		return nil, err
	}

	resource := v1beta1.MutatingWebhookConfiguration{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&resource); err != nil {
		return nil, errors.Wrapf(err, "failed to decode file %s", pathToYaml)
	}

	return &resource, nil
}
