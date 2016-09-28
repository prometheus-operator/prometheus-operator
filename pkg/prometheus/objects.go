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
	Service  apiV1.ServiceSpec `json:"service"`
	Monitors []MonitorRefSpec  `json:"monitors"`
	// Alerting AlertingSpec      `json:"alerting"`
}

type MonitorRefSpec struct {
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
	Endpoints      []Endpoint                   `json:"endpoints"`
	Selector       apiUnversioned.LabelSelector `json:"selector"`
	ScrapeInterval string                       `json:"scrapeInterval"`
	// Rules          []apiV1.ConfigMapVolumeSource `json:"rules"`
}

type Endpoint struct {
	Port intstr.IntOrString `json:"port"`
	Path string             `json:"path"`
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

func makeService(name string) *apiV1.Service {
	svc := &apiV1.Service{
		ObjectMeta: apiV1.ObjectMeta{
			Name: name,
		},
		Spec: apiV1.ServiceSpec{
			Ports: []apiV1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
					Protocol:   apiV1.ProtocolTCP,
					NodePort:   30900,
				},
			},
			Selector: map[string]string{
				"prometheus.coreos.com": name,
			},
			Type: apiV1.ServiceTypeNodePort,
		},
	}
	return svc
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
							Image: "quay.io/prometheus/prometheus:latest",
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
