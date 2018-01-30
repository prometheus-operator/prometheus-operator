package v1beta1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("admissionregistration.k8s.io", "v1beta1", "mutatingwebhookconfigurations", false, &MutatingWebhookConfiguration{})
	k8s.Register("admissionregistration.k8s.io", "v1beta1", "validatingwebhookconfigurations", false, &ValidatingWebhookConfiguration{})

	k8s.RegisterList("admissionregistration.k8s.io", "v1beta1", "mutatingwebhookconfigurations", false, &MutatingWebhookConfigurationList{})
	k8s.RegisterList("admissionregistration.k8s.io", "v1beta1", "validatingwebhookconfigurations", false, &ValidatingWebhookConfigurationList{})
}
