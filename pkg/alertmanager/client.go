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

package alertmanager

import (
	"encoding/json"
	"time"

	"github.com/coreos/prometheus-operator/pkg/spec"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const resyncPeriod = 5 * time.Minute

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
		Object spec.Alertmanager
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}

// NewAlertmanagerListWatch returns a new ListWatch on the Alertmanager resource.
func NewAlertmanagerListWatch(client *rest.RESTClient) *cache.ListWatch {
	return &cache.ListWatch{
		ListFunc: cache.ListFunc(func(options v1.ListOptions) (runtime.Object, error) {
			req := client.Get().
				Namespace(api.NamespaceAll).
				Resource("alertmanagers").
				// VersionedParams(&options, api.ParameterCodec)
				FieldsSelectorParam(nil)

			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}
			var p spec.AlertmanagerList
			return &p, json.Unmarshal(b, &p)
		}),
		WatchFunc: cache.WatchFunc(func(options v1.ListOptions) (watch.Interface, error) {
			r, err := client.Get().
				Prefix("watch").
				Namespace(api.NamespaceAll).
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
		}),
	}
}
