// Contains the Basic Charge Point functionality comparable with OCPP 1.5.
package core

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp"
)

// Needs to be implemented by Central systems for handling messages part of the OCPP 1.6 Core profile.
type CentralSystemHandler interface {
	OnAuthorize(ctx context.Context, chargePointId string, request *AuthorizeRequest) (confirmation *AuthorizeConfirmation, err error)
	OnBootNotification(ctx context.Context, chargePointId string, request *BootNotificationRequest) (confirmation *BootNotificationConfirmation, err error)
	OnDataTransfer(ctx context.Context, chargePointId string, request *DataTransferRequest) (confirmation *DataTransferConfirmation, err error)
	OnHeartbeat(ctx context.Context, chargePointId string, request *HeartbeatRequest) (confirmation *HeartbeatConfirmation, err error)
	OnMeterValues(ctx context.Context, chargePointId string, request *MeterValuesRequest) (confirmation *MeterValuesConfirmation, err error)
	OnStatusNotification(ctx context.Context, chargePointId string, request *StatusNotificationRequest) (confirmation *StatusNotificationConfirmation, err error)
	OnStartTransaction(ctx context.Context, chargePointId string, request *StartTransactionRequest) (confirmation *StartTransactionConfirmation, err error)
	OnStopTransaction(ctx context.Context, chargePointId string, request *StopTransactionRequest) (confirmation *StopTransactionConfirmation, err error)
}

// Needs to be implemented by Charge points for handling messages part of the OCPP 1.6 Core profile.
type ChargePointHandler interface {
	OnChangeAvailability(ctx context.Context, request *ChangeAvailabilityRequest) (confirmation *ChangeAvailabilityConfirmation, err error)
	OnChangeConfiguration(ctx context.Context, request *ChangeConfigurationRequest) (confirmation *ChangeConfigurationConfirmation, err error)
	OnClearCache(ctx context.Context, request *ClearCacheRequest) (confirmation *ClearCacheConfirmation, err error)
	OnDataTransfer(ctx context.Context, request *DataTransferRequest) (confirmation *DataTransferConfirmation, err error)
	OnGetConfiguration(ctx context.Context, request *GetConfigurationRequest) (confirmation *GetConfigurationConfirmation, err error)
	OnRemoteStartTransaction(ctx context.Context, request *RemoteStartTransactionRequest) (confirmation *RemoteStartTransactionConfirmation, err error)
	OnRemoteStopTransaction(ctx context.Context, request *RemoteStopTransactionRequest) (confirmation *RemoteStopTransactionConfirmation, err error)
	OnReset(ctx context.Context, request *ResetRequest) (confirmation *ResetConfirmation, err error)
	OnUnlockConnector(ctx context.Context, request *UnlockConnectorRequest) (confirmation *UnlockConnectorConfirmation, err error)
}

// THe profile name
var ProfileName = "Core"

// Provides support for Basic Charge Point functionality comparable with OCPP 1.5.
var Profile = ocpp.NewProfile(
	ProfileName,
	BootNotificationFeature{},
	AuthorizeFeature{},
	ChangeAvailabilityFeature{},
	ChangeConfigurationFeature{},
	ClearCacheFeature{},
	DataTransferFeature{},
	GetConfigurationFeature{},
	HeartbeatFeature{},
	MeterValuesFeature{},
	RemoteStartTransactionFeature{},
	RemoteStopTransactionFeature{},
	StartTransactionFeature{},
	StopTransactionFeature{},
	StatusNotificationFeature{},
	ResetFeature{},
	UnlockConnectorFeature{})
