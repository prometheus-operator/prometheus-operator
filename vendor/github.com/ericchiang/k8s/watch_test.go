package k8s_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

// configMapJSON is used to test the JSON serialization watch.
type configMapJSON struct {
	Metadata *metav1.ObjectMeta `json:"metadata"`
	Data     map[string]string  `json:"data"`
}

func (c *configMapJSON) GetMetadata() *metav1.ObjectMeta {
	return c.Metadata
}

func init() {
	k8s.Register("", "v1", "configmaps", true, &configMapJSON{})
}

func testWatch(t *testing.T, client *k8s.Client, namespace string, newCM func() k8s.Resource, update func(cm k8s.Resource)) {
	w, err := client.Watch(context.TODO(), namespace, newCM())
	if err != nil {
		t.Errorf("watch configmaps: %v", err)
	}
	defer w.Close()

	cm := newCM()
	want := func(eventType string) {
		got := newCM()
		eT, err := w.Next(got)
		if err != nil {
			t.Errorf("decode watch event: %v", err)
			return
		}
		if eT != eventType {
			t.Errorf("expected event type %q got %q", eventType, eT)
		}
		if !reflect.DeepEqual(got, cm) {
			t.Errorf("configmaps did not match\nwant=%#v\ngot=%#v", cm, got)
		}
	}

	if err := client.Create(context.TODO(), cm); err != nil {
		t.Errorf("create configmap: %v", err)
		return
	}
	want(k8s.EventAdded)

	update(cm)

	if err := client.Update(context.TODO(), cm); err != nil {
		t.Errorf("update configmap: %v", err)
		return
	}
	want(k8s.EventModified)

	if err := client.Delete(context.TODO(), cm); err != nil {
		t.Errorf("Delete configmap: %v", err)
		return
	}
	want(k8s.EventDeleted)
}

func TestWatchConfigMapJSON(t *testing.T) {
	withNamespace(t, func(client *k8s.Client, namespace string) {
		newCM := func() k8s.Resource {
			return &configMapJSON{
				Metadata: &metav1.ObjectMeta{
					Name:      k8s.String("my-configmap"),
					Namespace: &namespace,
				},
			}
		}

		updateCM := func(cm k8s.Resource) {
			(cm.(*configMapJSON)).Data = map[string]string{"hello": "world"}
		}
		testWatch(t, client, namespace, newCM, updateCM)
	})
}

func TestWatchConfigMapProto(t *testing.T) {
	withNamespace(t, func(client *k8s.Client, namespace string) {
		newCM := func() k8s.Resource {
			return &corev1.ConfigMap{
				Metadata: &metav1.ObjectMeta{
					Name:      k8s.String("my-configmap"),
					Namespace: &namespace,
				},
			}
		}

		updateCM := func(cm k8s.Resource) {
			(cm.(*corev1.ConfigMap)).Data = map[string]string{"hello": "world"}
		}
		testWatch(t, client, namespace, newCM, updateCM)
	})
}
