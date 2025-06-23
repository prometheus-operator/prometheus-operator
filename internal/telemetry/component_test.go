package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestComponentTracer(t *testing.T) {
	// Create a test tracer provider with an in-memory span recorder
	spanRecorder := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))

	// Create a component tracer
	component := "test-component"
	tracer := &componentTracer{
		tracer:    tp.Tracer("test-tracer"),
		component: component,
	}

	// Create a span using our component tracer
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	// Get the recorded spans
	spans := spanRecorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	// Check that the component attribute was added
	recordedSpan := spans[0]
	componentAttr := attribute.String("component", component)
	found := false
	for _, attr := range recordedSpan.Attributes() {
		if attr.Key == componentAttr.Key && attr.Value.AsString() == componentAttr.Value.AsString() {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Component attribute not found in span. Attributes: %v", recordedSpan.Attributes())
	}

	// Also verify the span name is correct
	if recordedSpan.Name() != "test-span" {
		t.Errorf("Expected span name 'test-span', got '%s'", recordedSpan.Name())
	}
}
