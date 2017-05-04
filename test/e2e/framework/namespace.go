package framework

import (
	"fmt"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/pkg/errors"
)

func CreateNamespace(kubeClient kubernetes.Interface, name string) (*v1.Namespace, error) {
	namespace, err := kubeClient.Core().Namespaces().Create(&v1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create namespace with name %v", name))
	}
	return namespace, nil
}

func (ctx *TestCtx) CreateNamespace(t *testing.T, kubeClient kubernetes.Interface) string {
	name := ctx.GetObjID()
	if _, err := CreateNamespace(kubeClient, name); err != nil {
		t.Fatal(err)
	}

	namespaceFinalizerFn := func() error {
		if err := DeleteNamespace(kubeClient, name); err != nil {
			return err
		}
		return nil
	}

	ctx.AddFinalizerFn(namespaceFinalizerFn)

	return name
}

func DeleteNamespace(kubeClient kubernetes.Interface, name string) error {
	return kubeClient.Core().Namespaces().Delete(name, nil)
}
