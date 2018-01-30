package v1alpha1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("scheduling.k8s.io", "v1alpha1", "priorityclasss", false, &PriorityClass{})

	k8s.RegisterList("scheduling.k8s.io", "v1alpha1", "priorityclasss", false, &PriorityClassList{})
}
