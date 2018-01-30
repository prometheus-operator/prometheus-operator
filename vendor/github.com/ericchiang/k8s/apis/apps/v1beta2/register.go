package v1beta2

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("apps", "v1beta2", "controllerrevisions", true, &ControllerRevision{})
	k8s.Register("apps", "v1beta2", "daemonsets", true, &DaemonSet{})
	k8s.Register("apps", "v1beta2", "deployments", true, &Deployment{})
	k8s.Register("apps", "v1beta2", "replicasets", true, &ReplicaSet{})
	k8s.Register("apps", "v1beta2", "statefulsets", true, &StatefulSet{})

	k8s.RegisterList("apps", "v1beta2", "controllerrevisions", true, &ControllerRevisionList{})
	k8s.RegisterList("apps", "v1beta2", "daemonsets", true, &DaemonSetList{})
	k8s.RegisterList("apps", "v1beta2", "deployments", true, &DeploymentList{})
	k8s.RegisterList("apps", "v1beta2", "replicasets", true, &ReplicaSetList{})
	k8s.RegisterList("apps", "v1beta2", "statefulsets", true, &StatefulSetList{})
}
