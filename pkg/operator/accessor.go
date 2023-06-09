// Copyright 2023 The prometheus-operator Authors
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

package operator

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// Accessor can manipulate objects returned by informers and handlers.
type Accessor struct {
	logger log.Logger
}

func NewAccessor(l log.Logger) *Accessor {
	return &Accessor{
		logger: l,
	}
}

// MetaNamespaceKey returns a key from the object's metadata.
// For namespaced objects, the format is `<namespace>/<name>`, otherwise
// `name`.
func (a *Accessor) MetaNamespaceKey(obj interface{}) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		level.Error(a.logger).Log("msg", "failed to retrieve object's key", "err", err)
		return k, false
	}

	return k, true
}

// ObjectMetadata returns the object's metadata and bool indicating if the
// conversion succeeded
func (a *Accessor) ObjectMetadata(obj interface{}) (metav1.Object, bool) {
	ts, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		obj = ts.Obj
	}

	o, err := meta.Accessor(obj)
	if err != nil {
		level.Error(a.logger).Log("msg", "get object failed", "err", err)
		return nil, false
	}
	return o, true
}
