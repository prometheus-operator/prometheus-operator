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
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"

	"github.com/prometheus-operator/prometheus-operator/pkg/server"
)

// Config defines configuration parameters for the Operator.
type Config struct {
	// Kubernetes client configuration.
	Host              string
	TLSInsecure       bool
	TLSConfig         rest.TLSClientConfig
	ImpersonateUser   string
	KubernetesVersion version.Info

	ClusterDomain                string
	KubeletObject                string
	KubeletSelector              string
	ListenAddress                string
	ServerTLSConfig              server.TLSServerConfig
	ReloaderConfig               ContainerConfig
	AlertmanagerDefaultBaseImage string
	PrometheusDefaultBaseImage   string
	ThanosDefaultBaseImage       string
	Namespaces                   Namespaces
	Annotations                  Map
	Labels                       Map
	LocalHost                    string
	LogLevel                     string
	LogFormat                    string
	PromSelector                 string
	AlertManagerSelector         string
	ThanosRulerSelector          string
	SecretListWatchSelector      string
}

func DefaultConfig(cpu, memory string) Config {
	return Config{
		ReloaderConfig: ContainerConfig{
			CPURequests:    Quantity{q: resource.MustParse(cpu)},
			CPULimits:      Quantity{q: resource.MustParse(cpu)},
			MemoryRequests: Quantity{q: resource.MustParse(memory)},
			MemoryLimits:   Quantity{q: resource.MustParse(memory)},
		},
	}
}

// ContainerConfig holds some configuration for the ConfigReloader sidecar
// that can be set through prometheus-operator command line arguments
type ContainerConfig struct {
	// The struct tag are needed for github.com/mitchellh/hashstructure to take
	// the field values into account when generating the statefulset hash.
	CPURequests    Quantity `hash:"string"`
	CPULimits      Quantity `hash:"string"`
	MemoryRequests Quantity `hash:"string"`
	MemoryLimits   Quantity `hash:"string"`
	Image          string
	EnableProbes   bool
}

func (cc ContainerConfig) ResourceRequirements() v1.ResourceRequirements {
	resources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}

	if cc.CPURequests.String() != "0" {
		resources.Requests[v1.ResourceCPU] = cc.CPURequests.q
	}
	if cc.CPULimits.String() != "0" {
		resources.Limits[v1.ResourceCPU] = cc.CPULimits.q
	}
	if cc.MemoryRequests.String() != "0" {
		resources.Requests[v1.ResourceMemory] = cc.MemoryRequests.q
	}
	if cc.MemoryLimits.String() != "0" {
		resources.Limits[v1.ResourceMemory] = cc.MemoryLimits.q
	}

	return resources
}

type Quantity struct {
	q resource.Quantity
}

var _ = fmt.Stringer(Quantity{})

// String implements the flag.Value and fmt.Stringer interfaces.
func (q Quantity) String() string {
	return q.q.String()
}

// Set implements the flag.Value interface.
func (q *Quantity) Set(value string) error {
	if value == "" {
		return nil
	}

	quantity, err := resource.ParseQuantity(value)
	if err == nil {
		q.q = quantity
	}

	return err
}

type Map map[string]string

// String implements the flag.Value interface
func (m *Map) String() string {
	if m == nil {
		return ""
	}

	kv := make([]string, 0, len(*m))
	for _, k := range m.SortedKeys() {
		kv = append(kv, fmt.Sprintf("%s=%s", k, (*m)[k]))
	}

	return strings.Join(kv, ",")
}

// Merge returns a map which is a merge of the original map and the other parameter.
// The keys of the original map take precedence over other.
func (m *Map) Merge(other map[string]string) map[string]string {
	merged := map[string]string{}

	for key, value := range other {
		merged[key] = value
	}

	if m == nil {
		return merged
	}

	for key, value := range *m {
		merged[key] = value
	}

	return merged
}

// Set implements the flag.Value interface.
func (m *Map) Set(value string) error {
	if value == "" {
		return nil
	}

	if *m == nil {
		*m = map[string]string{}
	}

	for _, pair := range strings.Split(value, ",") {
		pair := strings.Split(pair, "=")
		(*m)[pair[0]] = pair[1]
	}

	return nil
}

// SortedKeys returns a slice of the keys in increasing order.
func (m *Map) SortedKeys() []string {
	if m == nil {
		return nil
	}

	keys := maps.Keys(*m)
	sort.Strings(keys)

	return keys
}

type Namespaces struct {
	// Allow list/deny list for common custom resources.
	AllowList, DenyList map[string]struct{}
	// Allow list for prometheus/alertmanager custom resources.
	PrometheusAllowList, AlertmanagerAllowList, AlertmanagerConfigAllowList, ThanosRulerAllowList map[string]struct{}
}
