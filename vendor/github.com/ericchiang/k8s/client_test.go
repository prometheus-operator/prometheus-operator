package k8s

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/ericchiang/k8s/api/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

const skipMsg = `
warning: this package's test run using the default context of your "kubectl" command,
and will create resources on your cluster (mostly configmaps).

If you wish to continue set the following environment variable:

	export K8S_CLIENT_TEST=1

To suppress this message, set:

	export K8S_CLIENT_TEST=0
`

func newTestClient(t *testing.T) *Client {
	if os.Getenv("K8S_CLIENT_TEST") == "0" {
		t.Skip("")
	}
	if os.Getenv("K8S_CLIENT_TEST") != "1" {
		t.Skip(skipMsg)
	}

	cmd := exec.Command("kubectl", "config", "view", "-o", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("'kubectl config view -o json': %v %s", err, out)
	}

	config := new(Config)
	if err := json.Unmarshal(out, config); err != nil {
		t.Fatalf("parse kubeconfig: %v '%s'", err, out)
	}
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client
}

func newName() string {
	b := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func TestNewTestClient(t *testing.T) {
	newTestClient(t)
}

func TestHTTP2(t *testing.T) {
	client := newTestClient(t)
	req, err := http.NewRequest("GET", client.urlForPath("/api"), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !strings.HasPrefix(resp.Proto, "HTTP/2") {
		t.Errorf("expected proto=HTTP/2.X, got=", resp.Proto)
	}
}

func TestListNodes(t *testing.T) {
	client := newTestClient(t)
	if _, err := client.CoreV1().ListNodes(context.Background()); err != nil {
		t.Fatal("failed to list nodes: %v", err)
	}
}

func TestConfigMaps(t *testing.T) {
	client := newTestClient(t).CoreV1()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	name := newName()
	labelVal := newName()

	cm := &v1.ConfigMap{
		Metadata: &metav1.ObjectMeta{
			Name:      String(name),
			Namespace: String("default"),
			Labels: map[string]string{
				"testLabel": labelVal,
			},
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}
	got, err := client.CreateConfigMap(ctx, cm)
	if err != nil {
		t.Fatalf("create config map: %v", err)
	}
	got.Data["zam"] = "spam"
	_, err = client.UpdateConfigMap(ctx, got)
	if err != nil {
		t.Fatalf("update config map: %v", err)
	}

	tests := []struct {
		labelVal string
		expNum   int
	}{
		{labelVal, 1},
		{newName(), 0},
	}
	for _, test := range tests {
		l := new(LabelSelector)
		l.Eq("testLabel", test.labelVal)

		configMaps, err := client.ListConfigMaps(ctx, "default", l.Selector())
		if err != nil {
			t.Errorf("failed to list configmaps: %v", err)
			continue
		}
		got := len(configMaps.Items)
		if got != test.expNum {
			t.Errorf("expected selector to return %d items got %d", test.expNum, got)
		}
	}

	if err := client.DeleteConfigMap(ctx, *cm.Metadata.Name, *cm.Metadata.Namespace); err != nil {
		t.Fatalf("delete config map: %v", err)
	}

}

func TestWatch(t *testing.T) {
	client := newTestClient(t).CoreV1()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w, err := client.WatchConfigMaps(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	name := newName()
	labelVal := newName()

	cm := &v1.ConfigMap{
		Metadata: &metav1.ObjectMeta{
			Name:      String(name),
			Namespace: String("default"),
			Labels: map[string]string{
				"testLabel": labelVal,
			},
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}
	got, err := client.CreateConfigMap(ctx, cm)
	if err != nil {
		t.Fatalf("create config map: %v", err)
	}

	if event, gotFromWatch, err := w.Next(); err != nil {
		t.Errorf("failed to get next watch: %v", err)
	} else {
		if *event.Type != EventAdded {
			t.Errorf("expected event type %q got %q", EventAdded, *event.Type)
		}
		if !reflect.DeepEqual(got, gotFromWatch) {
			t.Errorf("object from add event did not match expected value")
		}
	}

	got.Data["zam"] = "spam"
	got, err = client.UpdateConfigMap(ctx, got)
	if err != nil {
		t.Fatalf("update config map: %v", err)
	}

	if event, gotFromWatch, err := w.Next(); err != nil {
		t.Errorf("failed to get next watch: %v", err)
	} else {
		if *event.Type != EventModified {
			t.Errorf("expected event type %q got %q", EventModified, *event.Type)
		}
		if !reflect.DeepEqual(got, gotFromWatch) {
			t.Errorf("object from modified event did not match expected value")
		}
	}

	tests := []struct {
		labelVal string
		expNum   int
	}{
		{labelVal, 1},
		{newName(), 0},
	}
	for _, test := range tests {
		l := new(LabelSelector)
		l.Eq("testLabel", test.labelVal)

		configMaps, err := client.ListConfigMaps(ctx, "default", l.Selector())
		if err != nil {
			t.Errorf("failed to list configmaps: %v", err)
			continue
		}
		got := len(configMaps.Items)
		if got != test.expNum {
			t.Errorf("expected selector to return %d items got %d", test.expNum, got)
		}
	}

	if err := client.DeleteConfigMap(ctx, *cm.Metadata.Name, *cm.Metadata.Namespace); err != nil {
		t.Fatalf("delete config map: %v", err)
	}
	if event, gotFromWatch, err := w.Next(); err != nil {
		t.Errorf("failed to get next watch: %v", err)
	} else {
		if *event.Type != EventDeleted {
			t.Errorf("expected event type %q got %q", EventDeleted, *event.Type)
		}

		// Resource version will be different after a delete
		got.Metadata.ResourceVersion = String("")
		gotFromWatch.Metadata.ResourceVersion = String("")

		if !reflect.DeepEqual(got, gotFromWatch) {
			t.Errorf("object from deleted event did not match expected value")
		}
	}
}

// TestWatchNamespace ensures that creating a configmap in a non-default namespace is not returned while watching the default namespace
func TestWatchNamespace(t *testing.T) {
	client := newTestClient(t).CoreV1()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defaultWatch, err := client.WatchConfigMaps(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	defer defaultWatch.Close()

	allWatch, err := client.WatchConfigMaps(ctx, AllNamespaces)
	if err != nil {
		t.Fatal(err)
	}
	defer allWatch.Close()

	nonDefaultNamespaceName := newName()
	defaultName := newName()
	name := newName()
	labelVal := newName()

	// Create a configmap in the default namespace so the "default" watch has something to return
	defaultCM := &v1.ConfigMap{
		Metadata: &metav1.ObjectMeta{
			Name:      String(defaultName),
			Namespace: String("default"),
			Labels: map[string]string{
				"testLabel": labelVal,
			},
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}
	defaultGot, err := client.CreateConfigMap(ctx, defaultCM)
	if err != nil {
		t.Fatalf("create config map: %v", err)
	}

	// Create a non-default Namespace
	ns := &v1.Namespace{
		Metadata: &metav1.ObjectMeta{
			Name: String(nonDefaultNamespaceName),
		},
	}
	if _, err := client.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("create non-default-namespace: %v", err)
	}

	// Create a configmap in the non-default namespace
	nonDefaultCM := &v1.ConfigMap{
		Metadata: &metav1.ObjectMeta{
			Name:      String(name),
			Namespace: String(nonDefaultNamespaceName),
			Labels: map[string]string{
				"testLabel": labelVal,
			},
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}
	nonDefaultGot, err := client.CreateConfigMap(ctx, nonDefaultCM)
	if err != nil {
		t.Fatalf("create config map: %v", err)
	}

	// Watching the default namespace should not return the non-default namespace configmap,
	// and instead return the previously created configmap in the default namespace
	if _, gotFromWatch, err := defaultWatch.Next(); err != nil {
		t.Errorf("failed to get next watch: %v", err)
	} else {
		if reflect.DeepEqual(nonDefaultGot, gotFromWatch) {
			t.Errorf("config map in non-default namespace returned while watching default namespace")
		}
		if !reflect.DeepEqual(defaultGot, gotFromWatch) {
			t.Errorf("object from add event did not match expected value")
		}
	}

	// However, watching all-namespaces should contain both the default and non-default namespaced configmaps
	if _, gotFromWatch, err := allWatch.Next(); err != nil {
		t.Errorf("failed to get next watch: %v", err)
	} else {
		if !reflect.DeepEqual(defaultGot, gotFromWatch) {
			t.Errorf("watching all namespaces did not return the expected configmap")
		}
	}

	if _, gotFromWatch, err := allWatch.Next(); err != nil {
		t.Errorf("failed to get next watch: %v", err)
	} else {
		if !reflect.DeepEqual(nonDefaultGot, gotFromWatch) {
			t.Errorf("watching all namespaces did not return the expected configmap")
		}
	}

	// Delete the config map in the default namespace first, then delete the non-default namespace config map.
	// Only the former should be noticed by the default-watch.

	if err := client.DeleteConfigMap(ctx, *defaultCM.Metadata.Name, *defaultCM.Metadata.Namespace); err != nil {
		t.Fatalf("delete config map: %v", err)
	}
	if err := client.DeleteConfigMap(ctx, *nonDefaultCM.Metadata.Name, *nonDefaultCM.Metadata.Namespace); err != nil {
		t.Fatalf("delete config map: %v", err)
	}

	if event, gotFromWatch, err := defaultWatch.Next(); err != nil {
		t.Errorf("failed to get next watch: %v", err)
	} else {
		if *event.Type != EventDeleted {
			t.Errorf("expected event type %q got %q", EventDeleted, *event.Type)
		}

		// Resource version will be different after a delete
		nonDefaultGot.Metadata.ResourceVersion = String("")
		gotFromWatch.Metadata.ResourceVersion = String("")

		if reflect.DeepEqual(nonDefaultGot, gotFromWatch) {
			t.Errorf("should not have received event from non-default namespace while watching default namespace")
		}
	}

	if err := client.DeleteNamespace(ctx, nonDefaultNamespaceName); err != nil {
		t.Fatalf("delete namespace: %v", err)
	}
}

func TestDefaultNamespace(t *testing.T) {
	c := &Config{
		Clusters: []NamedCluster{
			{
				Name: "local",
				Cluster: Cluster{
					Server: "http://localhost:8080",
				},
			},
		},
		AuthInfos: []NamedAuthInfo{
			{
				Name: "local",
			},
		},
	}
	cli, err := NewClient(c)
	if err != nil {
		t.Fatal(err)
	}
	if cli.Namespace != "default" {
		t.Errorf("expected namespace=%q got=%q", "default", cli.Namespace)
	}
}
