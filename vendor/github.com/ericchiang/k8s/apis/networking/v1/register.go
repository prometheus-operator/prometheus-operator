package v1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("networking.k8s.io", "v1", "networkpolicies", true, &NetworkPolicy{})

	k8s.RegisterList("networking.k8s.io", "v1", "networkpolicies", true, &NetworkPolicyList{})
}
