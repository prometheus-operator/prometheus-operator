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
	"github.com/coreos/prometheus-operator/pkg/queue"
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
	"k8s.io/kubernetes/pkg/api/meta"
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

	queue *queue.Queue

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
		queue:   queue.New(),
		host:    cfg.Host,
	}, nil
}

// Run the controller.
func (c *Operator) Run(stopc <-chan struct{}) error {
	defer c.queue.ShutDown()
	go c.worker()

	errChan := make(chan error)
	go func() {
		v, err := c.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- fmt.Errorf("communicating with server failed: %s", err)
			return
		}
		c.logger.Log("msg", "connection established", "cluster-version", v)

		if err := c.createTPRs(); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		c.logger.Log("msg", "TPR API endpoints ready")
	case <-stopc:
		return nil
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
		AddFunc:    c.handleAlertmanagerAdd,
		DeleteFunc: c.handleAlertmanagerDelete,
		UpdateFunc: c.handleAlertmanagerUpdate,
	})
	c.psetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePetSetAdd,
		DeleteFunc: c.handlePetSetDelete,
		UpdateFunc: c.handlePetSetUpdate,
	})

	go c.alrtInf.Run(stopc)
	go c.psetInf.Run(stopc)

	<-stopc
	return nil
}

func (c *Operator) keyFunc(obj interface{}) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.logger.Log("msg", "creating key failed", "err", err)
		return k, false
	}
	return k, true
}

func (c *Operator) getObject(obj interface{}) (meta.Object, bool) {
	ts, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		obj = ts.Obj
	}

	o, err := meta.Accessor(obj)
	if err != nil {
		c.logger.Log("msg", "get object failed", "err", err)
		return nil, false
	}
	return o, true
}

// enqueue adds a key to the queue. If obj is a key already it gets added directly.
// Otherwise, the key is extracted via keyFunc.
func (c *Operator) enqueue(obj interface{}) {
	if obj == nil {
		return
	}

	key, ok := obj.(string)
	if !ok {
		key, ok = c.keyFunc(obj)
		if !ok {
			return
		}
	}

	c.queue.Add(key)
}

// enqueueForNamespace enqueues all Prometheus object keys that belong to the given namespace.
func (c *Operator) enqueueForNamespace(ns string) {
	cache.ListAll(c.alrtInf.GetStore(), labels.Everything(), func(obj interface{}) {
		am := obj.(*spec.Alertmanager)
		if am.Namespace == ns {
			c.enqueue(am)
		}
	})
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *Operator) worker() {
	for {
		key, quit := c.queue.Get()
		if quit {
			return
		}
		if err := c.sync(key.(string)); err != nil {
			utilruntime.HandleError(fmt.Errorf("reconciliation failed, re-enqueueing: %s", err))
			// We only mark the item as done after waiting. In the meantime
			// other items can be processed but the same item won't be processed again.
			// This is a trivial form of rate-limiting that is sufficient for our throughput
			// and latency expectations.
			go func() {
				time.Sleep(3 * time.Second)
				c.queue.Done(key)
			}()
			continue
		}

		c.queue.Done(key)
	}
}

func (c *Operator) alertmanagerForPetSet(ps interface{}) *spec.Alertmanager {
	key, ok := c.keyFunc(ps)
	if !ok {
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

func (c *Operator) handleAlertmanagerAdd(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	analytics.AlertmanagerCreated()
	c.logger.Log("msg", "Alertmanager added", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleAlertmanagerDelete(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	analytics.AlertmanagerDeleted()
	c.logger.Log("msg", "Alertmanager deleted", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleAlertmanagerUpdate(old, cur interface{}) {
	key, ok := c.keyFunc(cur)
	if !ok {
		return
	}

	c.logger.Log("msg", "Alertmanager updated", "key", key)
	c.enqueue(key)
}

func (c *Operator) handlePetSetDelete(obj interface{}) {
	if a := c.alertmanagerForPetSet(obj); a != nil {
		c.enqueue(a)
	}
}

func (c *Operator) handlePetSetAdd(obj interface{}) {
	if a := c.alertmanagerForPetSet(obj); a != nil {
		c.enqueue(a)
	}
}

func (c *Operator) handlePetSetUpdate(oldo, curo interface{}) {
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
		c.enqueue(a)
	}
}

func (c *Operator) sync(key string) error {
	obj, exists, err := c.alrtInf.GetIndexer().GetByKey(key)
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
		return c.destroyAlertmanager(key)
	}

	am := obj.(*spec.Alertmanager)

	c.logger.Log("msg", "sync alertmanager", "key", key)

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(am.Namespace)
	if _, err := svcClient.Create(makePetSetService(am)); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create petset service: %s", err)
	}

	psetClient := c.kclient.Apps().PetSets(am.Namespace)
	// Ensure we have a PetSet running Alertmanager deployed.
	obj, exists, err = c.psetInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := psetClient.Create(makePetSet(am, nil)); err != nil {
			return fmt.Errorf("create petset: %s", err)
		}
		return nil
	}
	if _, err := psetClient.Update(makePetSet(am, obj.(*v1alpha1.PetSet))); err != nil {
		return err
	}

	return c.syncVersion(am)
}

func listOptions(name string) api.ListOptions {
	return api.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app":          "alertmanager",
			"alertmanager": name,
		}),
	}
}

// syncVersion ensures that all running pods for a Alertmanager have the required version.
// It kills pods with the wrong version one-after-one and lets the PetSet controller
// create new pods.
//
// TODO(fabxc): remove this once the PetSet controller learns how to do rolling updates.
func (c *Operator) syncVersion(am *spec.Alertmanager) error {
	podClient := c.kclient.Core().Pods(am.Namespace)

Outer:
	for {
		pods, err := podClient.List(listOptions(am.Name))
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			return nil
		}
		for _, cp := range pods.Items {
			ready, err := k8sutil.PodRunningAndReady(cp)
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
			if !strings.HasSuffix(cp.Spec.Containers[0].Image, am.Spec.Version) {
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

func (c *Operator) destroyAlertmanager(key string) error {
	obj, exists, err := c.psetInf.GetStore().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	pset := obj.(*v1alpha1.PetSet)
	*pset.Spec.Replicas = 0

	// Update the replica count to 0 and wait for all pods to be deleted.
	psetClient := c.kclient.Apps().PetSets(pset.Namespace)

	if _, err := psetClient.Update(pset); err != nil {
		return err
	}

	podClient := c.kclient.Core().Pods(pset.Namespace)

	// TODO(fabxc): temporary solution until PetSet status provides necessary info to know
	// whether scale-down completed.
	for {
		pods, err := podClient.List(listOptions(pset.Name))
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// PetSet scaled down, we can delete it.
	if err := psetClient.Delete(pset.Name, nil); err != nil {
		return err
	}

	// Delete the auto-generate configuration.
	// TODO(fabxc): add an ownerRef at creation so we don't delete config maps
	// manually created for Alertmanager servers with no ServiceMonitor selectors.
	cm := c.kclient.Core().ConfigMaps(pset.Namespace)
	if err := cm.Delete(pset.Name, nil); err != nil {
		return err
	}
	if err := cm.Delete(fmt.Sprintf("%s-rules", pset.Name), nil); err != nil {
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
