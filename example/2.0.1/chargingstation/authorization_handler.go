package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/authorization"
)

func (handler *ChargingStationHandler) OnClearCache(ctx context.Context, request *authorization.ClearCacheRequest) (confirmation *authorization.ClearCacheResponse, err error) {
	logDefault(request.GetFeatureName()).Infof("cleared mocked cache")
	return authorization.NewClearCacheResponse(authorization.ClearCacheStatusAccepted), nil
}
