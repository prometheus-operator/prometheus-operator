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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"time"
)

func (f *Framework) CreateServiceAndWaitUntilReady(service *v1.Service) error {
	if _, err := f.KubeClient.CoreV1().Services(f.Namespace.Name).Create(service); err != nil {
		return err
	}

	if err := f.WaitForServiceReady(service.Name); err != nil {
		return err
	}
	return nil
}

func (f *Framework) WaitForServiceReady(serviceName string) error {
	err := f.Poll(time.Minute*5, time.Second, func() (bool, error) {
		endpoints, err := f.KubeClient.CoreV1().Endpoints(f.Namespace.Name).Get(serviceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(endpoints.Subsets) != 0 && len(endpoints.Subsets[0].Addresses) > 0 {
			return true, nil
		}
		return false, nil
	})
	return err
}

func (f *Framework) DeleteService(serviceName string) error {
	if err := f.KubeClient.CoreV1().Services(f.Namespace.Name).Delete(serviceName, nil); err != nil {
		return err
	}
	return nil
}
