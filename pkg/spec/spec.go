package spec

import (
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/util/intstr"
)

// Prometheus defines a Prometheus deployment.
type Prometheus struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                 PrometheusSpec `json:"spec"`
}

// PrometheusList is a list of Prometheuses.
type PrometheusList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*Prometheus `json:"items"`
}

// PrometheusSpec holds specification parameters of a Prometheus deployment.
type PrometheusSpec struct {
	ServiceMonitors    []ServiceMonitorSelection `json:"serviceMonitors"`
	EvaluationInterval string                    `json:"evaluationInterval"`
	// Retention       string                     `json:"retention"`
	// Replicas        int                        `json:"replicas"`
	// Resources       apiV1.ResourceRequirements `json:"resources"`
	// Alerting        AlertingSpec               `json:"alerting"`
	// Remote          RemoteSpec                 `json:"remote"`
	// Persistence...
	// Sharding...
}

// ServiceMonitorSelection selects service monitors by their labels.
type ServiceMonitorSelection struct {
	Selector unversioned.LabelSelector `json:"selector"`
}

// ServiceMonitor defines monitoring for a set of services.
type ServiceMonitor struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                 ServiceMonitorSpec `json:"spec"`
}

// ServiceMonitorSpec contains specification parameters for a ServiceMonitor.
type ServiceMonitorSpec struct {
	Endpoints []Endpoint                `json:"endpoints"`
	Selector  unversioned.LabelSelector `json:"selector"`
	// Rules          []apiV1.ConfigMapVolumeSource `json:"rules"`
}

// Endpoint defines a scrapeable endpoint serving Prometheus metrics.
type Endpoint struct {
	Port       string             `json:"port"`
	TargetPort intstr.IntOrString `json:"targetPort"`
	Path       string             `json:"path"`
	Scheme     string             `json:"scheme"`
	Interval   string             `json:"interval"`
}

// ServiceMonitorList is a list of ServiceMonitors.
type ServiceMonitorList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []ServiceMonitor `json:"items"`
}
