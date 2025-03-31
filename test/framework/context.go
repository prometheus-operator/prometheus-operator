// Copyright 2017 The prometheus-operator Authors
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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type TestCtx struct {
	id         string
	namespaces []string
	cleanUpFns []FinalizerFn
}

type FinalizerFn func() error

type diagnosticWriter interface {
	io.Writer
	io.Closer
	StartCollection(string)
}

// stdoutDiagnosticWriter writes collected information to stdout.
type stdoutDiagnosticWriter struct {
	b bytes.Buffer
}

func (sdw *stdoutDiagnosticWriter) Write(b []byte) (int, error) { return sdw.b.Write(b) }
func (sdw *stdoutDiagnosticWriter) Close() error                { return nil }
func (sdw *stdoutDiagnosticWriter) StartCollection(name string) {
	fmt.Fprintf(&sdw.b, "=== %s\n", name)
}

// fileDiagnosticWriter writes collected information to disk.
type fileDiagnosticWriter struct {
	dir string
	f   *os.File
}

func (fdw *fileDiagnosticWriter) Write(b []byte) (int, error) {
	if fdw.f == nil {
		return 0, nil
	}

	return fdw.f.Write(b)
}

func (fdw *fileDiagnosticWriter) Close() error {
	if fdw.f == nil {
		return nil
	}

	return fdw.f.Close()
}

func (fdw *fileDiagnosticWriter) StartCollection(name string) {
	if fdw.f != nil {
		fdw.f.Close()
		fdw.f = nil
	}

	fullpath := filepath.Join(fdw.dir, name)
	if err := os.MkdirAll(filepath.Dir(fullpath), 0755); err != nil {
		return
	}

	f, err := os.Create(fullpath)
	if err != nil {
		return
	}

	fdw.f = f
}

func (f *Framework) NewTestCtx(t *testing.T) *TestCtx {
	// TestCtx is used among others for namespace names where '/' is forbidden
	prefix := strings.TrimPrefix(
		strings.ReplaceAll(
			strings.ToLower(t.Name()),
			"/",
			"-",
		),
		"test",
	)

	tc := &TestCtx{
		id: prefix + "-" + strconv.FormatInt(time.Now().Unix(), 36),
	}

	tc.cleanUpFns = []FinalizerFn{
		func() error {
			t.Helper()
			if !t.Failed() {
				return nil
			}

			// We can collect more information as we see fit over time.
			var (
				dw  diagnosticWriter
				dir = os.Getenv("E2E_DIAGNOSTIC_DIRECTORY")
			)

			if dir != "" {
				dw = &fileDiagnosticWriter{
					dir: filepath.Join(dir, t.Name()),
				}
			} else {
				dw = &stdoutDiagnosticWriter{}
			}
			defer dw.Close()

			dw.StartCollection("alertmanagers")
			tc.collectAlertmanagers(dw, f)
			dw.StartCollection("prometheuses")
			tc.collectPrometheuses(dw, f)
			dw.StartCollection("thanosrulers")
			tc.collectThanosRulers(dw, f)
			dw.StartCollection("prometheusagents")
			tc.collectPrometheusAgents(dw, f)

			tc.collectLogs(dw, f)

			tc.collectEvents(dw, f)

			if sdw, ok := dw.(*stdoutDiagnosticWriter); ok {
				t.Logf("== %s (start)", t.Name())
				t.Log("")
				t.Log(sdw.b.String())
				t.Logf("== %s (end)", t.Name())
			}

			return nil
		},
	}

	return tc
}

func (ctx *TestCtx) collectLogs(dw diagnosticWriter, f *Framework) {
	for _, ns := range ctx.namespaces {
		pods, err := f.KubeClient.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to get pods: %v\n", ns, err)
			continue
		}

		for _, pod := range pods.Items {
			dw.StartCollection(filepath.Join("logs", pod.Namespace, pod.Name))
			err := f.WritePodLogs(context.Background(), dw, ns, pod.Name, LogOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: failed to get pod logs: %v\n", ns, err)
				continue
			}
		}
	}
}

func (ctx *TestCtx) collectEvents(dw diagnosticWriter, f *Framework) {
	for _, ns := range ctx.namespaces {
		dw.StartCollection(filepath.Join("events", ns))
		err := f.WriteEvents(context.Background(), dw, ns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to get events: %v\n", ns, err)
		}
	}
}

func collectConditions(w io.Writer, prefix string, conditions []monitoringv1.Condition) {
	for _, c := range conditions {
		fmt.Fprintf(
			w,
			"%s: condition type=%q status=%q reason=%q message=%q\n",
			prefix,
			c.Type,
			c.Status,
			c.Reason,
			c.Message,
		)
	}
}

func (ctx *TestCtx) collectAlertmanagers(w io.Writer, f *Framework) {
	for _, ns := range ctx.namespaces {
		ams, err := f.MonClientV1.Alertmanagers(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(w, "%s: failed to get alertmanagers: %v\n", ns, err)
			continue
		}

		for _, am := range ams.Items {
			collectConditions(w, fmt.Sprintf("Alertmanager=%s/%s", am.Namespace, am.Name), am.Status.Conditions)
		}
	}
}

func (ctx *TestCtx) collectPrometheuses(w io.Writer, f *Framework) {
	for _, ns := range ctx.namespaces {
		ps, err := f.MonClientV1.Prometheuses(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(w, "%s: failed to get prometheuses: %v\n", ns, err)
			continue
		}

		for _, p := range ps.Items {
			collectConditions(w, fmt.Sprintf("Prometheus=%s/%s", p.Namespace, p.Name), p.Status.Conditions)
		}
	}
}

func (ctx *TestCtx) collectPrometheusAgents(w io.Writer, f *Framework) {
	for _, ns := range ctx.namespaces {
		ps, err := f.MonClientV1alpha1.PrometheusAgents(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(w, "%s: failed to get prometheusagents: %v\n", ns, err)
			continue
		}

		for _, p := range ps.Items {
			collectConditions(w, fmt.Sprintf("PrometheusAgent=%s/%s", p.Namespace, p.Name), p.Status.Conditions)
		}
	}
}

func (ctx *TestCtx) collectThanosRulers(w io.Writer, f *Framework) {
	for _, ns := range ctx.namespaces {
		trs, err := f.MonClientV1.ThanosRulers(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(w, "%s: failed to get thanosrulers: %v\n", ns, err)
			continue
		}

		for _, tr := range trs.Items {
			collectConditions(w, fmt.Sprintf("ThanosRuler=%s/%s", tr.Namespace, tr.Name), tr.Status.Conditions)
		}
	}
}

// ID returns an ascending ID based on the length of cleanUpFns. It is
// based on the premise that every new object also appends a new finalizerFn on
// cleanUpFns. This can e.g. be used to create multiple namespaces in the same
// test context.
func (ctx *TestCtx) ID() string {
	return ctx.id + "-" + strconv.Itoa(len(ctx.cleanUpFns))
}

func (ctx *TestCtx) Cleanup(t *testing.T) {
	t.Helper()
	var eg errgroup.Group

	for i := len(ctx.cleanUpFns) - 1; i >= 0; i-- {
		eg.Go(ctx.cleanUpFns[i])
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}

func (ctx *TestCtx) AddFinalizerFn(fn FinalizerFn) {
	ctx.cleanUpFns = append(ctx.cleanUpFns, fn)
}
