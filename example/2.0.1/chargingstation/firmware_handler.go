package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lorenzodonini/ocpp-go/ocpp"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/firmware"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/types"
	"github.com/lorenzodonini/ocpp-go/ocppj"
	"go.opentelemetry.io/otel"
)

func (handler *ChargingStationHandler) OnPublishFirmware(ctx context.Context, request *firmware.PublishFirmwareRequest) (response *firmware.PublishFirmwareResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnPublishFirmware.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}

func (handler *ChargingStationHandler) OnUnpublishFirmware(ctx context.Context, request *firmware.UnpublishFirmwareRequest) (response *firmware.UnpublishFirmwareResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnUnpublishFirmware.handler")
	defer span.End()
	logDefault(request.GetFeatureName()).Warnf("Unsupported feature")
	return nil, ocpp.NewHandlerError(ocppj.NotSupported, "Not supported")
}

func (handler *ChargingStationHandler) OnUpdateFirmware(ctx context.Context, request *firmware.UpdateFirmwareRequest) (response *firmware.UpdateFirmwareResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnUpdateFirmware.handler")
	defer span.End()
	retries := 0
	retryInterval := 30
	if request.Retries != nil {
		retries = *request.Retries
	}
	if request.RetryInterval != nil {
		retryInterval = *request.RetryInterval
	}
	logDefault(request.GetFeatureName()).Infof("starting update firmware procedure")
	go updateFirmware(ctx, request.Firmware.Location, request.Firmware.RetrieveDateTime, request.Firmware.InstallDateTime, retries, retryInterval)
	return firmware.NewUpdateFirmwareResponse(firmware.UpdateFirmwareStatusAccepted), nil
}

func updateFirmwareStatus(ctx context.Context, status firmware.FirmwareStatus, props ...func(request *firmware.FirmwareStatusNotificationRequest)) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "updateFirmwareStatus")
	defer span.End()
	statusConfirmation, err := chargingStation.FirmwareStatusNotification(context.Background(), status, props...)
	checkError(err)
	logDefault(statusConfirmation.GetFeatureName()).Infof("firmware status updated to %v", status)
}

// Retrieve data and install date are ignored for this test function.
func updateFirmware(ctx context.Context, location string, retrieveDate *types.DateTime, installDate *types.DateTime, retries int, retryInterval int) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "updateFirmware")
	defer span.End()

	updateFirmwareStatus(ctx, firmware.FirmwareStatusDownloading)
	err := downloadFile("/tmp/out.bin", location)
	if err != nil {
		logDefault(firmware.UpdateFirmwareFeatureName).Errorf("error while downloading file %v", err)
		updateFirmwareStatus(ctx, firmware.FirmwareStatusDownloadFailed)
		return
	}
	updateFirmwareStatus(ctx, firmware.FirmwareStatusDownloaded)
	// Simulate installation
	updateFirmwareStatus(ctx, firmware.FirmwareStatusInstalling)
	time.Sleep(time.Second * 5)
	// Notify completion
	updateFirmwareStatus(ctx, firmware.FirmwareStatusInstalled)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
