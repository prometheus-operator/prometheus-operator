package prometheus

import (
	"k8s.io/client-go/1.4/kubernetes"
	apierrors "k8s.io/client-go/1.4/pkg/api/errors"
)

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

	svcClient := p.kclient.Core().Services(p.Namespace)
	if _, err := svcClient.Create(makeService(p.Name)); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	rsClient := p.kclient.ExtensionsClient.ReplicaSets(p.Namespace)
	if _, err := rsClient.Create(makeReplicaSet(p.Name)); err != nil && !apierrors.IsAlreadyExists(err) {
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
