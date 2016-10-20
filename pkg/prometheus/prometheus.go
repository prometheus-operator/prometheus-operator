package prometheus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	apierrors "k8s.io/client-go/1.4/pkg/api/errors"
	"k8s.io/client-go/1.4/pkg/api/unversioned"
	apiExtensions "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/client-go/1.4/pkg/watch"
)

// Prometheus manages the life-cycle of a single Prometheus server
// in the cluster.
type Prometheus struct {
	*PrometheusObj

	ctx    context.Context
	cancel func()

	host    string
	hclient *http.Client
	kclient *kubernetes.Clientset
	logger  log.Logger
	actions chan func() error
}

// New returns a new Prometheus server manager for a newly created Prometheus.
func New(ctx context.Context, l log.Logger, host string, kc *kubernetes.Clientset, o *PrometheusObj) (*Prometheus, error) {
	ctx, cancel := context.WithCancel(ctx)
	p := &Prometheus{
		ctx:     ctx,
		cancel:  cancel,
		kclient: kc,
		hclient: kc.CoreClient.Client,
		host:    host,
		logger:  l,
		actions: make(chan func() error),
	}

	if err := p.update(o)(); err != nil {
		return nil, err
	}
	if err := p.generateConfig(); err != nil {
		return nil, err
	}

	go p.run()
	go p.runWatchServiceMonitors()

	return p, nil
}

// Update applies changes to the object.
func (p *Prometheus) Update(o *PrometheusObj) {
	p.actions <- p.update(o)
}

func (p *Prometheus) update(o *PrometheusObj) func() error {
	return func() error {
		p.PrometheusObj = o

		var rsClient = p.kclient.ExtensionsClient.ReplicaSets(p.Namespace)

		if _, err := rsClient.Create(makeReplicaSet(p.Name, 1)); err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create replica set: %s", err)
		}

		return nil
	}
}

// Event represents an event in the cluster.
type Event struct {
	Type   watch.EventType
	Object ServiceMonitorObj
}

func (p *Prometheus) runWatchServiceMonitors() {
	watchVersion := "0"
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}
		if err := p.watchServiceMonitors(&watchVersion); err != nil {
			p.logger.Log("msg", "watching service monitors failed", "err", err)
			watchVersion = "0"
		}
	}
}

func (p *Prometheus) watchServiceMonitors(watchVersion *string) error {
	resp, err := p.hclient.Get(p.host + "/apis/prometheus.coreos.com/v1alpha1/namespaces/" + p.Namespace + "/servicemonitors?watch=true&resourceVersion=" + *watchVersion)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unxpected status code %s", resp.Status)
	}
	p.logger.Log("msg", "watching Prometheus resource", "version", watchVersion)

	for {
		dec := json.NewDecoder(resp.Body)
		var evt Event
		if err := dec.Decode(&evt); err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}

		if evt.Type == "ERROR" {
			return fmt.Errorf("received error event")
		}
		p.logger.Log("msg", "Prometheus event", "type", evt.Type, "obj", fmt.Sprintf("%v:%v", evt.Object.Namespace, evt.Object.Name))
		*watchVersion = evt.Object.ObjectMeta.ResourceVersion

		if err := p.generateConfig(); err != nil {
			return err
		}
	}
}

func (p *Prometheus) generateConfig() error {
	if len(p.Spec.ServiceMonitors) == 0 {
		return nil
	}
	// TODO(fabxc): deduplicate job names for double matching monitors.
	monitors := map[string]ServiceMonitorObj{}
	for _, m := range p.Spec.ServiceMonitors {
		lsel, err := unversioned.LabelSelectorAsSelector(&m.Selector)
		if err != nil {
			return err
		}

		sms, err := p.getServiceMonitors(lsel)
		if err != nil {
			return err
		}
		for _, m := range sms.Items {
			monitors[m.Namespace+"\xff"+m.Name] = m
		}
	}

	tplcfg := &TemplateConfig{ServiceMonitors: monitors}

	// Update config map based on the most recent configuration.
	var buf bytes.Buffer
	if err := configTmpl.Execute(&buf, tplcfg); err != nil {
		return err
	}

	cmClient := p.kclient.Core().ConfigMaps(p.Namespace)

	cm := makeConfigMap(p.Name, map[string]string{
		"prometheus.yaml": buf.String(),
	})
	_, err := cmClient.Get(p.Name)
	if apierrors.IsNotFound(err) {
		_, err = cmClient.Create(cm)
	} else if err == nil {
		_, err = cmClient.Update(cm)
	}
	return err
}

func (p *Prometheus) getServiceMonitors(labelSelector labels.Selector) (*ServiceMonitorList, error) {
	path := "/apis/prometheus.coreos.com/v1/alpha1/namespaces/" + p.Namespace + "/servicemonitors"
	if labelSelector != nil {
		path += "?labelSelector=" + labelSelector.String()
	}

	req, err := http.NewRequest("GET", p.host+path, nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
	defer cancel()

	resp, err := p.hclient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res ServiceMonitorList
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete removes the Prometheus server deployment asynchronously.
func (p *Prometheus) Delete() {
	p.actions <- p.deleteReplicaSet
	p.actions <- p.deleteConfigMap
	p.cancel()
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
		case <-p.ctx.Done():
			return
		default:
		}
		select {
		case f := <-p.actions:
			if err := f(); err != nil {
				p.logger.Log("msg", "action failed", "err", err)
			}
		case <-p.ctx.Done():
			return
		}
	}
}
