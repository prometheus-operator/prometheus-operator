// Copyright 2024 The prometheus-operator Authors
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
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Nodes returns the list of nodes in the cluster.
func (f *Framework) Nodes(ctx context.Context) ([]v1.Node, error) {
	var (
		loopErr error
		nodes   *v1.NodeList
	)

	err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute*1, true, func(_ context.Context) (bool, error) {
		nodes, loopErr = f.KubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if loopErr != nil {
			return false, nil
		}

		if len(nodes.Items) < 1 {
			loopErr = errors.New("no nodes returned")
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v: %v", err, loopErr)
	}

	return nodes.Items, nil
}
