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
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	prometheusCRDName = "prometheus"
)

// GetCRD gets a custom resource definition from the apiserver
func (f *Framework) GetCRD(name string) (*v1beta1.CustomResourceDefinition, error) {
	crd, err := f.APIServerClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to get CRD with name %v", name))
	}
	return crd, nil
}

// ListCRDs gets a list of custom resource definitions from the apiserver
func (f *Framework) ListCRDs() (*v1beta1.CustomResourceDefinitionList, error) {
	crds, err := f.APIServerClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to list CRDs"))
	}
	return crds, nil
}
