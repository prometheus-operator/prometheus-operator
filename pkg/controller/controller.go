package controller

import (
	"fmt"
	"net/url"
	"os"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/rest"
)

const (
	tprMonitor    = "monitor.prometheus.coreos.com"
	tprPrometheus = "prometheus.prometheus.coreos.com"
)

// Controller manages lify cycle of Prometheus deployments and
// monitoring configurations.
type Controller struct {
	kclient *kubernetes.Clientset
	logger  log.Logger
}

type Config struct {
	Host        string
	TLSInsecure bool
	TLSConfig   rest.TLSClientConfig
}

func newClient(host string, tlsInsecure bool, tlsConfig *rest.TLSClientConfig) (*kubernetes.Clientset, error) {
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

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// New creates a new controller.
func New(c Config) (*Controller, error) {
	client, err := newClient(c.Host, c.TLSInsecure, &c.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Controller{
		kclient: client,
		logger:  log.NewLogfmtLogger(os.Stdout),
	}, nil
}

// Run the controller.
func (c *Controller) Run() error {
	v, err := c.kclient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("communicating with server failed: %s", err)
	}
	c.logger.Log("msg", "connection established", "version", v)

	select {}
	panic("unreachable")
}
