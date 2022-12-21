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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) createOrUpdateMutatingHook(ctx context.Context, certBytes []byte, namespace, source string) (FinalizerFn, error) {
	hook, err := parseMutatingHookYaml(source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed parsing mutating webhook")
	}

	hook.Webhooks[0].ClientConfig.Service.Namespace = namespace
	hook.Webhooks[0].ClientConfig.CABundle = certBytes

	h, err := f.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, hook.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get mutating webhook %s", hook.Name)
	}

	if apierrors.IsNotFound(err) {
		// MutatingWebhookConfiguration doesn't exists -> Create
		_, err = f.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, hook, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create mutating webhook %s", hook.Name)
		}
	} else {
		// must set this field from existing MutatingWebhookConfiguration to prevent update fail
		hook.ObjectMeta.ResourceVersion = h.ObjectMeta.ResourceVersion

		// MutatingWebhookConfiguration already exists -> Update
		_, err = f.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, hook, metav1.UpdateOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update mutating webhook %s", hook.Name)
		}
	}

	finalizerFn := func() error { return f.deleteMutatingWebhook(ctx, hook.Name) }

	return finalizerFn, nil
}

func (f *Framework) createOrUpdateValidatingHook(ctx context.Context, certBytes []byte, namespace, source string) (FinalizerFn, error) {
	hook, err := parseValidatingHookYaml(source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed parsing validating webhook")
	}

	hook.Webhooks[0].ClientConfig.Service.Namespace = namespace
	hook.Webhooks[0].ClientConfig.CABundle = certBytes

	h, err := f.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(ctx, hook.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get validating webhook %s", hook.Name)
	}

	if apierrors.IsNotFound(err) {
		// ValidatingWebhookConfiguration doesn't exists -> Create
		_, err = f.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(ctx, hook, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create validating webhook %s", hook.Name)
		}
	} else {
		// must set this field from existing ValidatingWebhookConfiguration to prevent update fail
		hook.ObjectMeta.ResourceVersion = h.ObjectMeta.ResourceVersion

		// ValidatingWebhookConfiguration already exists -> Update
		_, err = f.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(ctx, hook, metav1.UpdateOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update validating webhook %s", hook.Name)
		}
	}

	finalizerFn := func() error { return f.deleteValidatingWebhook(ctx, hook.Name) }

	return finalizerFn, nil
}

func (f *Framework) deleteMutatingWebhook(ctx context.Context, name string) error {
	return f.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, name, metav1.DeleteOptions{})
}

func (f *Framework) deleteValidatingWebhook(ctx context.Context, name string) error {
	return f.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(ctx, name, metav1.DeleteOptions{})
}

func parseValidatingHookYaml(source string) (*v1.ValidatingWebhookConfiguration, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}

	resource := v1.ValidatingWebhookConfiguration{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&resource); err != nil {
		return nil, errors.Wrapf(err, "failed to decode file %s", source)
	}

	return &resource, nil
}

func parseMutatingHookYaml(source string) (*v1.MutatingWebhookConfiguration, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, err
	}

	resource := v1.MutatingWebhookConfiguration{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&resource); err != nil {
		return nil, errors.Wrapf(err, "failed to decode file %s", source)
	}

	return &resource, nil
}
