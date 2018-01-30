// +build ignore

package customresources

import (
	"context"
	"fmt"

	"github.com/ericchiang/k8s"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

func init() {
	k8s.Register("resource.example.com", "v1", "myresource", true, &MyResource{})
	k8s.RegisterList("resource.example.com", "v1", "myresource", true, &MyResourceList{})
}

type MyResource struct {
	Metadata *metav1.ObjectMeta `json:"metadata"`
	Foo      string             `json:"foo"`
	Bar      int                `json:"bar"`
}

func (m *MyResource) GetMetadata() *metav1.ObjectMeta {
	return m.Metadata
}

type MyResourceList struct {
	Metadata *metav1.ListMeta `json:"metadata"`
	Items    []MyResource     `json:"items"`
}

func (m *MyResourceList) GetMetadata() *metav1.ListMeta {
	return m.Metadata
}

func do(ctx context.Context, client *k8s.Client, namespace string) error {
	r := &MyResource{
		Metadata: &metav1.ObjectMeta{
			Name:      k8s.String("my-custom-resource"),
			Namespace: &namespace,
		},
		Foo: "hello, world!",
		Bar: 42,
	}
	if err := client.Create(ctx, r); err != nil {
		return fmt.Errorf("create: %v", err)
	}
	r.Bar = -8
	if err := client.Update(ctx, r); err != nil {
		return fmt.Errorf("update: %v", err)
	}
	if err := client.Delete(ctx, r); err != nil {
		return fmt.Errorf("delete: %v", err)
	}
	return nil
}
