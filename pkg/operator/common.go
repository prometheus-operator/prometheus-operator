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
	"fmt"
	"log/slog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
)

// GetObjectFromKey retrieves an object from the informer cache using the provided key.
func GetObjectFromKey[T interface {
	DeepCopy() T
}](infs *informers.ForResource, key string, logger *slog.Logger) (T, error) {
	obj, err := infs.Get(key)
	var zero T
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Object not found", "key", key)
			return zero, nil
		}
		return zero, fmt.Errorf("failed to retrieve object from informer: %w", err)
	}

	typed, ok := any(obj).(T)
	if !ok {
		return zero, fmt.Errorf("unexpected type %T", obj)
	}

	return typed.DeepCopy(), nil
}
