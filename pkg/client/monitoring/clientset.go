// Copyright 2017 The prometheus-operator Authors
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

package monitoring

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

var _ Interface = &Clientset{}

type Interface interface {
	MonitoringV1alpha1() v1alpha1.MonitoringV1alpha1Interface
	MonitoringV1() v1.MonitoringV1Interface
}

type Clientset struct {
	*v1alpha1.MonitoringV1alpha1Client
	*v1.MonitoringV1Client
}

func (c *Clientset) MonitoringV1alpha1() v1alpha1.MonitoringV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.MonitoringV1alpha1Client
}

func (c *Clientset) MonitoringV1() v1.MonitoringV1Interface {
	if c == nil {
		return nil
	}
	return c.MonitoringV1Client
}

func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error

	cs.MonitoringV1alpha1Client, err = v1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.MonitoringV1Client, err = v1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
