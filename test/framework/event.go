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

package framework

import (
	"context"
	"fmt"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WriteEvents writes the Kubernetes events for the given namespace.
// If the namespace is empty, all events are written.
func (f *Framework) WriteEvents(ctx context.Context, w io.Writer, ns string) error {
	events, err := f.KubeClient.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, e := range events.Items {
		fmt.Fprintf(w, "timestamp='%v' namespace=%q reason=%q message=%q\n", e.FirstTimestamp, e.Namespace, e.Reason, e.Message)
	}

	return nil
}
