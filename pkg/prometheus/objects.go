package prometheus

import (
	apiUnversioned "k8s.io/client-go/1.4/pkg/api/unversioned"
	apiV1 "k8s.io/client-go/1.4/pkg/api/v1"
	apiExtensions "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/util/intstr"
)

// Object represents an Prometheus TPR API object.
type PrometheusObj struct {
	apiUnversioned.TypeMeta `json:",inline"`
	apiV1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                    PrometheusSpec `json:"spec"`
}

// Spec defines a Prometheus server.
type PrometheusSpec struct {
	ServiceMonitors []MonitorRefSpec `json:"serviceMonitors"`
	// Namespaces   []NamespaceRefSpec `json:"namespaces"`
	// Retention       string                     `json:"retention"`
	// Replicas        int                        `json:"replicas"`
	// Resources       apiV1.ResourceRequirements `json:"resources"`
	// Alerting        AlertingSpec               `json:"alerting"`
	// Remote          RemoteSpec                 `json:"remote"`
	// Persistence...
	// Sharding...
}

type MonitorRefSpec struct {
	Selector apiUnversioned.LabelSelector `json:"selector"`
}

type NamespaceRefSpec struct {
	Selector apiUnversioned.LabelSelector `json:"selector"`
}

// type AlertingSpec struct {
// 	Selector apiUnversioned.LabelSelector `json:"selector"`
// }

type ServiceMonitorObj struct {
	apiUnversioned.TypeMeta `json:",inline"`
	apiV1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                    ServiceMonitorSpec `json:"spec"`
}

type ServiceMonitorSpec struct {
	Endpoints []Endpoint                   `json:"endpoints"`
	Selector  apiUnversioned.LabelSelector `json:"selector"`
	// Rules          []apiV1.ConfigMapVolumeSource `json:"rules"`
}

type Endpoint struct {
	Port       string             `json:"port"`
	TargetPort intstr.IntOrString `json:"targetPort"`
	Path       string             `json:"path"`
	Scheme     string             `json:"scheme"`
	Interval   string             `json:"interval"`
}

type ServiceMonitorList struct {
	apiUnversioned.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	apiUnversioned.ListMeta `json:"metadata,omitempty"`
	// Items is a list of third party objects
	Items []ServiceMonitorObj `json:"items"`
}

func makeConfigMap(name string, data map[string]string) *apiV1.ConfigMap {
	cm := &apiV1.ConfigMap{
		ObjectMeta: apiV1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}
	return cm
}

func makeReplicaSet(name string, replicas int32) *apiExtensions.ReplicaSet {
	rs := &apiExtensions.ReplicaSet{
		ObjectMeta: apiV1.ObjectMeta{
			Name: name,
		},
		Spec: apiExtensions.ReplicaSetSpec{
			Replicas: &replicas,
			Template: apiV1.PodTemplateSpec{
				ObjectMeta: apiV1.ObjectMeta{
					Labels: map[string]string{
						"prometheus.coreos.com": name,
					},
				},
				Spec: apiV1.PodSpec{
					Containers: []apiV1.Container{
						{
							Name:  "prometheus",
							Image: "quay.io/fabxc/prometheus:v1.3.0-beta.0",
							Ports: []apiV1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 9090,
									Protocol:      apiV1.ProtocolTCP,
								},
							},
							Args: []string{
								"-storage.local.retention=12h",
								"-storage.local.memory-chunks=500000",
								"-config.file=/etc/prometheus/prometheus.yaml",
							},
							VolumeMounts: []apiV1.VolumeMount{
								{
									Name:      "config-volume",
									ReadOnly:  true,
									MountPath: "/etc/prometheus",
								},
							},
						}, {
							Name:  "reloader",
							Image: "jimmidyson/configmap-reload",
							Args: []string{
								"-webhook-url=http://localhost:9090/-/reload",
								"-volume-dir=/etc/prometheus/",
							},
							VolumeMounts: []apiV1.VolumeMount{
								{
									Name:      "config-volume",
									ReadOnly:  true,
									MountPath: "/etc/prometheus",
								},
							},
						},
					},
					Volumes: []apiV1.Volume{
						{
							Name: "config-volume",
							VolumeSource: apiV1.VolumeSource{
								ConfigMap: &apiV1.ConfigMapVolumeSource{
									LocalObjectReference: apiV1.LocalObjectReference{
										Name: name,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return rs
}
