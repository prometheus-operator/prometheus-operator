package prometheus

import (
	"bytes"
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
	stopc   chan struct{}
}

// New returns a new Prometheus server manager for a newly created Prometheus.
func New(l log.Logger, kc *kubernetes.Clientset, o *Object) (*Prometheus, error) {
	p := &Prometheus{
		kclient: kc,
		logger:  l,
		actions: make(chan func() error),
		stopc:   make(chan struct{}),
	}

	if err := p.update(o)(); err != nil {
		return nil, err
	}

	go p.run()
	return p, nil
}

// Update applies changes to the object.
func (p *Prometheus) Update(o *Object) {
	p.actions <- p.update(o)
}

func (p *Prometheus) update(o *Object) func() error {
	return func() error {
		p.Object = o

		var (
			svcClient = p.kclient.Core().Services(p.Namespace)
			rsClient  = p.kclient.ExtensionsClient.ReplicaSets(p.Namespace)
			cmClient  = p.kclient.Core().ConfigMaps(o.Namespace)
		)

		// XXX: for some reason creating an existing services does not return an
		// AlreadyExists error but complains about immutable attributes.
		if _, err := svcClient.Get(p.Name); apierrors.IsNotFound(err) {
			if _, err := svcClient.Create(makeService(p.Name)); err != nil {
				return fmt.Errorf("create service: %s", err)
			}
		} else if err != nil {
			return err
		}

		// Update config map based on the most recent configuration.
		var buf bytes.Buffer
		if err := configTmpl.Execute(&buf, nil); err != nil {
			return err
		}

		cm := makeConfigMap(p.Name, map[string]string{
			"prometheus.yaml": buf.String(),
		})
		if _, err := cmClient.Get(p.Name); apierrors.IsNotFound(err) {
			if _, err := cmClient.Create(cm); err != nil {
				return err
			}
		} else if err == nil {
			if _, err := cmClient.Update(cm); err != nil {
				return err
			}
		} else {
			return err
		}

		if _, err := rsClient.Create(makeReplicaSet(p.Name, 1)); err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create replica set: %s", err)
		}

		return nil
	}
}

// Delete removes the Prometheus server deployment asynchronously.
func (p *Prometheus) Delete() {
	p.logger.Log("exec", "delete", "prometheus", p.Name)
	p.actions <- p.deleteReplicaSet
	p.actions <- p.deleteService
	p.actions <- p.deleteConfigMap
	close(p.stopc)
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

func (p *Prometheus) deleteConfigMap() error {
	cm := p.kclient.Core().ConfigMaps(p.Namespace)
	if err := cm.Delete(p.Name, nil); err != nil {
		return err
	}
	return nil
}

func (p *Prometheus) run() {
	for {
		select {
		case <-p.stopc:
			return
		default:
		}
		select {
		case f := <-p.actions:
			if err := f(); err != nil {
				p.logger.Log("msg", "action failed", "err", err)
			}
		case <-p.stopc:
			return
		}
	}
}
