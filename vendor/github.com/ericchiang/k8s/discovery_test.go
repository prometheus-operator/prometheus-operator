package k8s_test

import (
	"context"
	"testing"

	"github.com/ericchiang/k8s"
)

func TestDiscovery(t *testing.T) {
	client := k8s.NewDiscoveryClient(newTestClient(t))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := client.Version(ctx); err != nil {
		t.Errorf("list version: %v", err)
	}

	if _, err := client.APIGroups(ctx); err != nil {
		t.Errorf("list api groups: %v", err)
	}

	if _, err := client.APIGroup(ctx, "extensions"); err != nil {
		t.Errorf("list api group: %v", err)
	}

	if _, err := client.APIResources(ctx, "extensions", "v1beta1"); err != nil {
		t.Errorf("list api group resources: %v", err)
	}
}
