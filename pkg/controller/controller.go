package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/coreos/kube-prometheus-controller/pkg/prometheus"
	"github.com/coreos/kube-prometheus-controller/pkg/spec"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.5/kubernetes"
	apierrors "k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/v1"
	extensionsobj "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
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
	kclient    *kubernetes.Clientset
	pclient    *rest.RESTClient
	logger     log.Logger
	host       string
	prometheis map[string]*prometheus.Prometheus

	promInf cache.SharedInformer
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
		kclient:    client,
		pclient:    promclient,
		logger:     logger,
		host:       cfg.Host,
		prometheis: map[string]*prometheus.Prometheus{},
	}, nil
}

// Run the controller.
func (c *Controller) Run() error {
	v, err := c.kclient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("communicating with server failed: %s", err)
	}
	c.logger.Log("msg", "connection established", "cluster-version", v)

	if err := c.createTPRs(); err != nil {
		return err
	}

	promInf := cache.NewSharedInformer(NewPrometheusListWatch(c.pclient), &spec.Prometheus{}, resyncPeriod)

	go promInf.Run(make(chan struct{}))

	for !promInf.HasSynced() {
		time.Sleep(100 * time.Millisecond)
	}

	promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(o interface{}) {
			p := o.(*spec.Prometheus)
			c.logger.Log("msg", "prometheus added", "prometheus", p.Name, "ns", p.Namespace)

			prom, err := prometheus.New(context.TODO(), c.logger, c.host, c.kclient, p)
			if err != nil {
				c.logger.Log("msg", "Prometheus creation failed", "err", err)
			} else {
				c.prometheis[p.Namespace+"\xff"+p.Name] = prom
			}
		},
		DeleteFunc: func(o interface{}) {
			p := o.(*spec.Prometheus)
			c.logger.Log("msg", "prometheus deleted", "prometheus", p.Name, "ns", p.Namespace)

			prom := c.prometheis[p.Namespace+"\xff"+p.Name]
			if prom != nil {
				prom.Delete()
				delete(c.prometheis, p.Namespace+"\xff"+p.Name)
			}
		},
		UpdateFunc: func(o, n interface{}) {
			p := n.(*spec.Prometheus)
			c.logger.Log("msg", "prometheus updated", "prometheus", p.Name, "ns", p.Namespace)

			prom := c.prometheis[p.Namespace+"\xff"+p.Name]
			if prom != nil {
				prom.Update(p)
			}
		},
	})

	select {}
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

// Event represents an event in the cluster.
type Event struct {
	Type   watch.EventType
	Object spec.Prometheus
}
