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
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespace(kubeClient kubernetes.Interface, name string) (*v1.Namespace, error) {
	namespace, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create namespace with name %v", name))
	}
	return namespace, nil
}

func (ctx *TestCtx) CreateNamespace(t *testing.T, kubeClient kubernetes.Interface) string {
	name := ctx.GetObjID()
	if _, err := CreateNamespace(kubeClient, name); err != nil {
		t.Fatal(err)
	}

	namespaceFinalizerFn := func() error {
		return DeleteNamespace(kubeClient, name)
	}

	ctx.AddFinalizerFn(namespaceFinalizerFn)

	return name
}

func DeleteNamespace(kubeClient kubernetes.Interface, name string) error {
	return kubeClient.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func AddLabelsToNamespace(kubeClient kubernetes.Interface, name string, additionalLabels map[string]string) error {
	ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}

	for k, v := range additionalLabels {
		ns.Labels[k] = v
	}

	_, err = kubeClient.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
