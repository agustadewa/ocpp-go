package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/smartcharging"
	"github.com/lorenzodonini/ocpp-go/ocppj"
	"go.opentelemetry.io/otel"
)

func (handler *ChargingStationHandler) OnClearChargingProfile(ctx context.Context, request *smartcharging.ClearChargingProfileRequest) (response *smartcharging.ClearChargingProfileResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnClearChargingProfile.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}

func (handler *ChargingStationHandler) OnGetChargingProfiles(ctx context.Context, request *smartcharging.GetChargingProfilesRequest) (response *smartcharging.GetChargingProfilesResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnGetChargingProfiles.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}

func (handler *ChargingStationHandler) OnGetCompositeSchedule(ctx context.Context, request *smartcharging.GetCompositeScheduleRequest) (response *smartcharging.GetCompositeScheduleResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnGetCompositeSchedule.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}

func (handler *ChargingStationHandler) OnSetChargingProfile(ctx context.Context, request *smartcharging.SetChargingProfileRequest) (response *smartcharging.SetChargingProfileResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnSetChargingProfile.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}
