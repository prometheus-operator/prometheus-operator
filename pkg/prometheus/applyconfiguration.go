// Copyright 2019 The prometheus-operator Authors
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

package prometheus

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringv1ac "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	monitoringv1alpha1ac "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1alpha1"
)

func ApplyConfigurationFromPrometheusAgent(p *monitoringv1alpha1.PrometheusAgent, updateScaleSubresource bool) *monitoringv1alpha1ac.PrometheusAgentApplyConfiguration {
	psac := prometheusStatusApplyConfigurationFromPrometheusStatus(&p.Status, updateScaleSubresource)
	return monitoringv1alpha1ac.PrometheusAgent(p.Name, p.Namespace).WithStatus(psac)
}

// ApplyConfigurationFromPrometheus updates the Prometheus/PrometheusAgent Status subresource.
// It can optionally update the scale subresource as well.
func ApplyConfigurationFromPrometheus(p *monitoringv1.Prometheus, updateScaleSubresource bool) *monitoringv1ac.PrometheusApplyConfiguration {
	psac := prometheusStatusApplyConfigurationFromPrometheusStatus(&p.Status, updateScaleSubresource)
	return monitoringv1ac.Prometheus(p.Name, p.Namespace).WithStatus(psac)
}

func prometheusStatusApplyConfigurationFromPrometheusStatus(status *monitoringv1.PrometheusStatus, updateScaleSubresource bool) *monitoringv1ac.PrometheusStatusApplyConfiguration {
	psac := monitoringv1ac.PrometheusStatus().
		WithPaused(status.Paused).
		WithReplicas(status.Replicas).
		WithAvailableReplicas(status.AvailableReplicas).
		WithUpdatedReplicas(status.UpdatedReplicas).
		WithUnavailableReplicas(status.UnavailableReplicas)

	if updateScaleSubresource {
		psac = psac.WithShards(status.Shards).WithSelector(status.Selector)
	}

	for _, condition := range status.Conditions {
		psac.WithConditions(
			monitoringv1ac.Condition().
				WithType(condition.Type).
				WithStatus(condition.Status).
				WithLastTransitionTime(condition.LastTransitionTime).
				WithReason(condition.Reason).
				WithMessage(condition.Message).
				WithObservedGeneration(condition.ObservedGeneration),
		)
	}

	for _, shardStatus := range status.ShardStatuses {
		psac.WithShardStatuses(
			monitoringv1ac.ShardStatus().
				WithShardID(shardStatus.ShardID).
				WithReplicas(shardStatus.Replicas).
				WithUpdatedReplicas(shardStatus.UpdatedReplicas).
				WithAvailableReplicas(shardStatus.AvailableReplicas).
				WithUnavailableReplicas(shardStatus.UnavailableReplicas),
		)
	}

	return psac
}

// ApplyConfigurationFromServiceMonitor updates the ServiceMonitor Status subresource.
func ApplyConfigurationFromServiceMonitor(sm *monitoringv1.ServiceMonitor) *monitoringv1ac.ServiceMonitorApplyConfiguration {
	smsac := configResourceStatusApplyConfigurationFromConfigResourceStatus(&sm.Status)
	return monitoringv1ac.ServiceMonitor(sm.Name, sm.Namespace).
		WithStatus(smsac)
}

func configResourceStatusApplyConfigurationFromConfigResourceStatus(status *monitoringv1.ConfigResourceStatus) *monitoringv1ac.ConfigResourceStatusApplyConfiguration {
	crsac := monitoringv1ac.ConfigResourceStatus()

	for _, binding := range status.Bindings {
		bg := monitoringv1ac.WorkloadBinding().
			WithGroup(binding.Group).
			WithName(binding.Name).
			WithNamespace(binding.Namespace).
			WithResource(binding.Resource)

		for _, condition := range binding.Conditions {
			bg.WithConditions(
				monitoringv1ac.ConfigResourceCondition().
					WithType(condition.Type).
					WithStatus(condition.Status).
					WithLastTransitionTime(condition.LastTransitionTime).
					WithReason(condition.Reason).
					WithMessage(condition.Message).
					WithObservedGeneration(condition.ObservedGeneration),
			)
		}

		crsac.WithBindings(bg)
	}

	return crsac
}
