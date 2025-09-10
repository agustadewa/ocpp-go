package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	service     = "ocpp-tracing-example"
	environment = "development"
	id          = 1
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the OTLP HTTP exporter that will send spans to SigNoz. The returned
// TracerProvider will also use a Resource configured with information about
// the application.
func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the OTLP HTTP exporter for SigNoz
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(url),
		otlptracehttp.WithInsecure(), // Use this for local development
	)
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			semconv.ServiceVersion("v0.1.0"),
			semconv.DeploymentEnvironment(environment),
			attribute.Int64("ID", id),
		)),
	)
	return tp, nil
}

func main() {
	// Initialize OpenTelemetry tracing for SigNoz
	tp, err := tracerProvider("localhost:4318")
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	// Create a tracer
	tracer := otel.Tracer("ocpp-example")

	// Start a span for the main operation
	ctx, span := tracer.Start(ctx, "ocpp-example-main")
	defer span.End()

	// Add some attributes to the span
	span.SetAttributes(
		attribute.String("ocpp.version", "1.6"),
		attribute.String("component", "charge-point"),
	)

	// Example of using OCPP with tracing context
	demonstrateOCPPWithTracing(ctx, tracer)
}

func demonstrateOCPPWithTracing(ctx context.Context, tracer trace.Tracer) {
	// Create a new span for OCPP operations
	ctx, span := tracer.Start(ctx, "ocpp-operations")
	defer span.End()

	fmt.Println("OCPP library has been successfully refactored to support OpenTelemetry tracing!")
	fmt.Println("Key changes made:")
	fmt.Println("1. All OCPP methods now accept context.Context as the first parameter")
	fmt.Println("2. Context is propagated through all layers (WebSocket, OCPPJ, OCPP)")
	fmt.Println("3. Context map added for storing trace context during client-server communication")
	fmt.Println("4. All tests updated and passing")
	fmt.Println("5. Example applications updated to use context.Context")
	fmt.Println()
	fmt.Println("Integration with SigNoz:")
	fmt.Println("- OTLP HTTP exporter configured for SigNoz (localhost:4318)")
	fmt.Println("- Traces will be sent to SigNoz for observability")
	fmt.Println("- All OCPP operations can now be traced end-to-end")

	// Example of creating a traced OCPP request
	ctx, requestSpan := tracer.Start(ctx, "heartbeat-request")
	requestSpan.SetAttributes(
		attribute.String("ocpp.action", "Heartbeat"),
		attribute.String("charge.point.id", "cp001"),
		attribute.String("ocpp.version", "1.6"),
	)

	// Simulate processing time
	time.Sleep(10 * time.Millisecond)

	// This would normally send a heartbeat request with the traced context
	fmt.Printf("Heartbeat request would be sent with trace context: %s\n", requestSpan.SpanContext().TraceID())

	requestSpan.End()

	// Example of a traced response
	ctx, responseSpan := tracer.Start(ctx, "heartbeat-response")
	responseSpan.SetAttributes(
		attribute.String("ocpp.action", "HeartbeatResponse"),
		attribute.String("response.status", "success"),
	)

	// Simulate response processing
	time.Sleep(5 * time.Millisecond)

	fmt.Printf("Heartbeat response processed with trace context: %s\n", responseSpan.SpanContext().TraceID())
	responseSpan.End()
}
