# OCPP Incoming Call Tracing

This document explains how distributed tracing has been implemented for incoming OCPP requests (one-way operations where you receive a request from a peer and send back a response).

## Overview

The tracing implementation adds OpenTelemetry spans to track incoming OCPP requests at two levels:

1. **OCPP Layer Tracing**: Automatic tracing in the OCPP library itself
2. **Handler Level Tracing**: Optional detailed tracing in your request handlers

## OCPP Layer Tracing (Automatic)

### Charging Station Side

When your charging station receives an incoming request from the CSMS (like `ChangeAvailability`, `Reset`, etc.), the OCPP library automatically creates a span:

- **Span Name**: `server.handle_request`
- **Tracer**: `ocpp-charging-station`
- **Attributes**:
  - `ocpp.action`: The OCPP action name (e.g., "ChangeAvailability")
  - `ocpp.request_id`: The unique request ID
  - `ocpp.direction`: "incoming"
  - `ocpp.role`: "charging_station"

### CSMS Side

When your CSMS receives an incoming request from a charging station (like `BootNotification`, `Heartbeat`, etc.), the OCPP library automatically creates a span:

- **Span Name**: `csms.handle_request`
- **Tracer**: `ocpp-csms`
- **Attributes**:
  - `ocpp.action`: The OCPP action name (e.g., "BootNotification")
  - `ocpp.request_id`: The unique request ID
  - `ocpp.direction`: "incoming"
  - `ocpp.role`: "csms"
  - `ocpp.charging_station_id`: The ID of the charging station

## Handler Level Tracing (Optional)

You can add more detailed tracing within your individual request handlers. Here's an example from the `OnChangeAvailability` handler:

```go
func (handler *ChargingStationHandler) OnChangeAvailability(ctx context.Context, request *availability.ChangeAvailabilityRequest) (response *availability.ChangeAvailabilityResponse, err error) {
    // Create a child span for detailed handler tracing
    tracer := otel.Tracer("ocpp-charging-station-handler")
    ctx, span := tracer.Start(ctx, "handler.change_availability")
    defer span.End()
    
    // Add span attributes for detailed tracing
    span.SetAttributes(
        attribute.String("ocpp.operational_status", string(request.OperationalStatus)),
    )
    if request.Evse != nil {
        span.SetAttributes(
            attribute.Int("ocpp.evse_id", request.Evse.ID),
        )
        if request.Evse.ConnectorID != nil {
            span.SetAttributes(
                attribute.Int("ocpp.connector_id", *request.Evse.ConnectorID),
            )
        }
    }
    
    // Add events to track the flow
    if request.Evse == nil {
        span.AddEvent("changing_availability_for_entire_station")
        // ... handler logic
    } else {
        span.AddEvent("changing_availability_for_evse")
        // ... handler logic
    }
    
    return response, err
}
```

## Trace Structure

With both levels of tracing enabled, you'll see traces like this in Grafana Tempo:

```
client.example_routine (your application span)
└── client.send_request (outgoing request)
    └── server.handle_request (peer's incoming request handling)
        └── handler.change_availability (detailed handler logic)
```

## Error Handling

The tracing implementation automatically:

- Records errors on spans when handlers return errors
- Sets appropriate span status (Ok/Error)
- Adds error details to span attributes
- Handles unsupported actions and missing handlers

## Context Propagation

The traced context is properly propagated:

1. From the OCPP layer to your handlers
2. Through goroutines (for CSMS async processing)
3. To any child spans you create in your handlers

## Benefits

This tracing implementation provides:

1. **Complete visibility** into incoming request processing
2. **Error tracking** for failed request handling
3. **Performance monitoring** of handler execution times
4. **Request correlation** between outgoing and incoming requests
5. **Detailed debugging** information for troubleshooting

## Usage in Grafana Tempo

To view incoming request traces:

1. Search by service name: `ocpp-charging-station` or `ocpp-csms`
2. Filter by span name: `server.handle_request` or `csms.handle_request`
3. Look for traces with `ocpp.direction=incoming`
4. Examine the span hierarchy to see handler details

The traces will show the complete flow from when the request is received until the response is sent back to the peer.
