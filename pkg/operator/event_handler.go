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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// EventPayload represents the object(s) being notified.
//
// It contains both the raw object and the object's metadata to avoid redundant
// type-casting.
//
// The Old and OldMeta fields are only set for update events.
type EventPayload struct {
	EventType   EventType
	Old         interface{}
	OldMeta     metav1.Object
	Current     interface{}
	CurrentMeta metav1.Object
}

// EventType reports whether the event is a creation, update or deletion.
type EventType int

const (
	// EventTypeOnAdd represents a creation event.
	EventTypeOnAdd EventType = iota

	// EventTypeOnUpdate represents an update event.
	EventTypeOnUpdate

	// EventTypeOnDelete represents a deletion event.
	EventTypeOnDelete
)

// FilterFunc is a function that gets an EventPayload and returns true if the
// object should trigger a reconciliation.
type FilterFunc func(EventPayload) bool

// AnyFilter returns a FilterFunc which calls the filters sequentially and
// returns true as soon as one filter returns true.
func AnyFilter(filters ...FilterFunc) FilterFunc {
	return func(ep EventPayload) bool {
		for _, filter := range filters {
			if filter(ep) {
				return true
			}
		}

		return false
	}
}

// ResourceVersionChanged returns true if the old and current objects don't
// have the same resource version.
//
// It always returns true for creation and deletion events.
func ResourceVersionChanged(ep EventPayload) bool {
	if ep.EventType != EventTypeOnUpdate {
		return true
	}

	return ep.OldMeta.GetResourceVersion() != ep.CurrentMeta.GetResourceVersion()
}

// GenerationChanged returns true if the old and current objects don't have the
// same generation.
//
// It always returns true for creation and deletion events.
func GenerationChanged(ep EventPayload) bool {
	if ep.EventType != EventTypeOnUpdate {
		return true
	}

	return ep.OldMeta.GetGeneration() != ep.CurrentMeta.GetGeneration()
}

// LabelsChanged returns true if the old and current objects don't have the
// same labels.
//
// It always returns true for creation and deletion events.
func LabelsChanged(ep EventPayload) bool {
	if ep.EventType != EventTypeOnUpdate {
		return true
	}

	return !reflect.DeepEqual(ep.OldMeta.GetLabels(), ep.CurrentMeta.GetLabels())
}

// EventHandlerOption allows to configure the event handler.
type EventHandlerOption func(*EventHandler)

// WithFilter adds a filter function to the event handler.
func WithFilter(filter FilterFunc) EventHandlerOption {
	return func(e *EventHandler) {
		e.filterFuncs = append(e.filterFuncs, filter)
	}
}

// EventHandler implements the k8s.io/tools/cache.ResourceEventHandler interface.
type EventHandler struct {
	logger   *slog.Logger
	accessor *Accessor
	metrics  *Metrics

	objName     string
	enqueueFunc func(string)
	filterFuncs []FilterFunc
}

var _ = cache.ResourceEventHandler(&EventHandler{})

// NewEventHandler returns a new event handler.
func NewEventHandler(
	logger *slog.Logger,
	accessor *Accessor,
	metrics *Metrics,
	objName string,
	enqueueFunc func(ns string),
	options ...EventHandlerOption,
) *EventHandler {
	e := &EventHandler{
		logger:      logger,
		accessor:    accessor,
		metrics:     metrics,
		objName:     objName,
		enqueueFunc: enqueueFunc,
	}

	for _, opt := range options {
		opt(e)
	}

	return e
}

// OnAdd implements the k8s.io/tools/cache.ResourceEventHandler interface.
func (e *EventHandler) OnAdd(obj interface{}, _ bool) {
	o, ok := e.accessor.ObjectMetadata(obj)
	if !ok {
		return
	}

	for _, fn := range e.filterFuncs {
		if !fn(EventPayload{
			EventType:   EventTypeOnAdd,
			Current:     obj,
			CurrentMeta: o,
		}) {
			return
		}
	}

	e.recordEvent(AddEvent, o)
	e.enqueueFunc(o.GetNamespace())
}

// OnUpdate implements the k8s.io/tools/cache.ResourceEventHandler interface.
func (e *EventHandler) OnUpdate(old, cur interface{}) {
	oldMeta, ok := e.accessor.ObjectMetadata(old)
	if !ok {
		return
	}

	curMeta, ok := e.accessor.ObjectMetadata(cur)
	if !ok {
		return
	}

	for _, fn := range e.filterFuncs {
		if !fn(EventPayload{
			EventType:   EventTypeOnUpdate,
			Old:         old,
			OldMeta:     oldMeta,
			Current:     cur,
			CurrentMeta: curMeta,
		}) {
			return
		}
	}

	e.recordEvent(UpdateEvent, curMeta)
	e.enqueueFunc(curMeta.GetNamespace())
}

// OnDelete implements the k8s.io/tools/cache.ResourceEventHandler interface.
func (e *EventHandler) OnDelete(obj interface{}) {
	o, ok := e.accessor.ObjectMetadata(obj)
	if !ok {
		return
	}

	for _, fn := range e.filterFuncs {
		if !fn(EventPayload{
			EventType:   EventTypeOnDelete,
			Current:     obj,
			CurrentMeta: o,
		}) {
			return
		}
	}

	e.recordEvent(DeleteEvent, o)
	e.enqueueFunc(o.GetNamespace())
}

func (e *EventHandler) recordEvent(event HandlerEvent, obj metav1.Object) {
	eventName := string(event)
	if strings.HasSuffix(eventName, "e") {
		eventName += "d"
	} else {
		eventName += "ed"
	}

	e.logger.Debug(
		fmt.Sprintf("%s %s", e.objName, eventName),
		strings.ToLower(e.objName), fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
	)
	e.metrics.TriggerByCounter(e.objName, event).Inc()
}
