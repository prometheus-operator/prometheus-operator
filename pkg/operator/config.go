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
	"k8s.io/client-go/rest"

	"github.com/prometheus-operator/prometheus-operator/pkg/server"
)

// Config defines configuration parameters for the Operator.
type Config struct {
	Host                         string
	ClusterDomain                string
	KubeletObject                string
	KubeletSelector              string
	ListenAddress                string
	TLSInsecure                  bool
	TLSConfig                    rest.TLSClientConfig
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

// ContainerConfig holds some configuration for the ConfigReloader sidecar
// that can be set through prometheus-operator command line arguments
type ContainerConfig struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
	Image         string
	EnableProbes  bool
}

type Map map[string]string

// Implement the flag.Value interface
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

// Set implements the flag.Set interface.
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
