package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/coreos/kube-prometheus-controller/pkg/prometheus"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.4/kubernetes"
	apierrors "k8s.io/client-go/1.4/pkg/api/errors"
	api "k8s.io/client-go/1.4/pkg/api/v1"
	extensionsobj "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/watch"
	"k8s.io/client-go/1.4/rest"
)

const (
	tprServiceMonitor = "service-monitor.prometheus.coreos.com"
	tprPrometheus     = "prometheus.prometheus.coreos.com"
)

// Controller manages lify cycle of Prometheus deployments and
// monitoring configurations.
type Controller struct {
	kclient    *kubernetes.Clientset
	logger     log.Logger
	host       string
	prometheis map[string]*prometheus.Prometheus
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

	return &Controller{
		kclient:    client,
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

	events, errc := c.monitorPrometheusServers(c.kclient.CoreClient.Client, "0")
	for {
		select {
		case evt := <-events:
			switch evt.Type {
			case watch.Added:
				l := log.NewContext(c.logger).With("namespace", evt.Object.Namespace, "prometheus", evt.Object.Name)
				p, err := prometheus.New(context.TODO(), l, c.host, c.kclient, &evt.Object)
				if err != nil {
					c.logger.Log("msg", "Prometheus creation failed", "err", err)
				} else {
					c.prometheis[evt.Object.Namespace+"\xff"+evt.Object.Name] = p
				}

			case watch.Modified:
				p := c.prometheis[evt.Object.Namespace+"\xff"+evt.Object.Name]
				if p != nil {
					p.Update(&evt.Object)
				}
			case watch.Deleted:
				p := c.prometheis[evt.Object.Namespace+"\xff"+evt.Object.Name]
				if p != nil {
					p.Delete()
					delete(c.prometheis, evt.Object.Namespace+"\xff"+evt.Object.Name)
				}

			default:
				c.logger.Log("msg", "unknown event type received", "type", evt.Type)
			}
		case err := <-errc:
			c.logger.Log("msg", "received error on Prometheus watch", "err", err)
		}
	}
}

func (c *Controller) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: api.ObjectMeta{
				Name: tprServiceMonitor,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: "v1"},
			},
			Description: "Prometheus monitoring for a service",
		},
		{
			ObjectMeta: api.ObjectMeta{
				Name: tprPrometheus,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: "v1"},
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
	return nil
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
	Object prometheus.PrometheusObj
}

func (c *Controller) monitorPrometheusServers(client *http.Client, watchVersion string) (<-chan *Event, <-chan error) {
	var (
		events = make(chan *Event)
		errc   = make(chan error, 1)
	)
	go func() {
		for {
			resp, err := client.Get(c.host + "/apis/prometheus.coreos.com/v1/namespaces/default/prometheuses?watch=true&resourceVersion=" + watchVersion)
			if err != nil {
				errc <- err
				return
			}
			if resp.StatusCode != 200 {
				resp.Body.Close()
				errc <- errors.New("Invalid status code: " + resp.Status)
				return
			}
			c.logger.Log("msg", "watching Prometheus resource", "version", watchVersion)
			dec := json.NewDecoder(resp.Body)

			for {
				var evt Event
				if err := dec.Decode(&evt); err != nil {
					if err == io.EOF {
						break
					}
					c.logger.Log("msg", "failed to get event from apiserver", "err", err)
					errc <- err
					break
				}

				if evt.Type == "ERROR" {
					watchVersion = "0"
					c.logger.Log("msg", "failed to get event from apiserver", "errevt", fmt.Sprintf("%+v", evt))
					break
				}
				c.logger.Log("msg", "Prometheus event", "type", evt.Type, "obj", fmt.Sprintf("%v:%v", evt.Object.Namespace, evt.Object.Name))
				watchVersion = evt.Object.ObjectMeta.ResourceVersion

				events <- &evt
			}
			resp.Body.Close()
		}
	}()
	return events, errc
}
