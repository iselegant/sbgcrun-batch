package tracing

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer はOpenTelemetryトレーサーを初期化します。
// useCloudTrace=true の場合はCloud Traceエクスポーターを使用し、
// false の場合はスパンを記録しないNoopエクスポーターを使用します。
// 返り値のshutdown関数は、アプリケーション終了時に呼び出してください。
func InitTracer(serviceName string, useCloudTrace bool) (shutdown func(context.Context) error, err error) {
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	}

	if useCloudTrace {
		exporter, err := newCloudTraceExporter()
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloud Trace exporter: %w", err)
		}
		opts = append(opts, sdktrace.WithBatcher(exporter))
		log.Printf("Cloud Trace tracing initialized for service: %s", serviceName)
	} else {
		log.Printf("Tracing initialized with noop exporter for service: %s", serviceName)
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}
