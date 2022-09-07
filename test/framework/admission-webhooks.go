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
	v1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) createMutatingHook(ctx context.Context, certBytes []byte, namespace, yamlPath string) (FinalizerFn, error) {
	h, err := parseMutatingHookYaml(yamlPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed parsing mutating webhook")
	}

	h.Webhooks[0].ClientConfig.Service.Namespace = namespace
	h.Webhooks[0].ClientConfig.CABundle = certBytes

	_, err = f.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, h, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create mutating webhook %s", h.Name)
	}

	finalizerFn := func() error { return f.deleteMutatingWebhook(ctx, h.Name) }

	return finalizerFn, nil
}

func (f *Framework) createValidatingHook(ctx context.Context, certBytes []byte, namespace, yamlPath string) (FinalizerFn, error) {
	h, err := parseValidatingHookYaml(yamlPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed parsing validating webhook")
	}

	h.Webhooks[0].ClientConfig.Service.Namespace = namespace
	h.Webhooks[0].ClientConfig.CABundle = certBytes

	_, err = f.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(ctx, h, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create validating webhook %s", h.Name)
	}

	finalizerFn := func() error { return f.deleteValidatingWebhook(ctx, h.Name) }

	return finalizerFn, nil
}

func (f *Framework) deleteMutatingWebhook(ctx context.Context, name string) error {
	return f.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, name, metav1.DeleteOptions{})
}

func (f *Framework) deleteValidatingWebhook(ctx context.Context, name string) error {
	return f.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(ctx, name, metav1.DeleteOptions{})
}

func parseValidatingHookYaml(pathToYaml string) (*v1.ValidatingWebhookConfiguration, error) {
	manifest, err := PathToOSFile(pathToYaml)
	if err != nil {
		return nil, err
	}

	resource := v1.ValidatingWebhookConfiguration{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&resource); err != nil {
		return nil, errors.Wrapf(err, "failed to decode file %s", pathToYaml)
	}

	return &resource, nil
}

func parseMutatingHookYaml(pathToYaml string) (*v1.MutatingWebhookConfiguration, error) {
	manifest, err := PathToOSFile(pathToYaml)
	if err != nil {
		return nil, err
	}

	resource := v1.MutatingWebhookConfiguration{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&resource); err != nil {
		return nil, errors.Wrapf(err, "failed to decode file %s", pathToYaml)
	}

	return &resource, nil
}
