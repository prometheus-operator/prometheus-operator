package prometheus

import (
	apiUnversioned "k8s.io/client-go/1.4/pkg/api/unversioned"
	apiV1 "k8s.io/client-go/1.4/pkg/api/v1"
	apiExtensions "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/util/intstr"
)

// Object represents an Prometheus TPR API object.
type Object struct {
	apiUnversioned.TypeMeta `json:",inline"`
	apiV1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                    Spec `json:"spec"`
}

// Spec defines a Prometheus server.
type Spec struct {
	ServiceMonitors []SpecServiceMonitor `json:"serviceMonitors"`
}

// SpecServiceMonitor references a service monitor belonging to a Prometheus server.
type SpecServiceMonitor struct {
	Name string `json:"name"`
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
				},
			},
			Selector: map[string]string{
				"prometheus.coreos.com": name,
			},
		},
	}
	return svc
}

func makeReplicaSet(name string) *apiExtensions.ReplicaSet {
	replicas := int32(1)
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
						},
					},
				},
			},
		},
	}
	return rs
}
