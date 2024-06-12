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
	"flag"
	"fmt"
	"slices"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/version"
	k8sflag "k8s.io/component-base/cli/flag"
)

// Config defines configuration parameters for the Operator.
type Config struct {
	// Version reported by the Kubernetes API.
	KubernetesVersion version.Info

	// Cluster domain for Kubernetes services managed by the operator.
	ClusterDomain string

	// Global configuration for the reloader config sidecar.
	ReloaderConfig ContainerConfig

	// Base container images for operands.
	AlertmanagerDefaultBaseImage string
	PrometheusDefaultBaseImage   string
	ThanosDefaultBaseImage       string

	// Allow and deny lists for namespace watchers.
	Namespaces Namespaces

	// Metadata applied to all resources managed by the operator.
	Annotations Map
	Labels      Map

	// Custom name to use for "localhost".
	LocalHost string

	// Label and field selectors for resource watchers.
	PromSelector            LabelSelector
	AlertmanagerSelector    LabelSelector
	ThanosRulerSelector     LabelSelector
	SecretListWatchSelector FieldSelector

	// Controller id for pod ownership.
	ControllerID string

	// Feature gates.
	Gates *FeatureGates
}

// DefaultConfig returns a default operator configuration.
func DefaultConfig(cpu, memory string) Config {
	return Config{
		ReloaderConfig: ContainerConfig{
			CPURequests:    Quantity{q: resource.MustParse(cpu)},
			CPULimits:      Quantity{q: resource.MustParse(cpu)},
			MemoryRequests: Quantity{q: resource.MustParse(memory)},
			MemoryLimits:   Quantity{q: resource.MustParse(memory)},
		},
		Namespaces: Namespaces{
			AllowList:                   StringSet{},
			DenyList:                    StringSet{},
			PrometheusAllowList:         StringSet{},
			AlertmanagerAllowList:       StringSet{},
			AlertmanagerConfigAllowList: StringSet{},
			ThanosRulerAllowList:        StringSet{},
		},
		Gates: &FeatureGates{
			PrometheusAgentDaemonSetFeature: FeatureGate{
				description: "Enables the DaemonSet mode for PrometheusAgent",
				enabled:     false,
			},
		},
	}
}

func (c *Config) RegisterFeatureGatesFlags(fs *flag.FlagSet, flags *k8sflag.MapStringBool) {
	fs.Var(
		flags,
		"feature-gates",
		fmt.Sprintf("Feature gates are a set of key=value pairs that describe Prometheus-Operator features.\n"+
			"Available feature gates:\n  %s", strings.Join(c.Gates.Descriptions(), "\n  "),
		),
	)
}

// ContainerConfig holds some configuration for the ConfigReloader sidecar
// that can be set through prometheus-operator command line arguments.
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

// String implements the flag.Value interface.
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
	// Allow list for common custom resources.
	AllowList StringSet
	// Deny list for common custom resources.
	DenyList StringSet
	// Allow list for Prometheus custom resources.
	PrometheusAllowList StringSet
	// Allow list for Alertmanager custom resources.
	AlertmanagerAllowList StringSet
	// Allow list for AlertmanagerConfig custom resources.
	AlertmanagerConfigAllowList StringSet
	// Allow list for ThanosRuler custom resources.
	ThanosRulerAllowList StringSet
}

func (n *Namespaces) String() string {
	return fmt.Sprintf("{allow_list=%q,deny_list=%q,prometheus_allow_list=%q,alertmanager_allow_list=%q,alertmanagerconfig_allow_list=%q,thanosruler_allow_list=%q}",
		n.AllowList,
		n.DenyList,
		n.PrometheusAllowList,
		n.AlertmanagerAllowList,
		n.AlertmanagerConfigAllowList,
		n.ThanosRulerAllowList,
	)
}

func (n *Namespaces) Finalize() {
	if len(n.AllowList) == 0 {
		n.AllowList.Insert(v1.NamespaceAll)
	}

	if len(n.PrometheusAllowList) == 0 {
		n.PrometheusAllowList = n.AllowList
	}

	if len(n.AlertmanagerAllowList) == 0 {
		n.AlertmanagerAllowList = n.AllowList
	}

	if len(n.AlertmanagerConfigAllowList) == 0 {
		n.AlertmanagerConfigAllowList = n.AllowList
	}

	if len(n.ThanosRulerAllowList) == 0 {
		n.ThanosRulerAllowList = n.AllowList
	}
}

type LabelSelector string

// String implements the flag.Value interface.
func (ls *LabelSelector) String() string {
	if ls == nil {
		return ""
	}

	return string(*ls)
}

// Set implements the flag.Value interface.
func (ls *LabelSelector) Set(value string) error {
	if _, err := labels.Parse(value); err != nil {
		return err
	}

	*ls = LabelSelector(value)
	return nil
}

type NodeAddressPriority string

// String implements the flag.Value interface.
func (p *NodeAddressPriority) String() string {
	if p == nil || *p == "" {
		return "internal"
	}
	return string(*p)
}

// Set implements the flag.Value interface.
func (p *NodeAddressPriority) Set(value string) error {
	if value != "internal" && value != "external" {
		return fmt.Errorf("invalid value for node address priority, expected 'internal' or 'external' but got: %q", value)
	}
	*p = NodeAddressPriority(value)
	return nil
}

type FieldSelector string

// String implements the flag.Value interface.
func (fs *FieldSelector) String() string {
	if fs == nil {
		return ""
	}

	return string(*fs)
}

// Set implements the flag.Value interface.
func (fs *FieldSelector) Set(value string) error {
	if _, err := fields.ParseSelector(value); err != nil {
		return err
	}

	*fs = FieldSelector(value)
	return nil
}

// StringSet represents a list of comma-separated strings.
type StringSet map[string]struct{}

// Set implements the flag.Value interface.
func (s StringSet) Set(value string) error {
	if s == nil {
		return fmt.Errorf("expected StringSet variable to be initialized")
	}

	for _, v := range strings.Split(value, ",") {
		s[v] = struct{}{}
	}

	return nil
}

// String implements the flag.Value interface.
func (s StringSet) String() string {
	return strings.Join(s.Slice(), ",")
}

func (s StringSet) Insert(value string) {
	s[value] = struct{}{}
}

func (s StringSet) Slice() []string {
	ss := make([]string, 0, len(s))
	for k := range s {
		ss = append(ss, k)
	}

	slices.Sort(ss)
	return ss
}
