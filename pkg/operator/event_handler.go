package operator

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// todo: add event handler factory

type EventHandler struct {
	logger   log.Logger
	accessor *Accessor

	metrics                    *Metrics
	objName                    string
	enqueueForMonitorNamespace func(ns string)
}

func NewEventHandler(
	logger log.Logger,
	accessor *Accessor,
	metrics *Metrics,
	objName string,
	enqueueForMonitorNamespace func(ns string),
) *EventHandler {
	return &EventHandler{
		logger:                     logger,
		accessor:                   accessor,
		metrics:                    metrics,
		objName:                    objName,
		enqueueForMonitorNamespace: enqueueForMonitorNamespace,
	}
}

func (e *EventHandler) OnAdd(obj interface{}, isInInitialList bool) {
	o, ok := e.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(e.logger).Log("msg", fmt.Sprintf("%s added", e.objName))
		e.metrics.TriggerByCounter(e.objName, AddEvent).Inc()
		e.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

func (e *EventHandler) OnUpdate(old, cur interface{}) {
	if old.(metav1.Object).GetResourceVersion() == cur.(metav1.Object).GetResourceVersion() {
		return
	}

	if o, ok := e.accessor.ObjectMetadata(cur); ok {
		level.Debug(e.logger).Log("msg", fmt.Sprintf("%s updated", e.objName))
		e.metrics.TriggerByCounter(e.objName, UpdateEvent)
		e.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

func (e *EventHandler) OnDelete(obj interface{}) {
	if o, ok := e.accessor.ObjectMetadata(obj); ok {
		level.Debug(e.logger).Log("msg", fmt.Sprintf("%s deleted", e.objName))
		e.metrics.TriggerByCounter(e.objName, DeleteEvent).Inc()
		e.enqueueForMonitorNamespace(o.GetNamespace())
	}
}
