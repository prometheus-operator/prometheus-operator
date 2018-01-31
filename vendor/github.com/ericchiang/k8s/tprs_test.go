package k8s

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ericchiang/k8s/api/unversioned"
	"github.com/ericchiang/k8s/apis/extensions/v1beta1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

func TestTPRs(t *testing.T) {
	client := newTestClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type Metric struct {
		*unversioned.TypeMeta `json:",inline"`
		*metav1.ObjectMeta    `json:"metadata,omitempty"`

		Timestamp time.Time `json:"timestamp"`
		Value     int64     `json:"value"`
	}

	type MetricsList struct {
		*unversioned.TypeMeta `json:",inline"`
		*unversioned.ListMeta `json:"metadata,omitempty"`

		Items []Metric `json:"items"`
	}

	// Create a ThirdPartyResource
	tpr := &v1beta1.ThirdPartyResource{
		Metadata: &metav1.ObjectMeta{
			Name: String("metric.example.com"),
		},
		Description: String("A value and a timestamp"),
		Versions: []*v1beta1.APIVersion{
			{Name: String("v1")},
		},
	}
	_, err := client.ExtensionsV1Beta1().CreateThirdPartyResource(ctx, tpr)
	if err != nil {
		if apiErr, ok := err.(*APIError); !ok || apiErr.Code != http.StatusConflict {
			t.Fatalf("create third party resource: %v", err)
		}
	}

	metric := &Metric{
		ObjectMeta: &metav1.ObjectMeta{
			Name: String("foo"),
		},
		Timestamp: time.Now(),
		Value:     42,
	}

	myClient := client.ThirdPartyResources("example.com", "v1")
	err = myClient.Create(ctx, "metrics", "default", metric, metric)
	if err != nil {
		t.Errorf("create third party resource: %v", err)
	}

	var metrics MetricsList
	if err := myClient.List(ctx, "metrics", "default", &metrics); err != nil {
		t.Errorf("list third party resource: %v", err)
	}

	if err := myClient.Delete(ctx, "metrics", "default", "foo"); err != nil {
		t.Fatalf("delete tpr: %v", err)
	}
}
