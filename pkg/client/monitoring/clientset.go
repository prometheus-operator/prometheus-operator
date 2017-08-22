package monitoring

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

var _ Interface = &Clientset{}

type Interface interface {
	MonitoringV1alpha1() v1alpha1.MonitoringV1alpha1Interface
	MonitoringV1() v1.MonitoringV1Interface
}

type Clientset struct {
	*v1alpha1.MonitoringV1alpha1Client
	*v1.MonitoringV1Client
}

func (c *Clientset) MonitoringV1alpha1() v1alpha1.MonitoringV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.MonitoringV1alpha1Client
}

func (c *Clientset) MonitoringV1() v1.MonitoringV1Interface {
	if c == nil {
		return nil
	}
	return c.MonitoringV1Client
}

func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error

	cs.MonitoringV1alpha1Client, err = v1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.MonitoringV1Client, err = v1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
