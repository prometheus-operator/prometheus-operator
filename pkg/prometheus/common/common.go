package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	Kclient kubernetes.Interface
	Logger  log.Logger

	Key      string
	Instance monitoringv1.PrometheusInterface

	Reconciliations *operator.ReconciliationTracker
	SsetInfs        *informers.ForResource
	Rr              *operator.ResourceReconciler
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

	for shard := range prompkg.ExpectedStatefulSetShardNames(config.Instance) {
		ssetName := prompkg.KeyToStatefulSetKey(config.Instance, config.Key, shard)
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

	// Compute the Reconciled ConditionType.
	reconciledCondition := monitoringv1.Condition{
		Type:   monitoringv1.Reconciled,
		Status: monitoringv1.ConditionTrue,
		LastTransitionTime: metav1.Time{
			Time: time.Now().UTC(),
		},
		ObservedGeneration: config.Instance.GetObjectMeta().GetGeneration(),
	}
	reconciliationStatus, found := config.Reconciliations.GetStatus(config.Key)
	if !found {
		reconciledCondition.Status = monitoringv1.ConditionUnknown
		reconciledCondition.Reason = "NotFound"
		reconciledCondition.Message = fmt.Sprintf("object %q not found", config.Key)
	} else {
		if !reconciliationStatus.Ok() {
			reconciledCondition.Status = monitoringv1.ConditionFalse
		}
		reconciledCondition.Reason = reconciliationStatus.Reason()
		reconciledCondition.Message = reconciliationStatus.Message()
	}

	// Update the last transition times only if the status of the available condition has changed.
	for _, condition := range config.Instance.GetStatus().Conditions {
		if condition.Type == availableCondition.Type && condition.Status == availableCondition.Status {
			availableCondition.LastTransitionTime = condition.LastTransitionTime
			continue
		}

		if condition.Type == reconciledCondition.Type && condition.Status == reconciledCondition.Status {
			reconciledCondition.LastTransitionTime = condition.LastTransitionTime
		}
	}

	pStatus.Conditions = append(pStatus.Conditions, availableCondition, reconciledCondition)

	return &pStatus, nil
}
