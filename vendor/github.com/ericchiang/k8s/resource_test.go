package k8s

import (
	"testing"

	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

// Redefine types since all API groups import "github.com/ericchiang/k8s"
// We can't use them here because it'll create a circular import cycle.

type Pod struct {
	Metadata *metav1.ObjectMeta
}

type PodList struct {
	Metadata *metav1.ListMeta
}

func (p *Pod) GetMetadata() *metav1.ObjectMeta   { return p.Metadata }
func (p *PodList) GetMetadata() *metav1.ListMeta { return p.Metadata }

type Deployment struct {
	Metadata *metav1.ObjectMeta
}

type DeploymentList struct {
	Metadata *metav1.ListMeta
}

func (p *Deployment) GetMetadata() *metav1.ObjectMeta   { return p.Metadata }
func (p *DeploymentList) GetMetadata() *metav1.ListMeta { return p.Metadata }

type ClusterRole struct {
	Metadata *metav1.ObjectMeta
}

type ClusterRoleList struct {
	Metadata *metav1.ListMeta
}

func (p *ClusterRole) GetMetadata() *metav1.ObjectMeta   { return p.Metadata }
func (p *ClusterRoleList) GetMetadata() *metav1.ListMeta { return p.Metadata }

func init() {
	Register("", "v1", "pods", true, &Pod{})
	RegisterList("", "v1", "pods", true, &PodList{})

	Register("apps", "v1beta2", "deployments", true, &Deployment{})
	RegisterList("apps", "v1beta2", "deployments", true, &DeploymentList{})

	Register("rbac.authorization.k8s.io", "v1", "clusterroles", false, &ClusterRole{})
	RegisterList("rbac.authorization.k8s.io", "v1", "clusterroles", false, &ClusterRoleList{})
}

func TestResourceURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		resource Resource
		withName bool
		options  []Option
		want     string
		wantErr  bool
	}{
		{
			name:     "pod",
			endpoint: "https://example.com",
			resource: &Pod{
				Metadata: &metav1.ObjectMeta{
					Namespace: String("my-namespace"),
					Name:      String("my-pod"),
				},
			},
			want: "https://example.com/api/v1/namespaces/my-namespace/pods",
		},
		{
			name:     "deployment",
			endpoint: "https://example.com",
			resource: &Deployment{
				Metadata: &metav1.ObjectMeta{
					Namespace: String("my-namespace"),
					Name:      String("my-deployment"),
				},
			},
			want: "https://example.com/apis/apps/v1beta2/namespaces/my-namespace/deployments",
		},
		{
			name:     "deployment-with-name",
			endpoint: "https://example.com",
			resource: &Deployment{
				Metadata: &metav1.ObjectMeta{
					Namespace: String("my-namespace"),
					Name:      String("my-deployment"),
				},
			},
			withName: true,
			want:     "https://example.com/apis/apps/v1beta2/namespaces/my-namespace/deployments/my-deployment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resourceURL(test.endpoint, test.resource, test.withName, test.options...)
			if err != nil {
				if test.wantErr {
					return
				}
				t.Fatalf("constructing resource URL: %v", err)
			}
			if test.wantErr {
				t.Fatal("expected error")
			}
			if test.want != got {
				t.Errorf("wanted=%q, got=%q", test.want, got)
			}
		})
	}
}
