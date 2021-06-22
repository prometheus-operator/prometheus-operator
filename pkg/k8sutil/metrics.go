// Copyright 2020 The prometheus-operator Authors
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

package k8sutil

import (
	"context"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/metrics"
)

type clientGoHTTPMetricAdapter struct {
	count    *prometheus.CounterVec
	duration *prometheus.SummaryVec
}

type clientGoRateLimiterMetricAdapter struct {
	duration *prometheus.SummaryVec
}

// MustRegisterClientGoMetrics registers k8s.io/client-go metrics.
// It panics if it encounters an error (e.g. metrics already registered).
func MustRegisterClientGoMetrics(registerer prometheus.Registerer) {
	httpMetrics := &clientGoHTTPMetricAdapter{
		count: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prometheus_operator_kubernetes_client_http_requests_total",
				Help: "Total number of Kubernetes's client requests by status code.",
			},
			[]string{"status_code"},
		),
		duration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "prometheus_operator_kubernetes_client_http_request_duration_seconds",
				Help:       "Summary of latencies for the Kubernetes client's requests by endpoint.",
				Objectives: map[float64]float64{},
			},
			[]string{"endpoint"},
		),
	}

	rateLimiterMetrics := &clientGoRateLimiterMetricAdapter{
		duration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "prometheus_operator_kubernetes_client_rate_limiter_duration_seconds",
				Help:       "Summary of latencies for the Kuberntes client's rate limiter by endpoint.",
				Objectives: map[float64]float64{},
			},
			[]string{"endpoint"},
		),
	}

	metrics.Register(
		metrics.RegisterOpts{
			RequestLatency:     httpMetrics,
			RequestResult:      httpMetrics,
			RateLimiterLatency: rateLimiterMetrics,
		},
	)

	registerer.MustRegister(httpMetrics.count, httpMetrics.duration, rateLimiterMetrics.duration)
}

func (a *clientGoHTTPMetricAdapter) Increment(_ context.Context, code string, method string, host string) {
	a.count.WithLabelValues(code).Inc()
}

func (a *clientGoHTTPMetricAdapter) Observe(_ context.Context, verb string, u url.URL, latency time.Duration) {
	a.duration.WithLabelValues(u.EscapedPath()).Observe(latency.Seconds())
}

func (a *clientGoRateLimiterMetricAdapter) Observe(_ context.Context, verb string, u url.URL, latency time.Duration) {
	a.duration.WithLabelValues(u.EscapedPath()).Observe(latency.Seconds())
}
