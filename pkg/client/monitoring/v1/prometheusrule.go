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
	PrometheusRuleKind = "PrometheusRule"
	PrometheusRuleName = "prometheusrules"
)

type PrometheusRulesGetter interface {
	PrometheusRules(namespace string) PrometheusRuleInterface
}

var _ PrometheusRuleInterface = &prometheusrules{}

type PrometheusRuleInterface interface {
	Create(*PrometheusRule) (*PrometheusRule, error)
	Get(name string, opts metav1.GetOptions) (*PrometheusRule, error)
	Update(*PrometheusRule) (*PrometheusRule, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error
}

type prometheusrules struct {
	restClient rest.Interface
	crdKind    CrdKind
	ns         string
	timeout    time.Duration
}

func newPrometheusRules(r rest.Interface, crdKind CrdKind, namespace string, timeout time.Duration) *prometheusrules {
	return &prometheusrules{
		restClient: r,
		crdKind:    crdKind,
		ns:         namespace,
		timeout:    timeout,
	}
}

func (s *prometheusrules) Create(o *PrometheusRule) (*PrometheusRule, error) {
	result := &PrometheusRule{}

	err := s.restClient.Post().
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		Body(o).
		Timeout(s.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *prometheusrules) Get(name string, opts metav1.GetOptions) (*PrometheusRule, error) {
	result := &PrometheusRule{}

	err := s.restClient.Get().
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(s.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *prometheusrules) Update(o *PrometheusRule) (*PrometheusRule, error) {
	result := &PrometheusRule{}

	err := s.restClient.Put().
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		Body(o).
		Timeout(s.timeout).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *prometheusrules) Delete(name string, options *metav1.DeleteOptions) error {
	return s.restClient.Delete().
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		Name(name).
		Body(options).
		Timeout(s.timeout).
		Do().
		Error()
}

func (s *prometheusrules) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := s.restClient.Get().
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		Timeout(s.timeout)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var sm PrometheusRuleList
	return &sm, json.Unmarshal(b, &sm)
}

func (s *prometheusrules) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := s.restClient.Get().
		Prefix("watch").
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		Timeout(s.timeout).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&prometheusRuleDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

func (s *prometheusrules) DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error {
	timeout := s.timeout

	if lopts.TimeoutSeconds != nil {
		timeout = time.Duration(*lopts.TimeoutSeconds) * time.Second
	}

	return s.restClient.Delete().
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		VersionedParams(&lopts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(dopts).
		Do().
		Error()
}

type prometheusRuleDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *prometheusRuleDecoder) Close() {
	d.close()
}

func (d *prometheusRuleDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object PrometheusRule
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
