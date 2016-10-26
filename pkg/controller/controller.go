package controller

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/coreos/kube-prometheus-controller/pkg/spec"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	apierrors "k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	extensionsobj "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/labels"
	utilruntime "k8s.io/client-go/1.5/pkg/util/runtime"
	"k8s.io/client-go/1.5/pkg/util/wait"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/cache"
)

const (
	tprServiceMonitor = "service-monitor.prometheus.coreos.com"
	tprPrometheus     = "prometheus.prometheus.coreos.com"
)

// Controller manages lify cycle of Prometheus deployments and
// monitoring configurations.
type Controller struct {
	kclient *kubernetes.Clientset
	pclient *rest.RESTClient
	logger  log.Logger

	promInf cache.SharedIndexInformer
	smonInf cache.SharedIndexInformer
	cmapInf cache.SharedIndexInformer
	deplInf cache.SharedIndexInformer

	queue *queue

	host string
}

// Config defines configuration parameters for the Controller.
type Config struct {
	Host        string
	TLSInsecure bool
	TLSConfig   rest.TLSClientConfig
}

// New creates a new controller.
func New(c Config) (*Controller, error) {
	cfg, err := newClusterConfig(c.Host, c.TLSInsecure, &c.TLSConfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	logger := log.NewContext(log.NewLogfmtLogger(os.Stdout)).
		With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	promclient, err := newPrometheusRESTClient(*cfg)
	if err != nil {
		return nil, err
	}
	return &Controller{
		kclient: client,
		pclient: promclient,
		logger:  logger,
		queue:   newQueue(200),
		host:    cfg.Host,
	}, nil
}

// Run the controller.
func (c *Controller) Run(stopc <-chan struct{}) error {
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
	c.deplInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Extensions().GetRESTClient(), "deployments", api.NamespaceAll, nil),
		&extensionsobj.Deployment{}, resyncPeriod, cache.Indexers{},
	)

	go c.promInf.Run(stopc)
	go c.smonInf.Run(stopc)
	go c.cmapInf.Run(stopc)
	go c.deplInf.Run(stopc)

	for !c.promInf.HasSynced() || !c.smonInf.HasSynced() || !c.cmapInf.HasSynced() || !c.deplInf.HasSynced() {
		time.Sleep(100 * time.Millisecond)
	}

	c.promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(p interface{}) {
			c.logger.Log("msg", "prominf add")
			c.enqueuePrometheus(p)
		},
		DeleteFunc: func(p interface{}) {
			c.logger.Log("msg", "prominf del")
			c.enqueuePrometheus(p)
		},
		UpdateFunc: func(_, p interface{}) {
			c.logger.Log("msg", "prominf up")
			c.enqueuePrometheus(p)
		},
	})
	c.smonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(_ interface{}) { c.enqueueAll() },
		DeleteFunc: func(_ interface{}) { c.enqueueAll() },
		UpdateFunc: func(_, _ interface{}) { c.enqueueAll() },
	})
	c.cmapInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// TODO(fabxc): only enqueue Prometheus the ConfigMap belonged to.
		DeleteFunc: func(_ interface{}) { c.enqueueAll() },
	})
	c.deplInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// TODO(fabxc): only enqueue Prometheus an affected deployment belonged to.
		AddFunc:    func(_ interface{}) { c.enqueueAll() },
		DeleteFunc: func(_ interface{}) { c.enqueueAll() },
		UpdateFunc: func(_, _ interface{}) { c.enqueueAll() },
	})

	<-stopc
	return nil
}

type queue struct {
	ch chan *spec.Prometheus
}

func newQueue(size int) *queue {
	return &queue{ch: make(chan *spec.Prometheus, size)}
}

func (q *queue) add(p *spec.Prometheus) { q.ch <- p }
func (q *queue) close()                 { close(q.ch) }

func (q *queue) pop() (*spec.Prometheus, bool) {
	p, ok := <-q.ch
	return p, ok
}

var keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc

func (c *Controller) enqueuePrometheus(p interface{}) {
	c.queue.add(p.(*spec.Prometheus))
}

func (c *Controller) enqueueAll() {
	cache.ListAll(c.promInf.GetStore(), labels.Everything(), func(o interface{}) {
		c.enqueuePrometheus(o.(*spec.Prometheus))
	})
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *Controller) worker() {
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

func (c *Controller) reconcile(p *spec.Prometheus) error {
	key, err := keyFunc(p)
	if err != nil {
		return err
	}
	c.logger.Log("msg", "reconcile prometheus", "key", key)

	_, exists, err := c.promInf.GetStore().GetByKey(key)
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
		// set for now. This does not work if we delete Prometheus resources as the
		// controller is not running – that could be solved via garbage collection later.
		return c.deletePrometheus(p)
	}

	// Ensure we have a replica set running Prometheus deployed.
	// XXX: Selecting by ObjectMeta.Name gives an error. So use the label for now.
	selector, err := labels.Parse("prometheus.coreos.com/type=prometheus,prometheus.coreos.com/name=" + p.Name)
	if err != nil {
		return err
	}
	var rs []*extensionsobj.ReplicaSet
	cache.ListAllByNamespace(c.promInf.GetIndexer(), p.Namespace, selector, func(o interface{}) {
		rs = append(rs, o.(*extensionsobj.ReplicaSet))
	})
	if len(rs) == 0 {
		deplClient := c.kclient.ExtensionsClient.Deployments(p.Namespace)
		if _, err := deplClient.Create(makeDeployment(p.Name, 1)); err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create replica set: %s", err)
		}
	}

	// We just always regenerate the configuration to be safe.
	if err := c.createConfig(p); err != nil {
		return err
	}
	return nil
}

func (c *Controller) deletePrometheus(p *spec.Prometheus) error {
	// Update the replica count to 0 and wait for all pods to be deleted.
	deplClient := c.kclient.ExtensionsClient.Deployments(p.Namespace)
	// depl := c.kclient.Extensions().Deployments(p.Namespace)
	depl, err := deplClient.Update(makeDeployment(p.Name, 0))
	if err != nil {
		return err
	}

	// XXX: Selecting by ObjectMeta.Name gives an error. So use the label for now.
	selector, err := labels.Parse("prometheus.coreos.com/type=prometheus,prometheus.coreos.com/name=" + p.Name)
	if err != nil {
		return err
	}
	w, err := deplClient.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	if _, err := watch.Until(100*time.Second, w, func(e watch.Event) (bool, error) {
		curDepl, ok := e.Object.(*extensionsobj.Deployment)
		if !ok {
			return false, errors.New("not a replica set")
		}
		// Check if the replica set is scaled down and all replicas are gone.
		if curDepl.Status.ObservedGeneration >= depl.Status.ObservedGeneration && curDepl.Status.Replicas == *curDepl.Spec.Replicas {
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

	// Deployment scaled down, we can delete it.
	if err := deplClient.Delete(p.Name, nil); err != nil {
		return err
	}
	// Remove ReplicaSet of the deployment.
	rsClient := c.kclient.Extensions().ReplicaSets(p.Namespace)
	if err := rsClient.DeleteCollection(&api.DeleteOptions{}, api.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		return err
	}

	// Delete the auto-generate configuration.
	// TODO(fabxc): add an ownerRef at creation so we don't delete config maps
	// manually created for Prometheus servers with no ServiceMonitor selectors.
	cm := c.kclient.Core().ConfigMaps(p.Namespace)
	if err := cm.Delete(p.Name, nil); err != nil {
		return err
	}
	return nil
}

func (c *Controller) createConfig(p *spec.Prometheus) error {
	smons, err := c.selectServiceMonitors(p)
	if err != nil {
		return err
	}
	// Update config map based on the most recent configuration.
	var buf bytes.Buffer
	if err := configTmpl.Execute(&buf, &templateConfig{
		ServiceMonitors: smons,
		Prometheus:      p.Spec,
	}); err != nil {
		return err
	}

	cm := &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: p.Name,
		},
		Data: map[string]string{
			"prometheus.yaml": buf.String(),
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

func (c *Controller) selectServiceMonitors(p *spec.Prometheus) (map[string]*spec.ServiceMonitor, error) {
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*spec.ServiceMonitor)

	for _, smon := range p.Spec.ServiceMonitors {
		selector, err := unversioned.LabelSelectorAsSelector(&smon.Selector)
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
	}
	return res, nil
}

func (c *Controller) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprServiceMonitor,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: "v1alpha1"},
			},
			Description: "Prometheus monitoring for a service",
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprPrometheus,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: "v1alpha1"},
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
	err := wait.Poll(3*time.Second, 30*time.Second, func() (bool, error) {
		resp, err := c.kclient.CoreClient.Client.Get(c.host + "/apis/prometheus.coreos.com/v1alpha1/prometheuses")
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			return true, nil
		case http.StatusNotFound: // not set up yet. wait.
			return false, nil
		default:
			return false, fmt.Errorf("invalid status code: %v", resp.Status)
		}
	})
	if err != nil {
		return err
	}
	return wait.Poll(3*time.Second, 30*time.Second, func() (bool, error) {
		resp, err := c.kclient.CoreClient.Client.Get(c.host + "/apis/prometheus.coreos.com/v1alpha1/servicemonitors")
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			return true, nil
		case http.StatusNotFound: // not set up yet. wait.
			return false, nil
		default:
			return false, fmt.Errorf("invalid status code: %v", resp.Status)
		}
	})
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
