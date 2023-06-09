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
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (f *Framework) createReplicationControllerViaYml(ctx context.Context, namespace string, filepath string) error {
	manifest, err := os.Open(filepath)
	if err != nil {
		return err
	}

	var rC v1.ReplicationController
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&rC)
	if err != nil {
		return err
	}

	_, err = f.KubeClient.CoreV1().ReplicationControllers(namespace).Create(ctx, &rC, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (f *Framework) deleteReplicationControllerViaYml(ctx context.Context, namespace string, filepath string) error {
	manifest, err := os.Open(filepath)
	if err != nil {
		return err
	}

	var rC v1.ReplicationController
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&rC)
	if err != nil {
		return err
	}

	if err := f.scaleDownReplicationController(ctx, namespace, rC); err != nil {
		return err
	}

	return f.KubeClient.CoreV1().ReplicationControllers(namespace).Delete(ctx, rC.Name, metav1.DeleteOptions{})
}

func (f *Framework) scaleDownReplicationController(ctx context.Context, namespace string, rC v1.ReplicationController) error {
	*rC.Spec.Replicas = 0
	rCAPI := f.KubeClient.CoreV1().ReplicationControllers(namespace)

	_, err := f.KubeClient.CoreV1().ReplicationControllers(namespace).Update(ctx, &rC, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(ctx, time.Second, time.Minute*5, false, func(ctx context.Context) (bool, error) {
		currentRC, err := rCAPI.Get(ctx, rC.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if currentRC.Status.Replicas == 0 {
			return true, nil
		}
		return false, nil
	})
}
