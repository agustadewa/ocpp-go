package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/authorization"
	"go.opentelemetry.io/otel"
)

func (handler *ChargingStationHandler) OnClearCache(ctx context.Context, request *authorization.ClearCacheRequest) (confirmation *authorization.ClearCacheResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnClearCache.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Infof("cleared mocked cache")
	span.AddEvent("cache cleared")
	return authorization.NewClearCacheResponse(authorization.ClearCacheStatusAccepted), nil
}
