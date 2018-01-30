package v1alpha1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("admissionregistration.k8s.io", "v1alpha1", "initializerconfigurations", false, &InitializerConfiguration{})

	k8s.RegisterList("admissionregistration.k8s.io", "v1alpha1", "initializerconfigurations", false, &InitializerConfigurationList{})
}
