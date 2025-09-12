package main

import (
	"context"
	"fmt"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/availability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (handler *ChargingStationHandler) OnChangeAvailability(ctx context.Context, request *availability.ChangeAvailabilityRequest) (response *availability.ChangeAvailabilityResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnChangeAvailability.handler")
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

	if request.Evse == nil {
		// Changing availability for the entire charging station
		span.AddEvent(fmt.Sprintf("changing availability for entire station into: %s", request.OperationalStatus))
		handler.availability = request.OperationalStatus
		// TODO: recursively update the availability for all evse/connectors
		response = availability.NewChangeAvailabilityResponse(availability.ChangeAvailabilityStatusAccepted)
		return
	}
	reqEvse := request.Evse
	if e, ok := handler.evse[reqEvse.ID]; ok {
		// Changing availability for a specific EVSE
		span.AddEvent("changing_availability_for_evse")
		if reqEvse.ConnectorID != nil {
			// Changing availability for a specific connector
			span.AddEvent("changing_availability_for_connector")
			if !e.hasConnector(*reqEvse.ConnectorID) {
				span.AddEvent("connector_not_found")
				response = availability.NewChangeAvailabilityResponse(availability.ChangeAvailabilityStatusRejected)
			} else {
				connector := e.connectors[*reqEvse.ConnectorID]
				connector.availability = request.OperationalStatus
				e.connectors[*reqEvse.ConnectorID] = connector
				span.AddEvent("connector_availability_changed")
				response = availability.NewChangeAvailabilityResponse(availability.ChangeAvailabilityStatusAccepted)
			}
			return
		}
		e.availability = request.OperationalStatus
		span.AddEvent("evse_availability_changed")
		response = availability.NewChangeAvailabilityResponse(availability.ChangeAvailabilityStatusAccepted)
		return
	}
	// No EVSE with such ID found
	span.AddEvent("evse_not_found")
	response = availability.NewChangeAvailabilityResponse(availability.ChangeAvailabilityStatusRejected)
	return
}
