// Copyright 2016 The prometheus-operator Authors
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

package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	TPRAlertmanagersKind = "Alertmanager"
	TPRAlertmanagerName  = "alertmanagers"
)

type AlertmanagersGetter interface {
	Alertmanagers(namespace string) AlertmanagerInterface
}

type AlertmanagerInterface interface {
	Create(*Alertmanager) (*Alertmanager, error)
	Get(name string) (*Alertmanager, error)
	Update(*Alertmanager) (*Alertmanager, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type alertmanagers struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	ns         string
}

func newAlertmanagers(r rest.Interface, c *dynamic.Client, namespace string) *alertmanagers {
	return &alertmanagers{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       TPRAlertmanagersKind,
				Name:       TPRAlertmanagerName,
				Namespaced: true,
			},
			namespace,
		),
		namespace,
	}
}

func (a *alertmanagers) Create(o *Alertmanager) (*Alertmanager, error) {
	ua, err := UnstructuredFromAlertmanager(o)
	if err != nil {
		return nil, err
	}

	ua, err = a.client.Create(ua)
	if err != nil {
		return nil, err
	}

	return AlertmanagerFromUnstructured(ua)
}

func (a *alertmanagers) Get(name string) (*Alertmanager, error) {
	obj, err := a.client.Get(name)
	if err != nil {
		return nil, err
	}
	return AlertmanagerFromUnstructured(obj)
}

func (a *alertmanagers) Update(o *Alertmanager) (*Alertmanager, error) {
	ua, err := UnstructuredFromAlertmanager(o)
	if err != nil {
		return nil, err
	}

	ua, err = a.client.Update(ua)
	if err != nil {
		return nil, err
	}

	return AlertmanagerFromUnstructured(ua)
}

func (a *alertmanagers) Delete(name string, options *metav1.DeleteOptions) error {
	return a.client.Delete(name, options)
}

func (a *alertmanagers) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := a.restClient.Get().
		Namespace(a.ns).
		Resource("alertmanagers").
		// VersionedParams(&options, api.ParameterCodec)
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var p AlertmanagerList
	return &p, json.Unmarshal(b, &p)
}

func (a *alertmanagers) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := a.restClient.Get().
		Prefix("watch").
		Namespace(a.ns).
		Resource("alertmanagers").
		// VersionedParams(&options, api.ParameterCodec).
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&alertmanagerDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

// AlertmanagerFromUnstructured unmarshals an Alertmanager object from dynamic client's unstructured
func AlertmanagerFromUnstructured(r *unstructured.Unstructured) (*Alertmanager, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var a Alertmanager
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	a.TypeMeta.Kind = TPRAlertmanagersKind
	a.TypeMeta.APIVersion = TPRGroup + "/" + TPRVersion
	return &a, nil
}

// UnstructuredFromAlertmanager marshals an Alertmanager object into dynamic client's unstructured
func UnstructuredFromAlertmanager(a *Alertmanager) (*unstructured.Unstructured, error) {
	a.TypeMeta.Kind = TPRAlertmanagersKind
	a.TypeMeta.APIVersion = TPRGroup + "/" + TPRVersion
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	var r unstructured.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
}

type alertmanagerDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *alertmanagerDecoder) Close() {
	d.close()
}

func (d *alertmanagerDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object Alertmanager
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
