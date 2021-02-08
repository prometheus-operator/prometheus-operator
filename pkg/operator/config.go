// Copyright 2020 The prometheus-operator Authors
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

package operator

import (
	"strings"

	"k8s.io/client-go/rest"
)

// Config defines configuration parameters for the Operator.
type Config struct {
	Host                         string
	ClusterDomain                string
	KubeletObject                string
	ListenAddress                string
	TLSInsecure                  bool
	TLSConfig                    rest.TLSClientConfig
	ServerTLSConfig              TLSServerConfig
	ReloaderConfig               ReloaderConfig
	AlertmanagerDefaultBaseImage string
	PrometheusDefaultBaseImage   string
	ThanosDefaultBaseImage       string
	Namespaces                   Namespaces
	Labels                       Labels
	LocalHost                    string
	LogLevel                     string
	LogFormat                    string
	PromSelector                 string
	AlertManagerSelector         string
	ThanosRulerSelector          string
	SecretListWatchSelector      string
}

type ReloaderConfig struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
	Image         string
}

type Labels struct {
	LabelsString string
	LabelsMap    map[string]string
}

// Implement the flag.Value interface
func (labels *Labels) String() string {
	return labels.LabelsString
}

// Merge labels create a new map with labels merged.
func (labels *Labels) Merge(otherLabels map[string]string) map[string]string {
	mergedLabels := map[string]string{}

	for key, value := range otherLabels {
		mergedLabels[key] = value
	}

	for key, value := range labels.LabelsMap {
		mergedLabels[key] = value
	}
	return mergedLabels
}

// Set implements the flag.Set interface.
func (labels *Labels) Set(value string) error {
	m := map[string]string{}
	if value != "" {
		splited := strings.Split(value, ",")
		for _, pair := range splited {
			sp := strings.Split(pair, "=")
			m[sp[0]] = sp[1]
		}
	}
	(*labels).LabelsMap = m
	(*labels).LabelsString = value
	return nil
}

type Namespaces struct {
	// Allow list/deny list for common custom resources.
	AllowList, DenyList map[string]struct{}
	// Allow list for prometheus/alertmanager custom resources.
	PrometheusAllowList, AlertmanagerAllowList, ThanosRulerAllowList map[string]struct{}
}
