package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/reservation"
)

func (c *CSMSHandler) OnReservationStatusUpdate(ctx context.Context, chargingStationID string, request *reservation.ReservationStatusUpdateRequest) (confirmation *reservation.ReservationStatusUpdateResponse, err error) {
	logDefault(chargingStationID, request.GetFeatureName()).Infof("updated status of reservation %v to: %v", request.ReservationID, request.Status)
	confirmation = reservation.NewReservationStatusUpdateResponse()
	return
}
