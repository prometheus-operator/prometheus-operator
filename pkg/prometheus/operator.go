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

package prometheus

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

// Config defines the operator's parameters for the Prometheus controllers.
// Whenever the value of one of these parameters is changed, it triggers an
// update of the managed statefulsets.
type Config struct {
	LocalHost                  string
	ReloaderConfig             operator.ContainerConfig
	PrometheusDefaultBaseImage string
	ThanosDefaultBaseImage     string
	Annotations                operator.Map
	Labels                     operator.Map
}

type StatusReporter struct {
	Kclient         kubernetes.Interface
	Reconciliations *operator.ReconciliationTracker
	SsetInfs        *informers.ForResource
	Rr              *operator.ResourceReconciler
}

type ConfigResourceSyncer struct {
	gvr      schema.GroupVersionResource
	mclient  monitoringclient.Interface
	workload metav1.Object // Workload resource (Prometheus and PrometheusAgent) selecting the configuration resources.
}

func NewConfigResourceSyncer(gvr schema.GroupVersionResource, mclient monitoringclient.Interface, workload metav1.Object) *ConfigResourceSyncer {
	return &ConfigResourceSyncer{
		gvr:      gvr,
		mclient:  mclient,
		workload: workload,
	}
}

// UpdateBindingConditions returns the bindings slice with the conditions updated for the workload.
// The 2nd return value indicates if the slice has been updated.
func (crs *ConfigResourceSyncer) UpdateBindingConditions(bindings []monitoringv1.WorkloadBinding, conditions []monitoringv1.ConfigResourceCondition) ([]monitoringv1.WorkloadBinding, bool) {
	updated := monitoringv1.WorkloadBinding{
		Namespace:  crs.workload.GetNamespace(),
		Name:       crs.workload.GetName(),
		Resource:   crs.gvr.Resource,
		Group:      crs.gvr.Group,
		Conditions: conditions,
	}
	for i, binding := range bindings {
		if binding.Namespace != updated.Namespace ||
			binding.Name != updated.Name ||
			binding.Group != updated.Group ||
			binding.Resource != updated.Resource {
			continue
		}

		// No need to update the binding if the conditions haven't changed
		if equalConfigResourceConditions(binding.Conditions, updated.Conditions) {
			return nil, false
		}

		bindings[i] = updated
		return bindings, true
	}

	return append(bindings, updated), true
}

func KeyToStatefulSetKey(p monitoringv1.PrometheusInterface, key string, shard int) string {
	keyParts := strings.Split(key, "/")
	return fmt.Sprintf("%s/%s", keyParts[0], statefulSetNameFromPrometheusName(p, keyParts[1], shard))
}

func statefulSetNameFromPrometheusName(p monitoringv1.PrometheusInterface, name string, shard int) string {
	if shard == 0 {
		return fmt.Sprintf("%s-%s", Prefix(p), name)
	}
	return fmt.Sprintf("%s-%s-shard-%d", Prefix(p), name, shard)
}

func NewTLSAssetSecret(p monitoringv1.PrometheusInterface, config Config) *v1.Secret {
	s := &v1.Secret{
		Data: map[string][]byte{},
	}

	operator.UpdateObject(
		s,
		operator.WithLabels(config.Labels),
		operator.WithAnnotations(config.Annotations),
		operator.WithManagingOwner(p),
		operator.WithName(TLSAssetsSecretName(p)),
		operator.WithNamespace(p.GetObjectMeta().GetNamespace()),
	)

	return s
}

// validateRemoteWriteSpec checks that mutually exclusive configurations are not
// included in the Prometheus remoteWrite configuration section, while also validating
// the RemoteWriteSpec child fields.
// Reference:
// https://github.com/prometheus/prometheus/blob/main/docs/configuration/configuration.md#remote_write
func validateRemoteWriteSpec(spec monitoringv1.RemoteWriteSpec) error {
	var nonNilFields []string
	for k, v := range map[string]any{
		"basicAuth":     spec.BasicAuth,
		"oauth2":        spec.OAuth2,
		"authorization": spec.Authorization,
		"sigv4":         spec.Sigv4,
		"azureAd":       spec.AzureAD,
	} {
		if reflect.ValueOf(v).IsNil() {
			continue
		}

		nonNilFields = append(nonNilFields, fmt.Sprintf("%q", k))
	}

	if len(nonNilFields) > 1 {
		return fmt.Errorf("%s can't be set at the same time, at most one of them must be defined", strings.Join(nonNilFields, " and "))
	}

	if spec.AzureAD != nil {
		if spec.AzureAD.ManagedIdentity == nil && spec.AzureAD.OAuth == nil && spec.AzureAD.SDK == nil {
			return fmt.Errorf("must provide Azure Managed Identity or Azure OAuth or Azure SDK in the Azure AD config")
		}

		if spec.AzureAD.ManagedIdentity != nil && spec.AzureAD.OAuth != nil {
			return fmt.Errorf("cannot provide both Azure Managed Identity and Azure OAuth in the Azure AD config")
		}

		if spec.AzureAD.OAuth != nil && spec.AzureAD.SDK != nil {
			return fmt.Errorf("cannot provide both Azure OAuth and Azure SDK in the Azure AD config")
		}

		if spec.AzureAD.ManagedIdentity != nil && spec.AzureAD.SDK != nil {
			return fmt.Errorf("cannot provide both Azure Managed Identity and Azure SDK in the Azure AD config")
		}

		if spec.AzureAD.OAuth != nil {
			_, err := uuid.Parse(spec.AzureAD.OAuth.ClientID)
			if err != nil {
				return fmt.Errorf("the provided Azure OAuth clientId is invalid")
			}
		}
	}

	return spec.Validate()
}

// Process will determine the Status of a Prometheus resource (server or agent) depending on its current state in the cluster.
func (sr *StatusReporter) Process(ctx context.Context, p monitoringv1.PrometheusInterface, key string) (*monitoringv1.PrometheusStatus, error) {

	commonFields := p.GetCommonPrometheusFields()
	pStatus := monitoringv1.PrometheusStatus{
		Paused: commonFields.Paused,
	}

	var (
		availableStatus    = monitoringv1.ConditionTrue
		availableReason    string
		availableCondition = monitoringv1.Condition{
			Type: monitoringv1.Available,
			LastTransitionTime: metav1.Time{
				Time: time.Now().UTC(),
			},
			ObservedGeneration: p.GetObjectMeta().GetGeneration(),
		}
		messages []string
		replicas = 1
	)

	if commonFields.Replicas != nil {
		replicas = int(*commonFields.Replicas)
	}

	for shard := range ExpectedStatefulSetShardNames(p) {
		ssetName := KeyToStatefulSetKey(p, key, shard)

		obj, err := sr.SsetInfs.Get(ssetName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Statefulset hasn't been created or is already deleted.
				availableStatus = monitoringv1.ConditionFalse
				availableReason = "StatefulSetNotFound"
				messages = append(messages, fmt.Sprintf("shard %d: statefulset %s not found", shard, ssetName))
				pStatus.ShardStatuses = append(
					pStatus.ShardStatuses,
					monitoringv1.ShardStatus{
						ShardID: strconv.Itoa(shard),
					})

				continue
			}

			return nil, fmt.Errorf("failed to retrieve statefulset: %w", err)
		}

		sset := obj.(*appsv1.StatefulSet).DeepCopy()
		if sr.Rr.DeletionInProgress(sset) {
			continue
		}

		stsReporter, err := operator.NewStatefulSetReporter(ctx, sr.Kclient, sset)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve statefulset state: %w", err)
		}

		pStatus.Replicas += int32(len(stsReporter.Pods))
		pStatus.UpdatedReplicas += int32(len(stsReporter.UpdatedPods()))
		pStatus.AvailableReplicas += int32(len(stsReporter.ReadyPods()))
		pStatus.UnavailableReplicas += int32(len(stsReporter.Pods) - len(stsReporter.ReadyPods()))

		pStatus.ShardStatuses = append(
			pStatus.ShardStatuses,
			monitoringv1.ShardStatus{
				ShardID:             strconv.Itoa(shard),
				Replicas:            int32(len(stsReporter.Pods)),
				UpdatedReplicas:     int32(len(stsReporter.UpdatedPods())),
				AvailableReplicas:   int32(len(stsReporter.ReadyPods())),
				UnavailableReplicas: int32(len(stsReporter.Pods) - len(stsReporter.ReadyPods())),
			},
		)

		if len(stsReporter.ReadyPods()) >= replicas {
			// All pods are ready (or the desired number of replicas is zero).
			continue
		}

		switch {
		case len(stsReporter.ReadyPods()) == 0:
			availableReason = "NoPodReady"
			availableStatus = monitoringv1.ConditionFalse
		case availableCondition.Status != monitoringv1.ConditionFalse:
			availableReason = "SomePodsNotReady"
			availableStatus = monitoringv1.ConditionDegraded
		}

		for _, p := range stsReporter.Pods {
			if m := p.Message(); m != "" {
				messages = append(messages, fmt.Sprintf("shard %d: pod %s: %s", shard, p.Name, m))
			}
		}
	}

	pStatus.Conditions = operator.UpdateConditions(
		pStatus.Conditions,
		monitoringv1.Condition{
			Type:    monitoringv1.Available,
			Status:  availableStatus,
			Reason:  availableReason,
			Message: strings.Join(messages, "\n"),
			LastTransitionTime: metav1.Time{
				Time: time.Now().UTC(),
			},
			ObservedGeneration: p.GetObjectMeta().GetGeneration(),
		},
		sr.Reconciliations.GetCondition(key, p.GetObjectMeta().GetGeneration()),
	)

	return &pStatus, nil
}

// UpdateServiceMonitorStatus updates the status binding of the serviceMonitor
// for the given workload.
func UpdateServiceMonitorStatus(
	ctx context.Context,
	c *ConfigResourceSyncer,
	res TypedConfigurationResource[*monitoringv1.ServiceMonitor]) error {
	smon := res.resource

	bindings, updated := c.UpdateBindingConditions(smon.Status.Bindings, res.conditions(smon.Generation))
	if !updated {
		return nil
	}
	smon.Status.Bindings = bindings

	_, err := c.mclient.MonitoringV1().ServiceMonitors(smon.Namespace).ApplyStatus(
		ctx,
		ApplyConfigurationFromServiceMonitor(smon),
		metav1.ApplyOptions{FieldManager: operator.PrometheusOperatorFieldManager, Force: true},
	)
	return err
}

// equalConfigResourceConditions returns true when both slices are equal semantically.
func equalConfigResourceConditions(a, b []monitoringv1.ConfigResourceCondition) bool {
	if len(a) != len(b) {
		return false
	}

	ac, bc := slices.Clone(a), slices.Clone(b)

	slices.SortFunc(ac, func(a, b monitoringv1.ConfigResourceCondition) int {
		return cmp.Compare(a.Type, b.Type)
	})
	slices.SortFunc(bc, func(a, b monitoringv1.ConfigResourceCondition) int {
		return cmp.Compare(a.Type, b.Type)
	})

	return slices.EqualFunc(ac, bc, func(a, b monitoringv1.ConfigResourceCondition) bool {
		return a.Type == b.Type &&
			a.Status == b.Status &&
			a.Reason == b.Reason &&
			a.Message == b.Message &&
			a.ObservedGeneration == b.ObservedGeneration
	})
}

// RemoveServiceMonitorBinding removes the Prometheus or PrometheusAgent binding from the status remove the Prometheus or PrometheusAgent binding from the status
// subresource of status in serviceMonitor.
func RemoveServiceMonitorBinding(
	ctx context.Context,
	c *ConfigResourceSyncer,
	smon *monitoringv1.ServiceMonitor) error {
	for i := range smon.Status.Bindings {
		binding := &smon.Status.Bindings[i]
		if binding.Namespace == c.workload.GetNamespace() &&
			binding.Name == c.workload.GetName() &&
			binding.Resource == c.gvr.Resource {
			smon.Status.Bindings = append(smon.Status.Bindings[:i], smon.Status.Bindings[i+1:]...)
			break
		}
	}

	if len(smon.Status.Bindings) == 0 {
		_, err := c.mclient.MonitoringV1().ServiceMonitors(smon.Namespace).UpdateStatus(ctx, smon, metav1.UpdateOptions{FieldManager: operator.PrometheusOperatorFieldManager})
		return err
	}

	_, err := c.mclient.MonitoringV1().ServiceMonitors(smon.Namespace).ApplyStatus(ctx, ApplyConfigurationFromServiceMonitor(smon), metav1.ApplyOptions{FieldManager: operator.PrometheusOperatorFieldManager, Force: true})
	return err
}

func IsBindingPresent(bindings []monitoringv1.WorkloadBinding, p metav1.Object, resource string) bool {
	for _, binding := range bindings {
		if binding.Name == p.GetName() && binding.Namespace == p.GetNamespace() && binding.Resource == resource {
			return true
		}
	}
	return false
}
