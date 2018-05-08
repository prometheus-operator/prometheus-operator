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

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	RuleFilesKind = "RuleFile"
	RuleFileName  = "rulefiles"
)

type RuleFilesGetter interface {
	RuleFiles(namespace string) RuleFileInterface
}

var _ RuleFileInterface = &rulefiles{}

type RuleFileInterface interface {
	Create(*RuleFile) (*RuleFile, error)
	Get(name string, opts metav1.GetOptions) (*RuleFile, error)
	Update(*RuleFile) (*RuleFile, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error
}

type rulefiles struct {
	restClient rest.Interface
	client     dynamic.ResourceInterface
	crdKind    CrdKind
	ns         string
}

func newRuleFiles(r rest.Interface, c *dynamic.Client, crdKind CrdKind, namespace string) *rulefiles {
	return &rulefiles{
		restClient: r,
		client: c.Resource(
			&metav1.APIResource{
				Kind:       crdKind.Kind,
				Name:       crdKind.Plural,
				Namespaced: true,
			},
			namespace,
		),
		crdKind: crdKind,
		ns:      namespace,
	}
}

func (s *rulefiles) Create(o *RuleFile) (*RuleFile, error) {
	us, err := UnstructuredFromRuleFile(o)
	if err != nil {
		return nil, err
	}

	us, err = s.client.Create(us)
	if err != nil {
		return nil, err
	}

	return RuleFileFromUnstructured(us)
}

func (s *rulefiles) Get(name string, opts metav1.GetOptions) (*RuleFile, error) {
	obj, err := s.client.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return RuleFileFromUnstructured(obj)
}

func (s *rulefiles) Update(o *RuleFile) (*RuleFile, error) {
	us, err := UnstructuredFromRuleFile(o)
	if err != nil {
		return nil, err
	}

	curs, err := s.Get(o.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get current version for update")
	}
	us.SetResourceVersion(curs.ObjectMeta.ResourceVersion)

	us, err = s.client.Update(us)
	if err != nil {
		return nil, err
	}

	return RuleFileFromUnstructured(us)
}

func (s *rulefiles) Delete(name string, options *metav1.DeleteOptions) error {
	return s.client.Delete(name, options)
}

func (s *rulefiles) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := s.restClient.Get().
		Namespace(s.ns).
		Resource(s.crdKind.Plural)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var sm RuleFileList
	return &sm, json.Unmarshal(b, &sm)
}

func (s *rulefiles) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := s.restClient.Get().
		Prefix("watch").
		Namespace(s.ns).
		Resource(s.crdKind.Plural).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&ruleFileDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

func (s *rulefiles) DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error {
	return s.client.DeleteCollection(dopts, lopts)
}

// RuleFileFromUnstructured unmarshals a RuleFile object from dynamic client's unstructured
func RuleFileFromUnstructured(r *unstructured.Unstructured) (*RuleFile, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var s RuleFile
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	s.TypeMeta.Kind = RuleFilesKind
	s.TypeMeta.APIVersion = Group + "/" + Version
	return &s, nil
}

// UnstructuredFromRuleFile marshals a RuleFile object into dynamic client's unstructured
func UnstructuredFromRuleFile(s *RuleFile) (*unstructured.Unstructured, error) {
	s.TypeMeta.Kind = RuleFilesKind
	s.TypeMeta.APIVersion = Group + "/" + Version
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var r unstructured.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
}

type ruleFileDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *ruleFileDecoder) Close() {
	d.close()
}

func (d *ruleFileDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object RuleFile
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
