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

package common

import (
	"fmt"
	"log/slog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
)

// GetResourceFromKey returns a copy of the resource object identified by key.
// If the object is not found, it returns a nil pointer.
func GetResourceFromKey(key string, infs *informers.ForResource, logger *slog.Logger, resource string) (runtime.Object, error) {
	obj, err := infs.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("%s not found", "key", resource, key)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve %s from informer: %w", resource, err)
	}

	return obj, nil
}
