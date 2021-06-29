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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (f *Framework) CreateNamespace(t *testing.T, ctx *TestCtx) string {
	name := ctx.GetObjID()

	_, err := f.KubeClient.CoreV1().Namespaces().Create(f.Ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		t.Fatal(errors.Wrap(err, fmt.Sprintf("failed to create namespace with name %v", name)))
	}

	namespaceFinalizerFn := func() error {
		return f.DeleteNamespace(name)
	}

	ctx.AddFinalizerFn(namespaceFinalizerFn)

	return name
}

func (f *Framework) DeleteNamespace(name string) error {
	return f.KubeClient.CoreV1().Namespaces().Delete(f.Ctx, name, metav1.DeleteOptions{})
}

func (f *Framework) AddLabelsToNamespace(name string, additionalLabels map[string]string) error {
	ns, err := f.KubeClient.CoreV1().Namespaces().Get(f.Ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}

	for k, v := range additionalLabels {
		ns.Labels[k] = v
	}

	_, err = f.KubeClient.CoreV1().Namespaces().Update(f.Ctx, ns, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (f *Framework) RemoveLabelsFromNamespace(name string, labels ...string) error {
	ns, err := f.KubeClient.CoreV1().Namespaces().Get(f.Ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if len(ns.Labels) == 0 {
		return nil
	}

	type patch struct {
		Op   string `json:"op"`
		Path string `json:"path"`
	}

	var patches []patch
	for _, l := range labels {
		patches = append(patches, patch{Op: "remove", Path: "/metadata/labels/" + l})
	}
	b, err := json.Marshal(patches)
	if err != nil {
		return err
	}

	_, err = f.KubeClient.CoreV1().Namespaces().Patch(f.Ctx, name, types.JSONPatchType, b, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}
