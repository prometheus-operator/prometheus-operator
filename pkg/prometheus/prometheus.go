package prometheus

import (
	"k8s.io/client-go/1.4/kubernetes"
	core "k8s.io/client-go/1.4/kubernetes/typed/core/v1"
	extensions "k8s.io/client-go/1.4/kubernetes/typed/extensions/v1beta1"
	apierrors "k8s.io/client-go/1.4/pkg/api/errors"
	unversionedapi "k8s.io/client-go/1.4/pkg/api/unversioned"
	apiv1 "k8s.io/client-go/1.4/pkg/api/v1"
	extensionsapi "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/util/intstr"
)

// Object represents an Prometheus TPR API object.
type Object struct {
	unversionedapi.TypeMeta `json:",inline"`
	apiv1.ObjectMeta        `json:"metadata,omitempty"`
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

// Prometheus manages the life-cycle of a single Prometheus server
// in the cluster.
type Prometheus struct {
	*Object

	kclient *kubernetes.Clientset
}

// New returns a new Prometheus server manager for a newly created Prometheus.
func New(kc *kubernetes.Clientset, o *Object) (*Prometheus, error) {
	p := &Prometheus{
		Object:  o,
		kclient: kc,
	}
	if err := createService(p.kclient.Core().Services(p.Namespace), p.Name); err != nil {
		return nil, err
	}
	if err := createReplicaSet(p.kclient.ExtensionsClient.ReplicaSets(p.Namespace), p.Name); err != nil {
		return nil, err
	}
	go p.run()
	return p, nil
}

// Delete romves the Prometheus server deployment.
func (p *Prometheus) Delete() error {
	rs := p.kclient.ExtensionsClient.ReplicaSets(p.Namespace)
	if err := rs.Delete(p.Name, nil); err != nil {
		return err
	}
	svc := p.kclient.Core().Services(p.Namespace)
	if err := svc.Delete(p.Name, nil); err != nil {
		return err
	}
	return nil
}

func (p *Prometheus) run() {
}

func createReplicaSet(client extensions.ReplicaSetInterface, name string) error {
	replicas := int32(1)
	rs := &extensionsapi.ReplicaSet{
		ObjectMeta: apiv1.ObjectMeta{
			Name: name,
		},
		Spec: extensionsapi.ReplicaSetSpec{
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					Labels: map[string]string{
						"prometheus.coreos.com": name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "prometheus",
							Image: "quay.io/prometheus/prometheus:latest",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 9090,
									Protocol:      apiv1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
	if _, err := client.Create(rs); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func createService(client core.ServiceInterface, name string) error {
	svc := &apiv1.Service{
		ObjectMeta: apiv1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
					Protocol:   apiv1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"prometheus.coreos.com": name,
			},
		},
	}
	if _, err := client.Create(svc); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
