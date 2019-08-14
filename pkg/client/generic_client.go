package client

import (
	monitoringclientset "github.com/coreos/prometheus-operator/pkg/client/versioned"
	kubeclientset "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
)

type GenericClientset struct {
	KubeClient   kubeclientset.Interface
	MonitoringClient monitoringclientset.Interface
}

// NewForConfig creates a new Clientset for the given config.
func newForConfig(c *rest.Config) (*GenericClientset, error) {
	kubeClient, err := kubeclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	monitoringClient, err := monitoringclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &GenericClientset{
		KubeClient:   kubeClient,
		MonitoringClient: monitoringClient,
	}, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func newForConfigOrDie(c *rest.Config) *GenericClientset {
	return &GenericClientset{
		KubeClient:   kubeclientset.NewForConfigOrDie(c),
		MonitoringClient: monitoringclientset.NewForConfigOrDie(c),
	}
}

// New creates a new Clientset for the given RESTClient.
func new(c rest.Interface) *GenericClientset {
	return &GenericClientset{
		KubeClient:   kubeclientset.New(c),
		MonitoringClient: monitoringclientset.New(c),
	}
}
