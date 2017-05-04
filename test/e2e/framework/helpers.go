package framework

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/rest"

	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/pkg/errors"
)

func PathToOSFile(relativPath string) (*os.File, error) {
	path, err := filepath.Abs(relativPath)
	if err != nil {
		return nil, err
	}

	manifest, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

// WaitForPodsReady waits for a selection of Pods to be running and each
// container to pass its readiness check.
func WaitForPodsReady(kubeClient kubernetes.Interface, namespace string, timeout time.Duration, expectedReplicas int, opts v1.ListOptions) error {
	return wait.Poll(time.Second, timeout, func() (bool, error) {
		pl, err := kubeClient.Core().Pods(namespace).List(opts)
		if err != nil {
			return false, err
		}

		runningAndReady := 0
		for _, p := range pl.Items {
			isRunningAndReady, err := k8sutil.PodRunningAndReady(p)
			if err != nil {
				return false, err
			}

			if isRunningAndReady {
				runningAndReady++
			}
		}

		if runningAndReady == expectedReplicas {
			return true, nil
		}
		return false, nil
	})
}

func WaitForPodsRunImage(kubeClient kubernetes.Interface, namespace string, expectedReplicas int, image string, opts v1.ListOptions) error {
	return wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		pl, err := kubeClient.Core().Pods(namespace).List(opts)
		if err != nil {
			return false, err
		}

		runningImage := 0
		for _, p := range pl.Items {
			if podRunsImage(p, image) {
				runningImage++
			}
		}

		if runningImage == expectedReplicas {
			return true, nil
		}
		return false, nil
	})
}

func WaitForHTTPSuccessStatusCode(timeout time.Duration, url string) error {
	var resp *http.Response
	err := wait.Poll(time.Second, timeout, func() (bool, error) {
		var err error
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			return true, nil
		}
		return false, nil
	})

	return errors.Wrap(err, fmt.Sprintf(
		"waiting for %v to return a successfull status code timed out. Last response from server was: %v",
		url,
		resp,
	))
}

func podRunsImage(p v1.Pod, image string) bool {
	for _, c := range p.Spec.Containers {
		if image == c.Image {
			return true
		}
	}

	return false
}

func GetLogs(kubeClient kubernetes.Interface, namespace string, podName, containerName string) (string, error) {
	logs, err := kubeClient.Core().RESTClient().Get().
		Resource("pods").
		Namespace(namespace).
		Name(podName).SubResource("log").
		Param("container", containerName).
		Do().
		Raw()
	if err != nil {
		return "", err
	}
	return string(logs), err
}

func (f *Framework) Poll(timeout, pollInterval time.Duration, pollFunc func() (bool, error)) error {
	t := time.After(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t:
			return fmt.Errorf("timed out")
		case <-ticker.C:
			b, err := pollFunc()
			if err != nil {
				return err
			}
			if b {
				return nil
			}
		}
	}
}

func ProxyGetPod(kubeClient kubernetes.Interface, namespace string, podName string, port string, path string) *rest.Request {
	return kubeClient.CoreV1().RESTClient().Get().Prefix("proxy").Namespace(namespace).Resource("pods").Name(podName + ":" + port).Suffix(path)
}
