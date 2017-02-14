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
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"time"
)

func createReplicationControllerViaYml(filepath string, f *Framework) error {
	manifest, err := os.Open(filepath)
	if err != nil {
		return err
	}

	var rC v1.ReplicationController
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&rC)
	if err != nil {
		return err
	}

	_, err = f.KubeClient.CoreV1().ReplicationControllers(f.Namespace.Name).Create(&rC)
	if err != nil {
		return err
	}

	return nil
}

func deleteReplicationControllerViaYml(filepath string, f *Framework) error {
	manifest, err := os.Open(filepath)
	if err != nil {
		return err
	}

	var rC v1.ReplicationController
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&rC)
	if err != nil {
		return err
	}

	if err := scaleDownReplicationController(f, rC); err != nil {
		return err
	}

	if err := f.KubeClient.CoreV1().ReplicationControllers(f.Namespace.Name).Delete(rC.Name, nil); err != nil {
		return err
	}

	return nil
}

func scaleDownReplicationController(f *Framework, rC v1.ReplicationController) error {
	*rC.Spec.Replicas = 0
	rCAPI := f.KubeClient.CoreV1().ReplicationControllers(f.Namespace.Name)

	_, err := f.KubeClient.CoreV1().ReplicationControllers(f.Namespace.Name).Update(&rC)
	if err != nil {
		return err
	}

	return f.Poll(time.Minute*5, time.Second, func() (bool, error) {
		currentRC, err := rCAPI.Get(rC.Name, apimetav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if currentRC.Status.Replicas == 0 {
			return true, nil
		}
		return false, nil
	})
}
