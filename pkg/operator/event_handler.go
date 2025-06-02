// Copyright 2023 The prometheus-operator Authors
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
	"log/slog"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventHandler implements the k8s.io/tools/cache.ResourceEventHandler interface.
type EventHandler struct {
	logger   *slog.Logger
	accessor *Accessor
	metrics  *Metrics

	objName     string
	enqueueFunc func(string)
}

func NewEventHandler(
	logger *slog.Logger,
	accessor *Accessor,
	metrics *Metrics,
	objName string,
	enqueueFunc func(ns string),
) *EventHandler {
	return &EventHandler{
		logger:      logger,
		accessor:    accessor,
		metrics:     metrics,
		objName:     objName,
		enqueueFunc: enqueueFunc,
	}
}

func (e *EventHandler) OnAdd(obj interface{}, _ bool) {
	o, ok := e.accessor.ObjectMetadata(obj)
	if ok {
		e.logger.Debug(fmt.Sprintf("%s added", e.objName))
		e.metrics.TriggerByCounter(e.objName, AddEvent).Inc()
		e.enqueueFunc(o.GetNamespace())
	}
}

func (e *EventHandler) OnUpdate(old, cur interface{}) {
	oldSpec := getSpec(old)
	curSpec := getSpec(cur)

	oldMeta, ok1 := old.(metav1.Object)
	curMeta, ok2 := cur.(metav1.Object)

	if !ok1 || !ok2 {
		// cannot compare metadata if casting fails
		return
	}
	// Compare Spec
	specEqual := reflect.DeepEqual(oldSpec, curSpec)

	// TODO(yp969803): also deleteTimestamp and generation ?
	// Compare Metadata fields
	labelsEqual := reflect.DeepEqual(oldMeta.GetLabels(), curMeta.GetLabels())
	annotationsEqual := reflect.DeepEqual(oldMeta.GetAnnotations(), curMeta.GetAnnotations())
	finalizersEqual := reflect.DeepEqual(oldMeta.GetFinalizers(), curMeta.GetFinalizers())

	if specEqual && labelsEqual && annotationsEqual && finalizersEqual {
		return
	}
	if o, ok := e.accessor.ObjectMetadata(cur); ok {
		e.logger.Debug(fmt.Sprintf("%s updated", e.objName))
		e.metrics.TriggerByCounter(e.objName, UpdateEvent)
		e.enqueueFunc(o.GetNamespace())
	}
}

func (e *EventHandler) OnDelete(obj interface{}) {
	if o, ok := e.accessor.ObjectMetadata(obj); ok {
		e.logger.Debug(fmt.Sprintf("%s deleted", e.objName))
		e.metrics.TriggerByCounter(e.objName, DeleteEvent).Inc()
		e.enqueueFunc(o.GetNamespace())
	}
}

func getSpec(obj interface{}) interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	specField := v.FieldByName("Spec")
	if !specField.IsValid() {
		return nil
	}
	return specField.Interface()
}
