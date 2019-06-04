/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package watchmanager

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Manager is the interface for registering and unregistering
// objects referenced by pods in the underlying cache and
// extracting those from that cache if needed.
type Manager interface {
	// Get object by its namespace and name.
	GetObject(namespace, name string) (runtime.Object, error)

	// WARNING: Register/UnregisterPod functions should be efficient,
	// i.e. should not block on network operations.

	// RegisterSMon registers all objects referenced from a given ServiceMonitor.
	RegisterSMon(sMon *monitoringv1.ServiceMonitor)

	// UnregisterSMon unregisters objects referenced from a given ServiceMonitor that are not
	// used by any other registered ServiceMonitor.
	UnregisterSMon(sMon *monitoringv1.ServiceMonitor)
}
