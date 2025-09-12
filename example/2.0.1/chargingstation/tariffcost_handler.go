package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/tariffcost"
	"go.opentelemetry.io/otel"
)

func (handler *ChargingStationHandler) OnCostUpdated(ctx context.Context, request *tariffcost.CostUpdatedRequest) (confirmation *tariffcost.CostUpdatedResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnCostUpdated.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Infof("accepted request to display cost for transaction %v: %v", request.TransactionID, request.TotalCost)
	// TODO: update internal display to show updated cost for transaction
	return tariffcost.NewCostUpdatedResponse(), nil
}
