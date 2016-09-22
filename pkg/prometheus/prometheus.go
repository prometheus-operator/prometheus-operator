package prometheus

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	apierrors "k8s.io/client-go/1.4/pkg/api/errors"
	apiExtensions "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/client-go/1.4/pkg/watch"
)

// Prometheus manages the life-cycle of a single Prometheus server
// in the cluster.
type Prometheus struct {
	*Object

	kclient *kubernetes.Clientset
	logger  log.Logger
	actions chan func() error
}

type event struct{}

// New returns a new Prometheus server manager for a newly created Prometheus.
func New(l log.Logger, kc *kubernetes.Clientset, o *Object) (*Prometheus, error) {
	p := &Prometheus{
		Object:  o,
		kclient: kc,
		logger:  l,
		actions: make(chan func() error),
	}

	svcClient := p.kclient.Core().Services(p.Namespace)
	if _, err := svcClient.Create(makeService(p.Name)); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	rsClient := p.kclient.ExtensionsClient.ReplicaSets(p.Namespace)
	if _, err := rsClient.Create(makeReplicaSet(p.Name, 1)); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	go p.run()
	return p, nil
}

// Delete removes the Prometheus server deployment asynchronously.
func (p *Prometheus) Delete() {
	p.actions <- p.deleteReplicaSet
	p.actions <- p.deleteService
}

func (p *Prometheus) deleteReplicaSet() error {
	// Update the replica count to 0 and wait for all pods to be deleted.
	rs := p.kclient.ExtensionsClient.ReplicaSets(p.Namespace)
	replicaSet, err := rs.Update(makeReplicaSet(p.Name, 0))
	if err != nil {
		return err
	}

	// XXX: Selecting by ObjectMeta.Name gives an error. So use the label for now.
	selector, err := labels.Parse("prometheus.coreos.com=" + p.Name)
	if err != nil {
		return err
	}
	w, err := rs.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	if _, err := watch.Until(100*time.Second, w, func(e watch.Event) (bool, error) {
		rs, ok := e.Object.(*apiExtensions.ReplicaSet)
		if !ok {
			fmt.Println(e.Object, e.Type)
			return false, errors.New("not a replica set")
		}
		// Check if the replica set is scaled down and all replicas are gone.
		if rs.Status.ObservedGeneration >= replicaSet.Status.ObservedGeneration && rs.Status.Replicas == *rs.Spec.Replicas {
			return true, nil
		}

		switch e.Type {
		// Deleted before we could validate it was scaled down correctly.
		case watch.Deleted:
			return true, errors.New("replica set deleted")
		case watch.Error:
			return false, errors.New("watch error")
		}
		return false, nil
	}); err != nil {
		return err
	}

	// Replica set scaled down, we can delete it.
	if err := rs.Delete(p.Name, nil); err != nil {
		return err
	}
	return nil
}

func (p *Prometheus) deleteService() error {
	svc := p.kclient.Core().Services(p.Namespace)
	if err := svc.Delete(p.Name, nil); err != nil {
		return err
	}
	return nil
}

func (p *Prometheus) run() {
	for f := range p.actions {
		if err := f(); err != nil {
			p.logger.Log("msg", "action failed", "err", err)
		}
	}
}
