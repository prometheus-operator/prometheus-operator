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
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var prometheusKeyInShardStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)-shard-[1-9][0-9]*$")
var prometheusKeyInStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)$")

type Config struct {
	Kclient kubernetes.Interface
	Logger  log.Logger

	Key      string
	Instance monitoringv1.PrometheusInterface

	Reconciliations *operator.ReconciliationTracker
	SsetInfs        *informers.ForResource
	Rr              *operator.ResourceReconciler
}

func StatefulSetKeyToPrometheusKey(key string) (bool, string) {
	r := prometheusKeyInStatefulSet
	if prometheusKeyInShardStatefulSet.MatchString(key) {
		r = prometheusKeyInShardStatefulSet
	}

	matches := r.FindAllStringSubmatch(key, 2)
	if len(matches) != 1 {
		return false, ""
	}
	if len(matches[0]) != 3 {
		return false, ""
	}
	return true, matches[0][1] + "/" + matches[0][2]
}

func KeyToStatefulSetKey(p monitoringv1.PrometheusInterface, key string, shard int) string {
	keyParts := strings.Split(key, "/")
	return fmt.Sprintf("%s/%s", keyParts[0], statefulSetNameFromPrometheusName(p, keyParts[1], shard))
}

func statefulSetNameFromPrometheusName(p monitoringv1.PrometheusInterface, name string, shard int) string {
	if shard == 0 {
		return fmt.Sprintf("%s-%s", prefix(p), name)
	}
	return fmt.Sprintf("%s-%s-shard-%d", prefix(p), name, shard)
}

func NewTLSAssetSecret(p monitoringv1.PrometheusInterface, labels map[string]string) *v1.Secret {
	objMeta := p.GetObjectMeta()
	typeMeta := p.GetTypeMeta()

	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   TLSAssetsSecretName(p),
			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         typeMeta.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               typeMeta.Kind,
					Name:               objMeta.GetName(),
					UID:                objMeta.GetUID(),
				},
			},
		},
		Data: map[string][]byte{},
	}
}

// ValidateRemoteWriteSpec checks that mutually exclusive configurations are not
// included in the Prometheus remoteWrite configuration section.
// Reference:
// https://github.com/prometheus/prometheus/blob/main/docs/configuration/configuration.md#remote_write
func ValidateRemoteWriteSpec(spec monitoringv1.RemoteWriteSpec) error {
	var nonNilFields []string
	for k, v := range map[string]interface{}{
		"basicAuth":     spec.BasicAuth,
		"oauth2":        spec.OAuth2,
		"authorization": spec.Authorization,
		"sigv4":         spec.Sigv4,
	} {
		if reflect.ValueOf(v).IsNil() {
			continue
		}
		nonNilFields = append(nonNilFields, fmt.Sprintf("%q", k))
	}

	if len(nonNilFields) > 1 {
		return errors.Errorf("%s can't be set at the same time, at most one of them must be defined", strings.Join(nonNilFields, " and "))
	}

	return nil
}

// GetStatus is a helper function that retrieves the status subresource of the object identified by the given
// key.
func GetStatus(ctx context.Context, config Config) (*monitoringv1.PrometheusStatus, error) {

	commonFields := config.Instance.GetCommonPrometheusFields()
	pStatus := monitoringv1.PrometheusStatus{
		Paused: commonFields.Paused,
	}

	logger := log.With(config.Logger, "key", config.Key)
	level.Info(logger).Log("msg", "update prometheus status")

	var (
		availableCondition = monitoringv1.Condition{
			Type:   monitoringv1.Available,
			Status: monitoringv1.ConditionTrue,
			LastTransitionTime: metav1.Time{
				Time: time.Now().UTC(),
			},
			ObservedGeneration: config.Instance.GetObjectMeta().GetGeneration(),
		}
		messages []string
		replicas = 1
	)

	if commonFields.Replicas != nil {
		replicas = int(*commonFields.Replicas)
	}

	for shard := range ExpectedStatefulSetShardNames(config.Instance) {
		ssetName := KeyToStatefulSetKey(config.Instance, config.Key, shard)
		logger := log.With(logger, "statefulset", ssetName, "shard", shard)

		obj, err := config.SsetInfs.Get(ssetName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Object not yet in the store or already deleted.
				level.Info(logger).Log("msg", "not found")
				continue
			}
			return nil, errors.Wrap(err, "failed to retrieve statefulset")
		}

		sset := obj.(*appsv1.StatefulSet)
		if config.Rr.DeletionInProgress(sset) {
			continue
		}

		stsReporter, err := operator.NewStatefulSetReporter(ctx, config.Kclient, sset)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve statefulset state")
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

		if len(stsReporter.ReadyPods()) == 0 {
			availableCondition.Reason = "NoPodReady"
			availableCondition.Status = monitoringv1.ConditionFalse
		} else if availableCondition.Status != monitoringv1.ConditionFalse {
			availableCondition.Reason = "SomePodsNotReady"
			availableCondition.Status = monitoringv1.ConditionDegraded
		}

		for _, p := range stsReporter.Pods {
			if m := p.Message(); m != "" {
				messages = append(messages, fmt.Sprintf("shard %d: pod %s: %s", shard, p.Name, m))
			}
		}
	}

	availableCondition.Message = strings.Join(messages, "\n")

	reconciledCondition := config.Reconciliations.GetCondition(config.Key, config.Instance.GetObjectMeta().GetGeneration())
	pStatus.Conditions = operator.UpdateConditions(pStatus.Conditions, availableCondition, reconciledCondition)

	return &pStatus, nil
}
