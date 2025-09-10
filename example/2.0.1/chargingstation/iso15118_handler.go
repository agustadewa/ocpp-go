package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/iso15118"
	"github.com/lorenzodonini/ocpp-go/ocppj"
)

func (handler *ChargingStationHandler) OnDeleteCertificate(ctx context.Context, request *iso15118.DeleteCertificateRequest) (response *iso15118.DeleteCertificateResponse, err error) {
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}

func (handler *ChargingStationHandler) OnGetInstalledCertificateIds(ctx context.Context, request *iso15118.GetInstalledCertificateIdsRequest) (response *iso15118.GetInstalledCertificateIdsResponse, err error) {
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}

func (handler *ChargingStationHandler) OnInstallCertificate(ctx context.Context, request *iso15118.InstallCertificateRequest) (response *iso15118.InstallCertificateResponse, err error) {
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}
