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
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

// GetCRD gets a custom resource definition from the apiserver.
func (f *Framework) GetCRD(name string) (*v1.CustomResourceDefinition, error) {
	crd, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get CRD with name %v", name)
	}
	return crd, nil
}

// ListCRDs gets a list of custom resource definitions from the apiserver.
func (f *Framework) ListCRDs() (*v1.CustomResourceDefinitionList, error) {
	crds, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list CRDs")
	}
	return crds, nil
}

// CreateCRD creates a custom resource definition on the apiserver.
func (f *Framework) CreateCRD(crd *v1.CustomResourceDefinition) error {
	_, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crd.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "getting CRD: %s", crd.Spec.Names.Kind)
	}

	if apierrors.IsNotFound(err) {
		_, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "create CRD: %s", crd.Spec.Names.Kind)
		}
	}
	return nil
}

// MakeCRD creates a CustomResourceDefinition object from yaml manifest.
func (f *Framework) MakeCRD(pathToYaml string) (*v1.CustomResourceDefinition, error) {
	manifest, err := ioutil.ReadFile(pathToYaml)
	if err != nil {
		return nil, errors.Wrapf(err, "read CRD asset file: %s", pathToYaml)
	}

	crd := v1.CustomResourceDefinition{}
	err = yaml.Unmarshal(manifest, &crd)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal CRD asset file: %s", pathToYaml)
	}

	return &crd, nil
}

// WaitForCRDReady waits for a Custom Resource Definition to be available for use.
func WaitForCRDReady(listFunc func(opts metav1.ListOptions) (runtime.Object, error)) error {
	err := wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := listFunc(metav1.ListOptions{})
		if err != nil {
			if se, ok := err.(*apierrors.StatusError); ok {
				if se.Status().Code == http.StatusNotFound {
					return false, nil
				}
			}
			return false, errors.Wrap(err, "failed to list CRD")
		}
		return true, nil
	})

	return errors.Wrap(err, "timed out waiting for Custom Resource")
}

// CreateCRDAndWaitUntilReady creates a Custom Resource Definition from yaml
// manifest on the apiserver and wait until it is available for use.
func (f *Framework) CreateCRDAndWaitUntilReady(crdName string, listFunc func(opts metav1.ListOptions) (runtime.Object, error)) error {
	crdName = strings.ToLower(crdName)
	group := monitoring.GroupName
	assetPath := "../../example/prometheus-operator-crd/" + group + "_" + crdName + ".yaml"

	crd, err := f.MakeCRD(assetPath)
	if err != nil {
		return errors.Wrapf(err, "create CRD: %s from manifest: %s", crdName, assetPath)
	}

	crd.ObjectMeta.Name = crd.Spec.Names.Plural + "." + group
	crd.Spec.Group = group

	err = f.CreateCRD(crd)
	if err != nil {
		return errors.Wrapf(err, "create CRD %s on the apiserver", crdName)
	}

	err = WaitForCRDReady(listFunc)
	if err != nil {
		return errors.Wrapf(err, "%s CRD not ready", crdName)
	}

	return nil
}
