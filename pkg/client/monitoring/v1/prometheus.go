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
	PrometheusesKind = "Prometheus"
	PrometheusName   = "prometheuses"
)

type PrometheusesGetter interface {
	Prometheuses(namespace string) PrometheusInterface
}

var _ PrometheusInterface = &prometheuses{}

type PrometheusInterface interface {
	Create(*Prometheus) (*Prometheus, error)
	Get(name string, opts metav1.GetOptions) (*Prometheus, error)
	Update(*Prometheus) (*Prometheus, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error
}

type prometheuses struct {
	restClient rest.Interface
	crdKind    CrdKind
	ns         string
	timeout    time.Duration
}

func newPrometheuses(r rest.Interface, crdKind CrdKind, namespace string, timeout time.Duration) *prometheuses {
	return &prometheuses{
		restClient: r,
		crdKind:    crdKind,
		ns:         namespace,
		timeout:    timeout,
	}
}

func (p *prometheuses) Create(o *Prometheus) (*Prometheus, error) {
	result := &Prometheus{}

	err := p.restClient.Post().
		Namespace(p.ns).
		Resource(p.crdKind.Plural).
		Body(o).
		Timeout(p.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *prometheuses) Get(name string, opts metav1.GetOptions) (*Prometheus, error) {
	result := &Prometheus{}

	err := p.restClient.Get().
		Namespace(p.ns).
		Resource(p.crdKind.Plural).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(p.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *prometheuses) Update(o *Prometheus) (*Prometheus, error) {
	result := &Prometheus{}

	err := p.restClient.Put().
		Namespace(p.ns).
		Resource(p.crdKind.Plural).
		Body(o).
		Timeout(p.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *prometheuses) Delete(name string, options *metav1.DeleteOptions) error {
	return p.restClient.Delete().
		Namespace(p.ns).
		Resource(p.crdKind.Plural).
		Name(name).
		Body(options).
		Timeout(p.timeout).
		Do().
		Error()
}

func (p *prometheuses) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := p.restClient.Get().
		Namespace(p.ns).
		Resource(p.crdKind.Plural).
		Timeout(p.timeout)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var prom PrometheusList
	return &prom, json.Unmarshal(b, &prom)
}

func (p *prometheuses) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := p.restClient.Get().
		Prefix("watch").
		Namespace(p.ns).
		Resource(p.crdKind.Plural).
		Timeout(p.timeout).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&prometheusDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

func (p *prometheuses) DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error {
	timeout := p.timeout

	if lopts.TimeoutSeconds != nil {
		timeout = time.Duration(*lopts.TimeoutSeconds) * time.Second
	}

	return p.restClient.Delete().
		Namespace(p.ns).
		Resource(p.crdKind.Plural).
		VersionedParams(&lopts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(dopts).
		Do().
		Error()
}

type prometheusDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *prometheusDecoder) Close() {
	d.close()
}

func (d *prometheusDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object Prometheus
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
