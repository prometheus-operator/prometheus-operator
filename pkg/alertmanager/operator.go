// Copyright 2016 The prometheus-operator Authors
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

package alertmanager

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/analytics"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	"github.com/coreos/prometheus-operator/pkg/spec"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	apierrors "k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/apps/v1alpha1"
	extensionsobj "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/labels"
	utilruntime "k8s.io/client-go/1.5/pkg/util/runtime"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/cache"
)

const (
	TPRGroup   = "monitoring.coreos.com"
	TPRVersion = "v1alpha1"

	TPRAlertmanagersKind = "alertmanagers"

	tprAlertmanager = "alertmanager." + TPRGroup
)

// Operator manages lify cycle of Alertmanager deployments and
// monitoring configurations.
type Operator struct {
	kclient *kubernetes.Clientset
	pclient *rest.RESTClient
	logger  log.Logger

	alrtInf cache.SharedIndexInformer
	psetInf cache.SharedIndexInformer

	queue *queue

	host string
}

// New creates a new controller.
func New(c prometheus.Config, logger log.Logger) (*Operator, error) {
	cfg, err := newClusterConfig(c.Host, c.TLSInsecure, &c.TLSConfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	promclient, err := newAlertmanagerRESTClient(*cfg)
	if err != nil {
		return nil, err
	}
	return &Operator{
		kclient: client,
		pclient: promclient,
		logger:  logger,
		queue:   newQueue(200),
		host:    cfg.Host,
	}, nil
}

// Run the controller.
func (c *Operator) Run(stopc <-chan struct{}) error {
	defer c.queue.close()
	go c.worker()

	v, err := c.kclient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("communicating with server failed: %s", err)
	}
	c.logger.Log("msg", "connection established", "cluster-version", v)

	if err := c.createTPRs(); err != nil {
		return err
	}

	c.alrtInf = cache.NewSharedIndexInformer(
		NewAlertmanagerListWatch(c.pclient),
		&spec.Alertmanager{}, resyncPeriod, cache.Indexers{},
	)
	c.psetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Apps().GetRESTClient(), "petsets", api.NamespaceAll, nil),
		&v1alpha1.PetSet{}, resyncPeriod, cache.Indexers{},
	)

	c.alrtInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(p interface{}) {
			c.logger.Log("msg", "enqueueAlertmanager", "trigger", "alertmanager add")
			analytics.AlertmanagerCreated()
			c.enqueueAlertmanager(p)
		},
		DeleteFunc: func(p interface{}) {
			c.logger.Log("msg", "enqueueAlertmanager", "trigger", "alertmanager del")
			analytics.AlertmanagerDeleted()
			c.enqueueAlertmanager(p)
		},
		UpdateFunc: func(_, p interface{}) {
			c.logger.Log("msg", "enqueueAlertmanager", "trigger", "alertmanager update")
			c.enqueueAlertmanager(p)
		},
	})
	c.psetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(d interface{}) {
			c.logger.Log("msg", "addPetSet", "trigger", "petset add")
			c.addPetSet(d)
		},
		DeleteFunc: func(d interface{}) {
			c.logger.Log("msg", "deletePetSet", "trigger", "petset delete")
			c.deletePetSet(d)
		},
		UpdateFunc: func(old, cur interface{}) {
			c.logger.Log("msg", "updatePetSet", "trigger", "petset update")
			c.updatePetSet(old, cur)
		},
	})

	go c.alrtInf.Run(stopc)
	go c.psetInf.Run(stopc)

	for !c.alrtInf.HasSynced() || !c.psetInf.HasSynced() {
		time.Sleep(100 * time.Millisecond)
	}

	<-stopc
	return nil
}

type queue struct {
	ch chan *spec.Alertmanager
}

func newQueue(size int) *queue {
	return &queue{ch: make(chan *spec.Alertmanager, size)}
}

func (q *queue) add(p *spec.Alertmanager) { q.ch <- p }
func (q *queue) close()                   { close(q.ch) }

func (q *queue) pop() (*spec.Alertmanager, bool) {
	p, ok := <-q.ch
	return p, ok
}

var keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc

func (c *Operator) enqueueAlertmanager(p interface{}) {
	c.queue.add(p.(*spec.Alertmanager))
}

func (c *Operator) enqueueAll() {
	cache.ListAll(c.alrtInf.GetStore(), labels.Everything(), func(o interface{}) {
		c.enqueueAlertmanager(o.(*spec.Alertmanager))
	})
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *Operator) worker() {
	for {
		p, ok := c.queue.pop()
		if !ok {
			return
		}
		if err := c.reconcile(p); err != nil {
			utilruntime.HandleError(fmt.Errorf("reconciliation failed: %s", err))
		}
	}
}

func (c *Operator) alertmanagerForPetSet(p *v1alpha1.PetSet) *spec.Alertmanager {
	key, err := keyFunc(p)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return nil
	}
	// Namespace/Name are one-to-one so the key will find the respective Alertmanager resource.
	a, exists, err := c.alrtInf.GetStore().GetByKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get Alertmanager resource: %s", err))
		return nil
	}
	if !exists {
		return nil
	}
	return a.(*spec.Alertmanager)
}

func (c *Operator) deletePetSet(o interface{}) {
	p := o.(*v1alpha1.PetSet)
	// Wake up Alertmanager resource the deployment belongs to.
	if a := c.alertmanagerForPetSet(p); a != nil {
		c.enqueueAlertmanager(a)
	}
}

func (c *Operator) addPetSet(o interface{}) {
	p := o.(*v1alpha1.PetSet)
	// Wake up Alertmanager resource the deployment belongs to.
	if a := c.alertmanagerForPetSet(p); a != nil {
		c.enqueueAlertmanager(a)
	}
}

func (c *Operator) updatePetSet(oldo, curo interface{}) {
	old := oldo.(*v1alpha1.PetSet)
	cur := curo.(*v1alpha1.PetSet)

	c.logger.Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the deployment without changes in-between.
	// Also breaks loops created by updating the resource ourselves.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	// Wake up Alertmanager resource the deployment belongs to.
	if a := c.alertmanagerForPetSet(cur); a != nil {
		c.enqueueAlertmanager(a)
	}
}

func (c *Operator) reconcile(p *spec.Alertmanager) error {
	key, err := keyFunc(p)
	if err != nil {
		return err
	}
	c.logger.Log("msg", "reconcile alertmanager", "key", key)

	_, exists, err := c.alrtInf.GetStore().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		// TODO(fabxc): we want to do server side deletion due to the variety of
		// resources we create.
		// Doing so just based on the deletion event is not reliable, so
		// we have to garbage collect the controller-created resources in some other way.
		//
		// Let's rely on the index key matching that of the created configmap and replica
		// set for now. This does not work if we delete Alertmanager resources as the
		// controller is not running – that could be solved via garbage collection later.
		return c.deleteAlertmanager(p)
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(p.Namespace)
	if _, err := svcClient.Create(makePetSetService(p)); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create petset service: %s", err)
	}

	psetClient := c.kclient.Apps().PetSets(p.Namespace)
	psetQ := &v1alpha1.PetSet{}
	psetQ.Namespace = p.Namespace
	psetQ.Name = p.Name
	obj, exists, err := c.psetInf.GetStore().Get(psetQ)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := psetClient.Create(makePetSet(p.Namespace, p, nil)); err != nil {
			return fmt.Errorf("create petset: %s", err)
		}
		return nil
	}
	if _, err := psetClient.Update(makePetSet(p.Namespace, p, obj.(*v1alpha1.PetSet))); err != nil {
		return err
	}

	return c.syncVersion(p)
}

func podRunningAndReady(pod v1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case v1.PodFailed, v1.PodSucceeded:
		return false, fmt.Errorf("pod completed")
	case v1.PodRunning:
		for _, cond := range pod.Status.Conditions {
			if cond.Type != v1.PodReady {
				continue
			}
			return cond.Status == v1.ConditionTrue, nil
		}
		return false, fmt.Errorf("pod ready conditation not found")
	}
	return false, nil
}

// syncVersion ensures that all running pods for a Alertmanager have the required version.
// It kills pods with the wrong version one-after-one and lets the PetSet controller
// create new pods.
//
// TODO(fabxc): remove this once the PetSet controller learns how to do rolling updates.
func (c *Operator) syncVersion(p *spec.Alertmanager) error {
	selector, err := labels.Parse("app=alertmanager,alertmanager=" + p.Name)
	if err != nil {
		return err
	}
	podClient := c.kclient.Core().Pods(p.Namespace)

Outer:
	for {
		pods, err := podClient.List(api.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			return nil
		}
		for _, cp := range pods.Items {
			ready, err := podRunningAndReady(cp)
			if err != nil {
				return err
			}
			if !ready {
				time.Sleep(200 * time.Millisecond)
				continue Outer
			}
		}
		var pod *v1.Pod
		for _, cp := range pods.Items {
			if !strings.HasSuffix(cp.Spec.Containers[0].Image, p.Spec.Version) {
				pod = &cp
				break
			}
		}
		if pod == nil {
			return nil
		}
		if err := podClient.Delete(pod.Name, nil); err != nil {
			return err
		}
	}
}

func (c *Operator) deleteAlertmanager(p *spec.Alertmanager) error {
	// Update the replica count to 0 and wait for all pods to be deleted.
	psetClient := c.kclient.Apps().PetSets(p.Namespace)

	key, err := keyFunc(p)
	if err != nil {
		return err
	}
	oldPsetO, _, err := c.psetInf.GetStore().GetByKey(key)
	if err != nil {
		return err
	}
	oldPset := oldPsetO.(*v1alpha1.PetSet)
	zero := int32(0)
	oldPset.Spec.Replicas = &zero

	if _, err := psetClient.Update(oldPset); err != nil {
		return err
	}

	// XXX: Selecting by ObjectMeta.Name gives an error. So use the label for now.
	selector, err := labels.Parse("app=alertmanager,alertmanager=" + p.Name)
	if err != nil {
		return err
	}
	podClient := c.kclient.Core().Pods(p.Namespace)

	// TODO(fabxc): temporary solution until PetSet status provides necessary info to know
	// whether scale-down completed.
	for {
		pods, err := podClient.List(api.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Deployment scaled down, we can delete it.
	if err := psetClient.Delete(p.Name, nil); err != nil {
		return err
	}

	// if err := c.kclient.Core().Services(p.Namespace).Delete(fmt.Sprintf("%s-petset", p.Name), nil); err != nil {
	// 	return err
	// }

	// Delete the auto-generate configuration.
	// TODO(fabxc): add an ownerRef at creation so we don't delete config maps
	// manually created for Alertmanager servers with no ServiceMonitor selectors.
	cm := c.kclient.Core().ConfigMaps(p.Namespace)
	if err := cm.Delete(p.Name, nil); err != nil {
		return err
	}
	if err := cm.Delete(fmt.Sprintf("%s-rules", p.Name), nil); err != nil {
		return err
	}
	return nil
}

func (c *Operator) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprAlertmanager,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: TPRVersion},
			},
			Description: "Managed Alertmanager cluster",
		},
	}
	tprClient := c.kclient.Extensions().ThirdPartyResources()

	for _, tpr := range tprs {
		if _, err := tprClient.Create(tpr); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		c.logger.Log("msg", "TPR created", "tpr", tpr.Name)
	}

	// We have to wait for the TPRs to be ready. Otherwise the initial watch may fail.
	return k8sutil.WaitForTPRReady(c.kclient.CoreClient.GetRESTClient(), TPRGroup, TPRVersion, TPRAlertmanagersKind)
}

func newClusterConfig(host string, tlsInsecure bool, tlsConfig *rest.TLSClientConfig) (*rest.Config, error) {
	var cfg *rest.Config
	var err error

	if len(host) == 0 {
		if cfg, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		cfg = &rest.Config{
			Host: host,
		}
		hostURL, err := url.Parse(host)
		if err != nil {
			return nil, fmt.Errorf("error parsing host url %s : %v", host, err)
		}
		if hostURL.Scheme == "https" {
			cfg.TLSClientConfig = *tlsConfig
			cfg.Insecure = tlsInsecure
		}
	}
	cfg.QPS = 100
	cfg.Burst = 100

	return cfg, nil
}
