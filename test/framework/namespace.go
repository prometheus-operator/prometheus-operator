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
	"maps"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
)

func (f *Framework) CreateNamespace(ctx context.Context, t *testing.T, testCtx *TestCtx) string {
	name := testCtx.ID()
	rn := k8s.ResourceNamer{}
	name, err := rn.UniqueDNS1123Label(name)
	if err != nil {
		t.Fatalf("failed to generate namespace %v: %v", name, err)
	}

	_, err = f.KubeClient.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"app.kubernetes.io/created-by": "e2e-test"},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespace %q: %v", name, err)
	}

	testCtx.namespaces = append(testCtx.namespaces, name)

	namespaceFinalizerFn := func() error {
		return f.DeleteNamespace(ctx, name)
	}

	testCtx.AddFinalizerFn(namespaceFinalizerFn)

	return name
}

func (f *Framework) DeleteNamespace(ctx context.Context, name string) error {
	return f.KubeClient.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

func (f *Framework) AddLabelsToNamespace(ctx context.Context, name string, additionalLabels map[string]string) error {
	ns, err := f.KubeClient.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}

	maps.Copy(ns.Labels, additionalLabels)

	_, err = f.KubeClient.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (f *Framework) RemoveLabelsFromNamespace(ctx context.Context, name string, labels ...string) error {
	b, err := removeLabelsPatch(labels...)
	if err != nil {
		return err
	}

	_, err = f.KubeClient.CoreV1().Namespaces().Patch(ctx, name, types.JSONPatchType, b, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}
