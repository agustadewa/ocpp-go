# OCPP OpenTelemetry Tracing Example

This example demonstrates how to use the refactored OCPP library with OpenTelemetry tracing and SigNoz for
observability.

## Overview

The OCPP library has been refactored to support `context.Context` throughout all layers:

1. **WebSocket Layer**: All WebSocket operations now accept and propagate context
2. **OCPPJ Layer**: JSON message handling includes context propagation
3. **OCPP Protocol Layer**: All OCPP 1.6 and 2.0.1 methods accept context as the first parameter

## Key Features

- **Context Propagation**: Trace context flows through all OCPP operations
- **Context Map**: Stores trace context during client-server communication
- **SigNoz Integration**: OTLP HTTP exporter configured for SigNoz observability
- **End-to-End Tracing**: Complete visibility from WebSocket to OCPP protocol level

## Running the Example

1. **Start SigNoz** (if you have it running locally):
   ```bash
   # SigNoz should be running on localhost:4318 for OTLP HTTP
   ```

2. **Run the tracing example**:
   ```bash
   cd example/tracing
   go run main.go
   ```

## Configuration

The example is configured to send traces to SigNoz at `localhost:4318`. You can modify the endpoint in the `main()`
function:

```go
tp, err := tracerProvider("localhost:4318")
```

For production use, you might want to:

- Use secure connections (remove `WithInsecure()`)
- Configure authentication headers
- Set up proper resource attributes
- Configure sampling rates

## Integration in Your Application

To use tracing in your OCPP application:

1. **Initialize OpenTelemetry**:
   ```go
   tp, err := tracerProvider("your-signoz-endpoint:4318")
   otel.SetTracerProvider(tp)
   ```

2. **Create traced contexts**:
   ```go
   tracer := otel.Tracer("your-app")
   ctx, span := tracer.Start(context.Background(), "ocpp-operation")
   defer span.End()
   ```

3. **Use context in OCPP calls**:
   ```go
   // All OCPP methods now accept context as first parameter
   response, err := chargePoint.Heartbeat(ctx, heartbeatRequest)
   ```

## Trace Attributes

The library automatically adds relevant attributes to spans:

- `ocpp.action`: The OCPP action being performed
- `ocpp.version`: OCPP protocol version (1.6 or 2.0.1)
- `charge.point.id`: Charge point identifier
- `message.id`: OCPP message ID
- `websocket.url`: WebSocket connection URL

## Benefits

- **Distributed Tracing**: Track requests across charge points and central systems
- **Performance Monitoring**: Identify bottlenecks in OCPP operations
- **Error Tracking**: Correlate errors with specific trace contexts
- **Debugging**: Follow the complete flow of OCPP messages
- **Observability**: Gain insights into system behavior and performance
