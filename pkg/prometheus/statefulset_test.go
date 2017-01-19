package prometheus

import (
	"reflect"
	"testing"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

func TestStatefulSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	sset := makeStatefulSet(v1alpha1.Prometheus{
		ObjectMeta: v1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, nil)

	if !reflect.DeepEqual(labels, sset.Labels) || !reflect.DeepEqual(annotations, sset.Annotations) {
		t.Fatal("Labels or Annotations are not properly being propagated to the StatefulSet")
	}
}
