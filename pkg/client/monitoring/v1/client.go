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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
)

const (
	Group = "monitoring.coreos.com"
)

var Version = "v1"

type MonitoringV1Interface interface {
	RESTClient() rest.Interface
	PrometheusesGetter
	AlertmanagersGetter
	ServiceMonitorsGetter
}

type MonitoringV1Client struct {
	restClient    rest.Interface
	dynamicClient *dynamic.Client
}

func (c *MonitoringV1Client) Prometheuses(namespace string) PrometheusInterface {
	return newPrometheuses(c.restClient, c.dynamicClient, namespace)
}

func (c *MonitoringV1Client) Alertmanagers(namespace string) AlertmanagerInterface {
	return newAlertmanagers(c.restClient, c.dynamicClient, namespace)
}

func (c *MonitoringV1Client) ServiceMonitors(namespace string) ServiceMonitorInterface {
	return newServiceMonitors(c.restClient, c.dynamicClient, namespace)
}

func (c *MonitoringV1Client) RESTClient() rest.Interface {
	return c.restClient
}

func NewForConfig(apiGroup string, c *rest.Config) (*MonitoringV1Client, error) {
	config := *c
	SetConfigDefaults(apiGroup, &config)
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewClient(&config)
	if err != nil {
		return nil, err
	}

	return &MonitoringV1Client{client, dynamicClient}, nil
}

func SetConfigDefaults(apiGroup string, config *rest.Config) {
	config.GroupVersion = &schema.GroupVersion{
		Group:   apiGroup,
		Version: Version,
	}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
	return
}
