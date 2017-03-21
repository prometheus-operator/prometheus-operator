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

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
)

const (
	TPRServiceMonitorsKind = "ServiceMonitor"
	TPRServiceMonitorName  = "servicemonitors"
)

type ServiceMonitorsGetter interface {
	ServiceMonitors(namespace string) *dynamic.ResourceClient
}

type ServiceMonitorInterface interface {
	Create(*ServiceMonitor) (*ServiceMonitor, error)
	Get(name string) (*ServiceMonitor, error)
	Update(*ServiceMonitor) (*ServiceMonitor, error)
	Delete(name string, options *v1.DeleteOptions) error
	List(opts api.ListOptions) (runtime.Object, error)
	Watch(opts api.ListOptions) (watch.Interface, error)
}

type servicemonitors struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	ns         string
}

func newServiceMonitors(r rest.Interface, c *dynamic.Client, namespace string) *servicemonitors {
	return &servicemonitors{
		r,
		c.Resource(
			&unversioned.APIResource{
				Kind:       TPRServiceMonitorsKind,
				Name:       TPRServiceMonitorName,
				Namespaced: true,
			},
			namespace,
		),
		namespace,
	}
}

func (s *servicemonitors) Create(o *ServiceMonitor) (*ServiceMonitor, error) {
	us, err := UnstructuredFromServiceMonitor(o)
	if err != nil {
		return nil, err
	}

	us, err = s.client.Create(us)
	if err != nil {
		return nil, err
	}

	return ServiceMonitorFromUnstructured(us)
}

func (s *servicemonitors) Get(name string) (*ServiceMonitor, error) {
	obj, err := s.client.Get(name)
	if err != nil {
		return nil, err
	}
	return ServiceMonitorFromUnstructured(obj)
}

func (s *servicemonitors) Update(o *ServiceMonitor) (*ServiceMonitor, error) {
	us, err := UnstructuredFromServiceMonitor(o)
	if err != nil {
		return nil, err
	}

	us, err = s.client.Update(us)
	if err != nil {
		return nil, err
	}

	return ServiceMonitorFromUnstructured(us)
}

func (s *servicemonitors) Delete(name string, options *v1.DeleteOptions) error {
	return s.client.Delete(name, options)
}

func (s *servicemonitors) List(opts api.ListOptions) (runtime.Object, error) {
	req := s.restClient.Get().
		Namespace(s.ns).
		Resource("servicemonitors").
		// VersionedParams(&options, v1.ParameterCodec)
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var sm ServiceMonitorList
	return &sm, json.Unmarshal(b, &sm)
}

func (s *servicemonitors) Watch(opts api.ListOptions) (watch.Interface, error) {
	r, err := s.restClient.Get().
		Prefix("watch").
		Namespace(s.ns).
		Resource("servicemonitors").
		// VersionedParams(&options, v1.ParameterCodec).
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&serviceMonitorDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

// ServiceMonitorFromUnstructured unmarshals a ServiceMonitor object from dynamic client's unstructured
func ServiceMonitorFromUnstructured(r *runtime.Unstructured) (*ServiceMonitor, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var s ServiceMonitor
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	s.TypeMeta.Kind = TPRServiceMonitorsKind
	s.TypeMeta.APIVersion = TPRGroup + "/" + TPRVersion
	return &s, nil
}

// UnstructuredFromServiceMonitor marshals a ServiceMonitor object into dynamic client's unstructured
func UnstructuredFromServiceMonitor(s *ServiceMonitor) (*runtime.Unstructured, error) {
	s.TypeMeta.Kind = TPRServiceMonitorsKind
	s.TypeMeta.APIVersion = TPRGroup + "/" + TPRVersion
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var r runtime.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
}

type serviceMonitorDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *serviceMonitorDecoder) Close() {
	d.close()
}

func (d *serviceMonitorDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object ServiceMonitor
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
