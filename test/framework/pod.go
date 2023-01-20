// Copyright 2016 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

// PrintPodLogs prints the logs of a specified Pod
func (f *Framework) PrintPodLogs(ctx context.Context, ns, p string) error {
	pod, err := f.KubeClient.CoreV1().Pods(ns).Get(ctx, p, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to print logs of pod '%v': failed to get pod", p)
	}

	for _, c := range pod.Spec.Containers {
		req := f.KubeClient.CoreV1().Pods(ns).GetLogs(p, &v1.PodLogOptions{Container: c.Name})
		resp, err := req.DoRaw(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve logs of pod '%v'", p)
		}

		fmt.Printf("=== Logs of %v/%v/%v:", ns, p, c.Name)
		fmt.Println(string(resp))
	}

	return nil
}

// GetPodRestartCount returns a map of container names and their restart counts for
// a given pod.
func (f *Framework) GetPodRestartCount(ctx context.Context, ns, podName string) (map[string]int32, error) {
	pod, err := f.KubeClient.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve pod to get restart count")
	}

	restarts := map[string]int32{}

	for _, status := range pod.Status.ContainerStatuses {
		restarts[status.Name] = status.RestartCount
	}

	return restarts, nil
}

// ExecOptions passed to ExecWithOptions
type ExecOptions struct {
	Command       []string
	Namespace     string
	PodName       string
	ContainerName string
	Stdin         io.Reader
	CaptureStdout bool
	CaptureStderr bool
	// If false, whitespace in std{err,out} will be removed.
	PreserveWhitespace bool
}

// ExecWithOptions executes a command in the specified container, returning
// stdout, stderr and error. `options` allowed for additional parameters to be
// passed. Inspired by
// https://github.com/kubernetes/kubernetes/blob/dde6e8e7465468c32642659cb708a5cc922add64/test/e2e/framework/exec_util.go#L36-L51
func (f *Framework) ExecWithOptions(ctx context.Context, options ExecOptions) (string, string, error) {
	const tty = false

	req := f.KubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(options.PodName).
		Namespace(options.Namespace).
		SubResource("exec").
		Param("container", options.ContainerName)
	req.VersionedParams(&v1.PodExecOptions{
		Container: options.ContainerName,
		Command:   options.Command,
		Stdin:     options.Stdin != nil,
		Stdout:    options.CaptureStdout,
		Stderr:    options.CaptureStderr,
		TTY:       tty,
	}, kscheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	err := execute(ctx, "POST", req.URL(), f.RestConfig, options.Stdin, &stdout, &stderr, tty)

	if options.PreserveWhitespace {
		return stdout.String(), stderr.String(), err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func execute(ctx context.Context, method string, url *url.URL, config *rest.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}

// StartPortForward initiates a port forwarding connection to a pod on the
// localhost interface. It returns a closer function that should be invoked to
// stop the proxy server.
// The function blocks until the port forwarding proxy server is ready to
// receive connections or the context is canceled.
func StartPortForward(ctx context.Context, config *rest.Config, scheme string, name string, ns string, port string) (func(), error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", ns, name)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{Scheme: scheme, Path: path, Host: hostIP}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)
	forwarder, err := portforward.New(dialer, []string{port}, stopChan, readyChan, out, errOut)
	if err != nil {
		return nil, err
	}

	forwardErr := make(chan error, 1)
	go func() {
		if err := forwarder.ForwardPorts(); err != nil {
			forwardErr <- err
		}
		close(forwardErr)
	}()

	select {
	case <-readyChan:
		return func() { close(stopChan) }, nil
	case <-ctx.Done():
		var err error
		select {
		case err = <-forwardErr:
		default:
		}
		return nil, fmt.Errorf("%v: %v", ctx.Err(), err)
	}
}
