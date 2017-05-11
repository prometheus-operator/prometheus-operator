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
	TPRPrometheusesKind = "Prometheus"
	TPRPrometheusName   = "prometheuses"
)

type PrometheusesGetter interface {
	Prometheuses(namespace string) PrometheusInterface
}

type PrometheusInterface interface {
	Create(*Prometheus) (*Prometheus, error)
	Get(name string) (*Prometheus, error)
	Update(*Prometheus) (*Prometheus, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type prometheuses struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	ns         string
}

func newPrometheuses(r rest.Interface, c *dynamic.Client, namespace string) *prometheuses {
	return &prometheuses{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       TPRPrometheusesKind,
				Name:       TPRPrometheusName,
				Namespaced: true,
			},
			namespace,
		),
		namespace,
	}
}

func (p *prometheuses) Create(o *Prometheus) (*Prometheus, error) {
	up, err := UnstructuredFromPrometheus(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Create(up)
	if err != nil {
		return nil, err
	}

	return PrometheusFromUnstructured(up)
}

func (p *prometheuses) Get(name string) (*Prometheus, error) {
	obj, err := p.client.Get(name)
	if err != nil {
		return nil, err
	}
	return PrometheusFromUnstructured(obj)
}

func (p *prometheuses) Update(o *Prometheus) (*Prometheus, error) {
	up, err := UnstructuredFromPrometheus(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Update(up)
	if err != nil {
		return nil, err
	}

	return PrometheusFromUnstructured(up)
}

func (p *prometheuses) Delete(name string, options *metav1.DeleteOptions) error {
	return p.client.Delete(name, options)
}

func (p *prometheuses) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := p.restClient.Get().
		Namespace(p.ns).
		Resource("prometheuses").
		// VersionedParams(&options, v1.ParameterCodec)
		FieldsSelectorParam(nil)

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
		Resource("prometheuses").
		// VersionedParams(&options, v1.ParameterCodec).
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&prometheusDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

// PrometheusFromUnstructured unmarshals a Prometheus object from dynamic client's unstructured
func PrometheusFromUnstructured(r *unstructured.Unstructured) (*Prometheus, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var p Prometheus
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	p.TypeMeta.Kind = TPRPrometheusesKind
	p.TypeMeta.APIVersion = TPRGroup + "/" + TPRVersion
	return &p, nil
}

// UnstructuredFromPrometheus marshals a Prometheus object into dynamic client's unstructured
func UnstructuredFromPrometheus(p *Prometheus) (*unstructured.Unstructured, error) {
	p.TypeMeta.Kind = TPRPrometheusesKind
	p.TypeMeta.APIVersion = TPRGroup + "/" + TPRVersion
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var r unstructured.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
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
