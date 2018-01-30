package v1beta1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("events.k8s.io", "v1beta1", "events", true, &Event{})

	k8s.RegisterList("events.k8s.io", "v1beta1", "events", true, &EventList{})
}
