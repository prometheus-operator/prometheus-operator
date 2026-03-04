// Copyright 2026 The prometheus-operator Authors
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

package k8s

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// PodRunningAndReady returns whether a pod is running and each container has
// passed it's ready state.
func PodRunningAndReady(pod v1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case v1.PodFailed, v1.PodSucceeded:
		return false, fmt.Errorf("pod completed with phase %s", pod.Status.Phase)
	case v1.PodRunning:
		for _, cond := range pod.Status.Conditions {
			if cond.Type != v1.PodReady {
				continue
			}
			return cond.Status == v1.ConditionTrue, nil
		}
		return false, fmt.Errorf("pod ready condition not found")
	}
	return false, nil
}

// UpdateDNSConfig updates the DNS configuration in a Pod spec.
func UpdateDNSConfig(podSpec *v1.PodSpec, config *monitoringv1.PodDNSConfig) {
	if config == nil {
		return
	}

	dnsConfig := v1.PodDNSConfig{
		Nameservers: config.Nameservers,
		Searches:    config.Searches,
	}

	for _, opt := range config.Options {
		dnsConfig.Options = append(dnsConfig.Options, v1.PodDNSConfigOption{
			Name:  opt.Name,
			Value: opt.Value,
		})
	}

	podSpec.DNSConfig = &dnsConfig
}

// UpdateDNSPolicy updates the DNS policy in a Pod spec.
func UpdateDNSPolicy(podSpec *v1.PodSpec, dnsPolicy *monitoringv1.DNSPolicy) {
	if dnsPolicy == nil {
		return
	}

	podSpec.DNSPolicy = v1.DNSPolicy(*dnsPolicy)
}
