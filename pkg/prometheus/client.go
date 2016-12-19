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

package prometheus

import (
	"encoding/json"
	"time"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/schema"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/coreos/prometheus-operator/pkg/spec"
)

const resyncPeriod = 5 * time.Minute

type MonitoringClient struct {
	*rest.RESTClient
}

func (c *MonitoringClient) Prometheuses(namespace string) PrometheusInterface {
	return newPrometheuses(c, namespace)
}

func NewForConfig(c *rest.Config) (*MonitoringClient, error) {
	client, err := NewPrometheusRESTClient(*c)
	if err != nil {
		return nil, err
	}

	return &MonitoringClient{client}, nil
}

func NewPrometheusRESTClient(c rest.Config) (*rest.RESTClient, error) {
	c.APIPath = "/apis"
	c.GroupVersion = &schema.GroupVersion{
		Group:   "monitoring.coreos.com",
		Version: "v1alpha1",
	}
	// TODO(fabxc): is this even used with our custom list/watch functions?
	c.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
	return rest.RESTClientFor(&c)
}

type PrometheusInterface interface {
	Get(name string) (*spec.Prometheus, error)
}

type prometheuses struct {
	client *MonitoringClient
	ns     string
}

func newPrometheuses(c *MonitoringClient, namespace string) *prometheuses {
	return &prometheuses{
		client: c,
		ns:     namespace,
	}
}

func (p *prometheuses) Get(name string) (result *spec.Prometheus, err error) {
	result = &spec.Prometheus{}
	req := p.client.Get().
		Namespace(p.ns).
		Resource("prometheuses").
		Name(name)
	body, err := req.DoRaw()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &result)
	return
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
		Object spec.Prometheus
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}

// NewPrometheusListWatch returns a new ListWatch on the Prometheus resource.
func NewPrometheusListWatch(client *rest.RESTClient) *cache.ListWatch {
	return &cache.ListWatch{
		ListFunc: cache.ListFunc(func(options v1.ListOptions) (runtime.Object, error) {
			req := client.Get().
				Namespace(v1.NamespaceAll).
				Resource("prometheuses").
				// VersionedParams(&options, v1.ParameterCodec)
				FieldsSelectorParam(nil)

			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}
			var p spec.PrometheusList
			return &p, json.Unmarshal(b, &p)
		}),
		WatchFunc: cache.WatchFunc(func(options v1.ListOptions) (watch.Interface, error) {
			r, err := client.Get().
				Prefix("watch").
				Namespace(v1.NamespaceAll).
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
		}),
	}
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
		Object spec.ServiceMonitor
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}

// NewServiceMonitorListWatch returns a new ListWatch on the ServiceMonitor resource.
func NewServiceMonitorListWatch(client *rest.RESTClient) *cache.ListWatch {
	return &cache.ListWatch{
		ListFunc: cache.ListFunc(func(options v1.ListOptions) (runtime.Object, error) {
			req := client.Get().
				Namespace(v1.NamespaceAll).
				Resource("servicemonitors").
				// VersionedParams(&options, v1.ParameterCodec)
				FieldsSelectorParam(nil)

			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}
			var sm spec.ServiceMonitorList
			return &sm, json.Unmarshal(b, &sm)
		}),
		WatchFunc: cache.WatchFunc(func(options v1.ListOptions) (watch.Interface, error) {
			r, err := client.Get().
				Prefix("watch").
				Namespace(v1.NamespaceAll).
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
		}),
	}
}
