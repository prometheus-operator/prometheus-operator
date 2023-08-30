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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
)

// GetCRD gets a custom resource definition from the apiserver.
func (f *Framework) GetCRD(ctx context.Context, name string) (*v1.CustomResourceDefinition, error) {
	crd, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get CRD with name %v", name)
	}
	return crd, nil
}

// ListCRDs gets a list of custom resource definitions from the apiserver.
func (f *Framework) ListCRDs(ctx context.Context) (*v1.CustomResourceDefinitionList, error) {
	crds, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list CRDs")
	}
	return crds, nil
}

// CreateOrUpdateCRD creates a custom resource definition on the apiserver.
func (f *Framework) CreateOrUpdateCRD(ctx context.Context, crd *v1.CustomResourceDefinition) error {
	c, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "getting CRD: %s", crd.Spec.Names.Kind)
	}

	if apierrors.IsNotFound(err) {
		// CRD doesn't exists -> Create
		_, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "create CRD: %s", crd.Spec.Names.Kind)
		}
	} else {
		// must set this field from existing CRD to prevent update fail
		crd.ObjectMeta.ResourceVersion = c.ObjectMeta.ResourceVersion

		// CRD already exists -> Update
		_, err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "update CRD: %s", crd.Spec.Names.Kind)
		}
	}
	return nil
}

func (f *Framework) DeleteCRD(ctx context.Context, name string) error {
	err := f.APIServerClient.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to delete CRD with name %v", name)
	}

	return nil
}

// MakeCRD creates a CustomResourceDefinition object from yaml manifest.
func (f *Framework) MakeCRD(source string) (*v1.CustomResourceDefinition, error) {
	manifest, err := SourceToIOReader(source)
	if err != nil {
		return nil, errors.Wrapf(err, "get manifest from source: %s", source)
	}

	content, err := io.ReadAll(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "get manifest content")
	}

	crd := v1.CustomResourceDefinition{}
	err = yaml.Unmarshal(content, &crd)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal CRD asset file: %s", source)
	}

	return &crd, nil
}

// WaitForCRDReady waits for a Custom Resource Definition to be available for use.
func WaitForCRDReady(listFunc func(opts metav1.ListOptions) (runtime.Object, error)) error {
	err := wait.PollUntilContextTimeout(context.Background(), 3*time.Second, 10*time.Minute, false, func(ctx context.Context) (bool, error) {
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
func (f *Framework) CreateOrUpdateCRDAndWaitUntilReady(ctx context.Context, crdName string, listFunc func(opts metav1.ListOptions) (runtime.Object, error)) error {
	crdName = strings.ToLower(crdName)
	group := monitoring.GroupName
	assetPath := f.exampleDir + "/prometheus-operator-crd-full/" + group + "_" + crdName + ".yaml"

	crd, err := f.MakeCRD(assetPath)
	if err != nil {
		return errors.Wrapf(err, "create CRD: %s from manifest: %s", crdName, assetPath)
	}

	crd.ObjectMeta.Name = crd.Spec.Names.Plural + "." + group
	crd.Spec.Group = group

	err = f.CreateOrUpdateCRD(ctx, crd)
	if err != nil {
		return errors.Wrapf(err, "create CRD %s on the apiserver", crdName)
	}

	err = WaitForCRDReady(listFunc)
	if err != nil {
		return errors.Wrapf(err, "%s CRD not ready", crdName)
	}

	return nil
}
