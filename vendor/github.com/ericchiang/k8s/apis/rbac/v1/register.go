package v1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("rbac.authorization.k8s.io", "v1", "clusterroles", false, &ClusterRole{})
	k8s.Register("rbac.authorization.k8s.io", "v1", "clusterrolebindings", false, &ClusterRoleBinding{})
	k8s.Register("rbac.authorization.k8s.io", "v1", "roles", true, &Role{})
	k8s.Register("rbac.authorization.k8s.io", "v1", "rolebindings", true, &RoleBinding{})

	k8s.RegisterList("rbac.authorization.k8s.io", "v1", "clusterroles", false, &ClusterRoleList{})
	k8s.RegisterList("rbac.authorization.k8s.io", "v1", "clusterrolebindings", false, &ClusterRoleBindingList{})
	k8s.RegisterList("rbac.authorization.k8s.io", "v1", "roles", true, &RoleList{})
	k8s.RegisterList("rbac.authorization.k8s.io", "v1", "rolebindings", true, &RoleBindingList{})
}
