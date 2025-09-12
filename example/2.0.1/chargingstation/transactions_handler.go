package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/transactions"
	"github.com/lorenzodonini/ocpp-go/ocppj"
	"go.opentelemetry.io/otel"
)

func (handler *ChargingStationHandler) OnGetTransactionStatus(ctx context.Context, request *transactions.GetTransactionStatusRequest) (response *transactions.GetTransactionStatusResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnGetTransactionStatus.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}
