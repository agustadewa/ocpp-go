package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/display"
)

func (handler *ChargingStationHandler) OnClearDisplay(ctx context.Context, request *display.ClearDisplayRequest) (response *display.ClearDisplayResponse, err error) {
	logDefault(request.GetFeatureName()).Infof("cleared display message %v", request.ID)
	response = display.NewClearDisplayResponse(display.ClearMessageStatusAccepted)
	return
}

func (handler *ChargingStationHandler) OnGetDisplayMessages(ctx context.Context, request *display.GetDisplayMessagesRequest) (response *display.GetDisplayMessagesResponse, err error) {
	logDefault(request.GetFeatureName()).Infof("request %v to send display messages ignored", request.RequestID)
	response = display.NewGetDisplayMessagesResponse(display.MessageStatusUnknown)
	return
}

func (handler *ChargingStationHandler) OnSetDisplayMessage(ctx context.Context, request *display.SetDisplayMessageRequest) (response *display.SetDisplayMessageResponse, err error) {
	logDefault(request.GetFeatureName()).Infof("accepted request to display message %v: %v", request.Message.ID, request.Message.Message.Content)
	response = display.NewSetDisplayMessageResponse(display.DisplayMessageStatusAccepted)
	return
}
