// Copyright 2025 The prometheus-operator Authors
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

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"k8s.io/client-go/rest"
)

// Telemetry holds the telemetry providers and shutdown functions.
type Telemetry struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	shutdown       []func(context.Context) error
}

// Shutdown gracefully shuts down all telemetry providers.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	for _, fn := range t.shutdown {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Setup initializes OpenTelemetry with autoexport exporters.
// It configures both tracing and metrics based on environment variables.
// See https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
func Setup(ctx context.Context, serviceName, serviceVersion string, logger *slog.Logger) (*Telemetry, error) {
	// Check if OTEL is disabled
	if os.Getenv("OTEL_SDK_DISABLED") == "true" {
		logger.Info("OpenTelemetry is disabled via OTEL_SDK_DISABLED")
		return &Telemetry{
			TracerProvider: tracenoop.NewTracerProvider(),
			MeterProvider:  noop.NewMeterProvider(),
		}, nil
	}

	// Create resource describing this application
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
		resource.WithFromEnv(),   // Merge with environment variables
		resource.WithProcess(),   // Add process attributes
		resource.WithOS(),        // Add OS attributes
		resource.WithContainer(), // Add container attributes if running in container
		resource.WithHost(),      // Add host attributes
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTelemetry resource: %w", err)
	}

	tel := &Telemetry{}

	// Setup trace provider with autoexport
	traceExporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	if traceExporter != nil && !autoexport.IsNoneSpanExporter(traceExporter) {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(traceExporter),
		)
		tel.TracerProvider = tp
		tel.shutdown = append(tel.shutdown, tp.Shutdown)

		// Set global tracer provider
		otel.SetTracerProvider(tp)

		logger.Info("OpenTelemetry tracing initialized", "service", serviceName, "version", serviceVersion)
	} else {
		logger.Info("OpenTelemetry tracing not configured - no exporter found")
		tel.TracerProvider = tracenoop.NewTracerProvider()
	}

	// Setup metric provider with autoexport
	metricReader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric reader: %w", err)
	}

	if metricReader != nil && !autoexport.IsNoneMetricReader(metricReader) {
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(metricReader),
		)
		tel.MeterProvider = mp
		tel.shutdown = append(tel.shutdown, mp.Shutdown)

		// Set global meter provider
		otel.SetMeterProvider(mp)

		logger.Info("OpenTelemetry metrics initialized", "service", serviceName, "version", serviceVersion)
	} else {
		logger.Info("OpenTelemetry metrics not configured - no exporter found")
		tel.MeterProvider = noop.NewMeterProvider()
	}

	return tel, nil
}

// WrapHTTPHandler wraps an HTTP handler with OpenTelemetry instrumentation.
// It adds automatic tracing and metrics for HTTP requests.
func WrapHTTPHandler(handler http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}

// WrapHTTPMux wraps an HTTP mux with OpenTelemetry instrumentation.
// It adds automatic tracing and metrics for all routes in the mux.
// Operation names will be in the format "METHOD /path" for better observability.
func WrapHTTPMux(mux *http.ServeMux) http.Handler {
	return otelhttp.NewHandler(mux, "",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}

// WrapRoundTripper wraps a Kubernetes client's RoundTripper with OpenTelemetry instrumentation.
// This provides automatic tracing for all Kubernetes API calls with meaningful operation names.
func WrapRoundTripper(rt http.RoundTripper, name string) http.RoundTripper {
	return otelhttp.NewTransport(rt,
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			// Create meaningful span names for Kubernetes API calls
			// Examples: "GET /api/v1/prometheuses", "PUT /api/v1/namespaces/default/statefulsets/prometheus-test"
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}

// InstrumentKubernetesConfig adds OpenTelemetry instrumentation to a Kubernetes rest.Config.
// This will trace all API calls made using clients created from this config.
func InstrumentKubernetesConfig(config *rest.Config, serviceName string) {
	// Wrap the existing transport with OTEL instrumentation
	config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return WrapRoundTripper(rt, serviceName)
	})
}

// StartSpan starts a new tracing span with the given name and returns the span and a context
// containing the span. The span should be ended by calling span.End().
func StartSpan(ctx context.Context, tracer trace.Tracer, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}

// RecordError records an error in the given span and sets the span status to error.
func RecordError(span trace.Span, err error, msg string) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, msg)
	}
}

// AddSpanAttributes adds attributes to the given span.
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// GetTracer returns a tracer for the given name.
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// GetMeter returns a meter for the given name.
func GetMeter(name string) metric.Meter {
	return otel.Meter(name)
}
