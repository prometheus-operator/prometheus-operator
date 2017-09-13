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

package alertrule

import (
	"fmt"
	"time"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	prometheusoperator "github.com/coreos/prometheus-operator/pkg/prometheus"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	resyncPeriod = 5 * time.Minute
)

type Operator struct {
	kclient   kubernetes.Interface
	mclient   monitoring.Interface
	crdclient apiextensionsclient.Interface
	logger    log.Logger

	alrtruleInf cache.SharedIndexInformer  // Alertrule Informer

	queue workqueue.RateLimitingInterface

	config Config
}

type Config struct {
}

// New creates a new controller.
func New(c prometheusoperator.Config, logger log.Logger) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(c.Host, c.TLSInsecure, &c.TLSConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating cluster config failed")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating kubernetes client failed")
	}

	mclient, err := monitoring.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}

	crdclient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating apiextensions client failed")
	}

	o := &Operator{
		kclient:   client,
		mclient:   mclient,
		crdclient: crdclient,
		logger:    logger,
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "alertrule"),
		config:    Config{},
	}

	o.alrtruleInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: o.mclient.MonitoringV1alpha1().Alertrules(api.NamespaceAll).List,
			WatchFunc: o.mclient.MonitoringV1alpha1().Alertrules(api.NamespaceAll).Watch,
		},
		&v1alpha1.Alertrule{}, resyncPeriod, cache.Indexers{},
	)

	o.alrtruleInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleAlertruleAdd,
		DeleteFunc: o.handleAlertruleDelete,
		UpdateFunc: o.handleAlertruleUpdate,
	})

	return o, nil
}

func (c *Operator) RegisterMetrics(r prometheus.Registerer) {
	//r.MustRegister(NewAlertmanagerCollector(c.alrtInf.GetStore()))
	//TODO:
}

// Run the controller.
func (c *Operator) Run(stopc <-chan struct{}) error {
	defer c.queue.ShutDown()

	errChan := make(chan error)
	go func() {
		v, err := c.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- errors.Wrap(err, "communicating with server failed")
			return
		}
		c.logger.Log("msg", "connection established", "cluster-version", v)

		mv, err := k8sutil.GetMinorVersion(c.kclient.Discovery())
		if mv < 7 {
			if err := c.createTPRs(); err != nil {
				errChan <- errors.Wrap(err, "creating TPRs failed")
				return
			}

			errChan <- nil
			return
		}

		if err := c.createCRDs(); err != nil {
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
		c.logger.Log("msg", "CRD API endpoints ready")
	case <-stopc:
		return nil
	}

	go c.worker()

	go c.alrtruleInf.Run(stopc)

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

func (c *Operator) getObject(obj interface{}) (metav1.Object, bool) {
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

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *Operator) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *Operator) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.sync(key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("Sync %q failed", key)))
	c.queue.AddRateLimited(key)

	return true
}

func (c *Operator) handleAlertruleAdd(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	c.logger.Log("msg", "Alertrule added", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleAlertruleDelete(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	c.logger.Log("msg", "Alertrule deleted", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleAlertruleUpdate(old, cur interface{}) {
	key, ok := c.keyFunc(cur)
	if !ok {
		return
	}

	c.logger.Log("msg", "Alertrule updated", "key", key)
	c.enqueue(key)
}

func (c *Operator) sync(key string) error {
	// TODO: add handling code here
	fmt.Printf("sync: key=%s\n", key)
	obj, exists, err := c.alrtruleInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if exists {
		fmt.Println(obj)
	} else {
		fmt.Printf("Not exists: %s\n", key)
	}
	return nil
}

func (c *Operator) createCRDs() error {
	crds := []*extensionsobj.CustomResourceDefinition{
		k8sutil.NewAlertruleCustomResourceDefinition(),
	}

	crdClient := c.crdclient.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crd := range crds {
		if _, err := crdClient.Create(crd); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "Creating CRD: %s", crd.Spec.Names.Kind)
		}
		c.logger.Log("msg", "CRD created", "crd", crd.Spec.Names.Kind)
	}

	// We have to wait for the CRDs to be ready. Otherwise the initial watch may fail.
	return k8sutil.WaitForCRDReady(c.mclient.MonitoringV1alpha1().Alertrules(api.NamespaceAll).List)
}

func (c *Operator) createTPRs() error {
	fmt.Println("TPR support to come")
	//tprs := []*extensionsobjold.ThirdPartyResource{
	//	k8sutil.NewAlertmanagerTPRDefinition(),
	//}
	//tprClient := c.kclient.Extensions().ThirdPartyResources()
	//
	//for _, tpr := range tprs {
	//	if _, err := tprClient.Create(tpr); err != nil && !apierrors.IsAlreadyExists(err) {
	//		return err
	//	}
	//	c.logger.Log("msg", "TPR created", "tpr", tpr.Name)
	//}
	//
	//// We have to wait for the TPRs to be ready. Otherwise the initial watch may fail.
	//return k8sutil.WaitForCRDReady(c.mclient.MonitoringV1alpha1().Alertmanagers(api.NamespaceAll).List)
	return nil
}
