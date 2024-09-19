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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

func (f *Framework) NewTestCtx(t *testing.T) *TestCtx {
	// TestCtx is used among others for namespace names where '/' is forbidden
	prefix := strings.TrimPrefix(
		strings.Replace(
			strings.ToLower(t.Name()),
			"/",
			"-",
			-1,
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
			b := &bytes.Buffer{}
			tc.collectAlertmanagers(b, f)
			tc.collectPrometheuses(b, f)
			tc.collectThanosRulers(b, f)
			tc.collectLogs(b, f)
			tc.collectEvents(b, f)

			t.Logf("=== %s (start)", t.Name())
			t.Log("")
			t.Log(b.String())
			t.Logf("=== %s (end)", t.Name())

			return nil
		},
	}

	return tc
}

func (ctx *TestCtx) collectLogs(w io.Writer, f *Framework) {
	for _, ns := range ctx.namespaces {
		pods, err := f.KubeClient.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(w, "%s: failed to get pods: %v\n", ns, err)
			continue
		}

		for _, pod := range pods.Items {
			err := f.WritePodLogs(context.Background(), w, ns, pod.Name, LogOptions{})
			if err != nil {
				fmt.Fprintf(w, "%s: failed to get pod logs: %v\n", ns, err)
				continue
			}
		}
	}
}

func (ctx *TestCtx) collectEvents(w io.Writer, f *Framework) {
	fmt.Fprintln(w, "=== Events")
	for _, ns := range ctx.namespaces {
		b := &bytes.Buffer{}
		err := f.WriteEvents(context.Background(), b, ns)
		if err != nil {
			fmt.Fprintf(w, "%s: failed to get events: %v\n", ns, err)
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
	fmt.Fprintln(w, "=== Alertmanagers")
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
	fmt.Fprintln(w, "=== Prometheuses")
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

func (ctx *TestCtx) collectThanosRulers(w io.Writer, f *Framework) {
	fmt.Fprintln(w, "=== ThanosRulers")
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
	var eg errgroup.Group

	for i := len(ctx.cleanUpFns) - 1; i >= 0; i-- {
		eg.Go(ctx.cleanUpFns[i])
	}

	err := eg.Wait()
	require.NoError(t, err)
}

func (ctx *TestCtx) AddFinalizerFn(fn FinalizerFn) {
	ctx.cleanUpFns = append(ctx.cleanUpFns, fn)
}
