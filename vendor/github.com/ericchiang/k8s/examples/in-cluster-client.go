// +build ignore

package nodes

import (
	"context"
	"fmt"
	"log"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
)

func listNodes() {
	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}

	var nodes corev1.NodeList
	if err := client.List(context.Background(), "", &nodes); err != nil {
		log.Fatal(err)
	}
	for _, node := range nodes.Items {
		fmt.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
	}
}
