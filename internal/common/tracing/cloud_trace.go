package tracing

import (
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func newCloudTraceExporter() (sdktrace.SpanExporter, error) {
	return texporter.New()
}
