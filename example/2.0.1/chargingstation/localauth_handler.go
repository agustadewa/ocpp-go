package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/localauth"
	"go.opentelemetry.io/otel"
)

func (handler *ChargingStationHandler) OnGetLocalListVersion(ctx context.Context, request *localauth.GetLocalListVersionRequest) (response *localauth.GetLocalListVersionResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnGetLocalListVersion.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Infof("returning current local list version: %v", handler.localAuthListVersion)
	return localauth.NewGetLocalListVersionResponse(handler.localAuthListVersion), nil
}

func (handler *ChargingStationHandler) OnSendLocalList(ctx context.Context, request *localauth.SendLocalListRequest) (response *localauth.SendLocalListResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnSendLocalList.handler")
	defer span.End()
	if request.VersionNumber <= handler.localAuthListVersion {
		logDefault(request.GetFeatureName()).
			Errorf("requested listVersion %v is lower/equal than the current list version %v", request.VersionNumber, handler.localAuthListVersion)
		return localauth.NewSendLocalListResponse(localauth.SendLocalListStatusVersionMismatch), nil
	}
	if request.UpdateType == localauth.UpdateTypeFull {
		handler.localAuthList = request.LocalAuthorizationList
		handler.localAuthListVersion = request.VersionNumber
	} else if request.UpdateType == localauth.UpdateTypeDifferential {
		handler.localAuthList = append(handler.localAuthList, request.LocalAuthorizationList...)
		handler.localAuthListVersion = request.VersionNumber
	}
	logDefault(request.GetFeatureName()).Errorf("accepted new local authorization list %v, %v",
		request.VersionNumber, request.UpdateType)
	return localauth.NewSendLocalListResponse(localauth.SendLocalListStatusAccepted), nil
}
