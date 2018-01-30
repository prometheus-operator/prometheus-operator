// +build ignore

package configmaps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ericchiang/k8s"
	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

// createConfigMap creates a configmap in the client's default namespace
// but does not return an error if a configmap of the same name already
// exists.
func createConfigMap(client *k8s.Client, name string, values map[string]string) error {
	cm := &v1.ConfigMap{
		Metadata: &metav1.ObjectMeta{
			Name:      &name,
			Namespace: &client.Namespace,
		},
		Data: values,
	}

	err := client.Create(context.Background(), cm)

	// If an HTTP error was returned by the API server, it will be of type
	// *k8s.APIError. This can be used to inspect the status code.
	if apiErr, ok := err.(*k8s.APIError); ok {
		// Resource already exists. Carry on.
		if apiErr.Code == http.StatusConflict {
			return nil
		}
	}
	return fmt.Errorf("create configmap: %v", err)
}
