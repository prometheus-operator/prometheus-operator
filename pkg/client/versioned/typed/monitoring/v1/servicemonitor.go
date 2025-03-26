// Copyright The prometheus-operator Authors
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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	context "context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	applyconfigurationmonitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	scheme "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// ServiceMonitorsGetter has a method to return a ServiceMonitorInterface.
// A group's client should implement this interface.
type ServiceMonitorsGetter interface {
	ServiceMonitors(namespace string) ServiceMonitorInterface
}

// ServiceMonitorInterface has methods to work with ServiceMonitor resources.
type ServiceMonitorInterface interface {
	Create(ctx context.Context, serviceMonitor *monitoringv1.ServiceMonitor, opts metav1.CreateOptions) (*monitoringv1.ServiceMonitor, error)
	Update(ctx context.Context, serviceMonitor *monitoringv1.ServiceMonitor, opts metav1.UpdateOptions) (*monitoringv1.ServiceMonitor, error)
	UpdateStatus(ctx context.Context, serviceMonitor *monitoringv1.ServiceMonitor, opts metav1.UpdateOptions) (*monitoringv1.ServiceMonitor, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*monitoringv1.ServiceMonitor, error)
	List(ctx context.Context, opts metav1.ListOptions) (*monitoringv1.ServiceMonitorList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *monitoringv1.ServiceMonitor, err error)
	Apply(ctx context.Context, serviceMonitor *applyconfigurationmonitoringv1.ServiceMonitorApplyConfiguration, opts metav1.ApplyOptions) (*monitoringv1.ServiceMonitor, error)
	ServiceMonitorExpansion
}

// serviceMonitors implements ServiceMonitorInterface
type serviceMonitors struct {
	*gentype.ClientWithListAndApply[*monitoringv1.ServiceMonitor, *monitoringv1.ServiceMonitorList, *applyconfigurationmonitoringv1.ServiceMonitorApplyConfiguration]
}

// newServiceMonitors returns a ServiceMonitors
func newServiceMonitors(c *MonitoringV1Client, namespace string) *serviceMonitors {
	return &serviceMonitors{
		gentype.NewClientWithListAndApply[*monitoringv1.ServiceMonitor, *monitoringv1.ServiceMonitorList, *applyconfigurationmonitoringv1.ServiceMonitorApplyConfiguration](
			"servicemonitors",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *monitoringv1.ServiceMonitor { return &monitoringv1.ServiceMonitor{} },
			func() *monitoringv1.ServiceMonitorList { return &monitoringv1.ServiceMonitorList{} },
		),
	}
}
