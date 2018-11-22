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

package v1

import (
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	AlertmanagersKind = "Alertmanager"
	AlertmanagerName  = "alertmanagers"
)

type AlertmanagersGetter interface {
	Alertmanagers(namespace string) AlertmanagerInterface
}

var _ AlertmanagerInterface = &alertmanagers{}

type AlertmanagerInterface interface {
	Create(*Alertmanager) (*Alertmanager, error)
	Get(name string, opts metav1.GetOptions) (*Alertmanager, error)
	Update(*Alertmanager) (*Alertmanager, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error
}

type alertmanagers struct {
	restClient rest.Interface
	crdKind    CrdKind
	ns         string
	timeout    time.Duration
}

func newAlertmanagers(r rest.Interface, crdKind CrdKind, namespace string, timeout time.Duration) *alertmanagers {
	return &alertmanagers{
		restClient: r,
		crdKind:    crdKind,
		ns:         namespace,
		timeout:    timeout,
	}
}

func (a *alertmanagers) Create(o *Alertmanager) (*Alertmanager, error) {
	result := &Alertmanager{}
	o.TypeMeta.Kind = AlertmanagersKind
	o.TypeMeta.APIVersion = Group + "/" + Version

	err := a.restClient.Post().
		Namespace(a.ns).
		Resource(a.crdKind.Plural).
		Body(o).
		Timeout(a.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *alertmanagers) Get(name string, opts metav1.GetOptions) (*Alertmanager, error) {
	result := &Alertmanager{}

	err := a.restClient.Get().
		Namespace(a.ns).
		Resource(a.crdKind.Plural).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(a.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *alertmanagers) Update(o *Alertmanager) (*Alertmanager, error) {
	result := &Alertmanager{}

	err := a.restClient.Put().
		Namespace(a.ns).
		Resource(a.crdKind.Plural).
		Body(o).
		Timeout(a.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *alertmanagers) Delete(name string, options *metav1.DeleteOptions) error {
	return a.restClient.Delete().
		Namespace(a.ns).
		Resource(a.crdKind.Plural).
		Name(name).
		Body(options).
		Timeout(a.timeout).
		Do().
		Error()
}

func (a *alertmanagers) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := a.restClient.Get().
		Namespace(a.ns).
		Resource(a.crdKind.Plural).
		Timeout(a.timeout)

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
		Resource(a.crdKind.Plural).
		Timeout(a.timeout).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&alertmanagerDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil

}

func (a *alertmanagers) DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error {
	timeout := a.timeout

	if lopts.TimeoutSeconds != nil {
		timeout = time.Duration(*lopts.TimeoutSeconds) * time.Second
	}

	return a.restClient.Delete().
		Namespace(a.ns).
		Resource(a.crdKind.Plural).
		VersionedParams(&lopts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(dopts).
		Do().
		Error()
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
