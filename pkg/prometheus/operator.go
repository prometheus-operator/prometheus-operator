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

package prometheus

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/analytics"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/queue"
	"github.com/coreos/prometheus-operator/pkg/spec"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	apierrors "k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
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

	TPRPrometheusesKind    = "prometheuses"
	TPRServiceMonitorsKind = "servicemonitors"

	tprServiceMonitor = "service-monitor." + TPRGroup
	tprPrometheus     = "prometheus." + TPRGroup
)

// Operator manages lify cycle of Prometheus deployments and
// monitoring configurations.
type Operator struct {
	kclient *kubernetes.Clientset
	pclient *rest.RESTClient
	logger  log.Logger

	promInf cache.SharedIndexInformer
	smonInf cache.SharedIndexInformer
	cmapInf cache.SharedIndexInformer
	psetInf cache.SharedIndexInformer
	epntInf cache.SharedIndexInformer

	queue *queue.Queue

	host string
}

// Config defines configuration parameters for the Operator.
type Config struct {
	Host        string
	TLSInsecure bool
	TLSConfig   rest.TLSClientConfig
}

// New creates a new controller.
func New(conf Config, logger log.Logger) (*Operator, error) {
	cfg, err := newClusterConfig(conf.Host, conf.TLSInsecure, &conf.TLSConfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	promclient, err := NewPrometheusRESTClient(*cfg)
	if err != nil {
		return nil, err
	}
	c := &Operator{
		kclient: client,
		pclient: promclient,
		logger:  logger,
		queue:   queue.New(),
		host:    cfg.Host,
	}

	c.promInf = cache.NewSharedIndexInformer(
		NewPrometheusListWatch(c.pclient),
		&spec.Prometheus{}, resyncPeriod, cache.Indexers{},
	)
	c.smonInf = cache.NewSharedIndexInformer(
		NewServiceMonitorListWatch(c.pclient),
		&spec.ServiceMonitor{}, resyncPeriod, cache.Indexers{},
	)
	c.cmapInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Core().GetRESTClient(), "configmaps", api.NamespaceAll, nil),
		&v1.ConfigMap{}, resyncPeriod, cache.Indexers{},
	)
	c.psetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Apps().GetRESTClient(), "petsets", api.NamespaceAll, nil),
		&v1alpha1.PetSet{}, resyncPeriod, cache.Indexers{},
	)

	c.promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addPrometheus,
		DeleteFunc: c.deletePrometheus,
		UpdateFunc: c.updatePrometheus,
	})
	c.smonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(_ interface{}) {
			c.logger.Log("msg", "enqueueAll", "trigger", "smon add")
			c.enqueueAll()
		},
		DeleteFunc: func(_ interface{}) {
			c.logger.Log("msg", "enqueueAll", "trigger", "smon del")
			c.enqueueAll()
		},
		UpdateFunc: func(_, _ interface{}) {
			c.logger.Log("msg", "enqueueAll", "trigger", "smon update")
			c.enqueueAll()
		},
	})
	c.cmapInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// TODO(fabxc): only enqueue Prometheus the ConfigMap belonged to.
		DeleteFunc: func(_ interface{}) {
			c.logger.Log("msg", "enqueueAll", "trigger", "cmap del")
			c.enqueueAll()
		},
	})
	c.psetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// TODO(fabxc): only enqueue Prometheus an affected deployment belonged to.
		AddFunc: func(d interface{}) {
			c.logger.Log("msg", "addDeployment", "trigger", "depl add")
			c.addPetSet(d)
		},
		DeleteFunc: func(d interface{}) {
			c.logger.Log("msg", "deleteDeployment", "trigger", "depl delete")
			c.deletePetSet(d)
		},
		UpdateFunc: func(old, cur interface{}) {
			c.logger.Log("msg", "updateDeployment", "trigger", "depl update")
			c.updatePetSet(old, cur)
		},
	})

	return c, nil
}

// Run the controller.
func (c *Operator) Run(stopc <-chan struct{}) error {
	defer c.queue.ShutDown()

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
	case <-stopc:
		return nil
	}

	go c.worker()

	go c.promInf.Run(stopc)
	go c.smonInf.Run(stopc)
	go c.cmapInf.Run(stopc)
	go c.psetInf.Run(stopc)

	<-stopc
	return nil
}

var keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc

func (c *Operator) addPrometheus(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return
	}

	analytics.PrometheusCreated()
	c.logger.Log("msg", "Prometheus added", "key", key)

	c.queue.Add(key)
}

func (c *Operator) deletePrometheus(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return
	}

	analytics.PrometheusDeleted()
	c.logger.Log("msg", "Prometheus deleted", "key", key)

	c.queue.Add(key)
}

func (c *Operator) updatePrometheus(old, cur interface{}) {
	// oldp := old.(*spec.Prometheus)
	// curp := cur.(*spec.Prometheus)

	key, err := keyFunc(cur)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return
	}

	c.logger.Log("msg", "Prometheus updated", "key", key)

	c.queue.Add(key)
}

func (c *Operator) enqueueAll() {
	cache.ListAll(c.promInf.GetStore(), labels.Everything(), func(obj interface{}) {
		key, err := keyFunc(obj)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
			return
		}
		c.queue.Add(key)
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

func (c *Operator) prometheusForDeployment(d *v1alpha1.PetSet) *spec.Prometheus {
	key, err := keyFunc(d)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return nil
	}
	// Namespace/Name are one-to-one so the key will find the respective Prometheus resource.
	p, exists, err := c.promInf.GetStore().GetByKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get Prometheus resource: %s", err))
		return nil
	}
	if !exists {
		return nil
	}
	return p.(*spec.Prometheus)
}

func (c *Operator) deletePetSet(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return
	}

	_, exists, err := c.promInf.GetIndexer().GetByKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get PetSet from index: %s", err))
		return
	}
	// No Prometheus setup this PetSet relates to.
	if !exists {
		return
	}

	c.queue.Add(key)
}

func (c *Operator) addPetSet(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return
	}

	d := obj.(*v1alpha1.PetSet)
	// Wake up Prometheus resource the deployment belongs to.
	if p := c.prometheusForDeployment(d); p != nil {
		c.queue.Add(key)
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

	key, err := keyFunc(curo)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("creating key: %s", err))
		return
	}

	// Wake up Prometheus resource the deployment belongs to.
	if p := c.prometheusForDeployment(cur); p != nil {
		c.queue.Add(key)
	}
}

func (c *Operator) sync(key string) error {
	c.logger.Log("msg", "reconcile prometheus", "key", key)

	obj, exists, err := c.promInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		// TODO(fabxc): we want to do server side deletion due to the variety of
		// resources we create.
		// Doing so just based on the deletion event is not reliable, so
		// we have to garbage collect the controller-created resources in some other way.
		//
		// Let's rely on the index key matching that of the created configmap and PetSet for now.
		// This does not work if we delete Prometheus resources as the
		// controller is not running – that could be solved via garbage collection later.
		return c.destroyPrometheus(key)
	}

	p := obj.(*spec.Prometheus)

	// If no service monitor selectors are configured, the user wants to manage
	// configuration himself.
	if p.Spec.ServiceMonitorSelector != nil {
		// We just always regenerate the configuration to be safe.
		if err := c.createConfig(p); err != nil {
			return err
		}
	}

	// Create ConfigMaps if they don't exist.
	cmClient := c.kclient.Core().ConfigMaps(p.Namespace)
	if _, err := cmClient.Create(makeEmptyConfig(p.Name)); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	if _, err := cmClient.Create(makeEmptyRules(p.Name)); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(p.Namespace)
	if _, err := svcClient.Create(makePetSetService(p)); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create petset service: %s", err)
	}

	psetClient := c.kclient.Apps().PetSets(p.Namespace)
	// Ensure we have a PetSet running Prometheus deployed.
	obj, exists, err = c.psetInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := psetClient.Create(makePetSet(*p, nil)); err != nil {
			return fmt.Errorf("create petset: %s", err)
		}
		return nil
	}
	if _, err := psetClient.Update(makePetSet(*p, obj.(*v1alpha1.PetSet))); err != nil {
		return err
	}

	return c.syncVersion(key, p)
}

func ListOptions(name string) api.ListOptions {
	return api.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app":        "prometheus",
			"prometheus": name,
		}),
	}
}

// syncVersion ensures that all running pods for a Prometheus have the required version.
// It kills pods with the wrong version one-after-one and lets the PetSet controller
// create new pods.
//
// TODO(fabxc): remove this once the PetSet controller learns how to do rolling updates.
func (c *Operator) syncVersion(key string, p *spec.Prometheus) error {
	podClient := c.kclient.Core().Pods(p.Namespace)

	pods, err := podClient.List(ListOptions(p.Name))
	if err != nil {
		return err
	}

	// If the PetSet is still busy scaling, don't interfere by killing pods.
	// We enqueue ourselves again to until the PetSet is ready.
	if len(pods.Items) != int(p.Spec.Replicas) {
		return fmt.Errorf("scaling in progress")
	}
	if len(pods.Items) == 0 {
		return nil
	}

	var oldPods []*v1.Pod
	allReady := true
	// Only proceed if all existing pods are running and ready.
	for _, pod := range pods.Items {
		ready, err := k8sutil.PodRunningAndReady(pod)
		if err != nil {
			c.logger.Log("msg", "cannot determine pod ready state", "err", err)
		}
		if ready {
			// TODO(fabxc): detect other fields of the pod template that are mutable.
			if !strings.HasSuffix(pod.Spec.Containers[0].Image, p.Spec.Version) {
				oldPods = append(oldPods, &pod)
			}
			continue
		}
		allReady = false
	}

	if len(oldPods) == 0 {
		return nil
	}
	if !allReady {
		return fmt.Errorf("waiting for pods to become ready")
	}

	// TODO(fabxc): delete oldest pod first.
	if err := podClient.Delete(oldPods[0].Name, nil); err != nil {
		return err
	}
	// If there are further pods that need updating, we enqueue ourselves again.
	if len(oldPods) > 1 {
		return fmt.Errorf("%d out-of-date pods remaining", len(oldPods)-1)
	}
	return nil
}

func (c *Operator) destroyPrometheus(key string) error {
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

	// TODO(fabxc): temprorary solution until PetSet status provides necessary info to know
	// whether scale-down completed.
	for {
		pods, err := podClient.List(ListOptions(pset.Name))
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Deployment scaled down, we can delete it.
	if err := psetClient.Delete(pset.Name, nil); err != nil {
		return err
	}

	// Delete the auto-generate configuration.
	// TODO(fabxc): add an ownerRef at creation so we don't delete config maps
	// manually created for Prometheus servers with no ServiceMonitor selectors.
	cm := c.kclient.Core().ConfigMaps(pset.Namespace)

	if err := cm.Delete(pset.Name, nil); err != nil {
		return err
	}
	if err := cm.Delete(fmt.Sprintf("%s-rules", pset.Name), nil); err != nil {
		return err
	}
	return nil
}

func (c *Operator) createConfig(p *spec.Prometheus) error {
	smons, err := c.selectServiceMonitors(p)
	if err != nil {
		return err
	}
	// Update config map based on the most recent configuration.
	b, err := generateConfig(p, smons)
	if err != nil {
		return fmt.Errorf("generating config failed: %s", err)
	}

	cm := &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: p.Name,
		},
		Data: map[string]string{
			"prometheus.yaml": string(b),
		},
	}

	cmClient := c.kclient.Core().ConfigMaps(p.Namespace)

	_, err = cmClient.Get(p.Name)
	if apierrors.IsNotFound(err) {
		_, err = cmClient.Create(cm)
	} else if err == nil {
		_, err = cmClient.Update(cm)
	}
	return err
}

func (c *Operator) selectServiceMonitors(p *spec.Prometheus) (map[string]*spec.ServiceMonitor, error) {
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*spec.ServiceMonitor)

	selector, err := unversioned.LabelSelectorAsSelector(p.Spec.ServiceMonitorSelector)
	if err != nil {
		return nil, err
	}

	// Only service monitors within the same namespace as the Prometheus
	// object can belong to it.
	cache.ListAllByNamespace(c.smonInf.GetIndexer(), p.Namespace, selector, func(obj interface{}) {
		k, err := keyFunc(obj)
		if err != nil {
			// Keep going for other items.
			utilruntime.HandleError(fmt.Errorf("key func failed: %s", err))
			return
		}
		res[k] = obj.(*spec.ServiceMonitor)
	})

	return res, nil
}

func (c *Operator) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprServiceMonitor,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: TPRVersion},
			},
			Description: "Prometheus monitoring for a service",
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprPrometheus,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: TPRVersion},
			},
			Description: "Managed Prometheus server",
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
	err := k8sutil.WaitForTPRReady(c.kclient.CoreClient.GetRESTClient(), TPRGroup, TPRVersion, TPRPrometheusesKind)
	if err != nil {
		return err
	}
	return k8sutil.WaitForTPRReady(c.kclient.CoreClient.GetRESTClient(), TPRGroup, TPRVersion, TPRServiceMonitorsKind)
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
