package main

import (
	"context"
	"encoding/json"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/data"
	"go.opentelemetry.io/otel"
)

type DataSample struct {
	SampleString string  `json:"sample_string"`
	SampleValue  float64 `json:"sample_value"`
}

func (handler *ChargingStationHandler) OnDataTransfer(ctx context.Context, request *data.DataTransferRequest) (confirmation *data.DataTransferResponse, err error) {
	ctx, span := otel.Tracer("ocpp-charging-station").Start(ctx, "OnDataTransfer.handler")
	defer span.End()

	var dataSample DataSample
	err = json.Unmarshal(request.Data.([]byte), &dataSample)
	if err != nil {
		logDefault(request.GetFeatureName()).
			Errorf("invalid data received: %v", request.Data)
		span.RecordError(err)
		return nil, err
	}
	logDefault(request.GetFeatureName()).
		Infof("data received: %v, %v", dataSample.SampleString, dataSample.SampleValue)
	return data.NewDataTransferResponse(data.DataTransferStatusAccepted), nil
}
