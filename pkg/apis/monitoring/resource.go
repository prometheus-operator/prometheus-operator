// Copyright 2018 The prometheus-operator Authors
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

package monitoring

import (
	"fmt"
)

const (
	PrometheusesKind  = "Prometheus"
	PrometheusName    = "prometheuses"
	PrometheusKindKey = "prometheus"

	AlertmanagersKind   = "Alertmanager"
	AlertmanagerName    = "alertmanagers"
	AlertManagerKindKey = "alertmanager"

	AlertmanagerConfigKind    = "AlertmanagerConfig"
	AlertmanagerConfigName    = "alertmanagerconfigs"
	AlertmanagerConfigKindKey = "alertmanagerconfig"

	ServiceMonitorsKind   = "ServiceMonitor"
	ServiceMonitorName    = "servicemonitors"
	ServiceMonitorKindKey = "servicemonitor"

	PodMonitorsKind   = "PodMonitor"
	PodMonitorName    = "podmonitors"
	PodMonitorKindKey = "podmonitor"

	PrometheusAgentsKind   = "PrometheusAgent"
	PrometheusAgentName    = "prometheusagents"
	PrometheusAgentKindKey = "prometheusagent"

	PrometheusRuleKind    = "PrometheusRule"
	PrometheusRuleName    = "prometheusrules"
	PrometheusRuleKindKey = "prometheusrule"

	ProbesKind   = "Probe"
	ProbeName    = "probes"
	ProbeKindKey = "probe"

	ScrapeConfigsKind   = "ScrapeConfig"
	ScrapeConfigName    = "scrapeconfigs"
	ScrapeConfigKindKey = "scrapeconfig"

	ThanosRulerKind    = "ThanosRuler"
	ThanosRulerName    = "thanosrulers"
	ThanosRulerKindKey = "thanosrulers"
)

var resourceToKindMap = map[string]string{
	PrometheusName:     PrometheusesKind,
	AlertmanagerName:   AlertmanagersKind,
	ServiceMonitorName: ServiceMonitorsKind,
	PodMonitorName:     PodMonitorsKind,
	PrometheusRuleName: PrometheusRuleKind,
	ProbeName:          ProbesKind,
	ScrapeConfigName:   ScrapeConfigsKind,
}

func ResourceToKind(s string) string {
	kind, found := resourceToKindMap[s]
	if !found {
		panic(fmt.Sprintf("failed to map resource %q to a kind", s))
	}
	return kind
}
