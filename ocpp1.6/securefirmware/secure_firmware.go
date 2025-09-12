// The diagnostics functional block contains OCPP 1.6J extension features than enable remote firmware updates on charging stations.
package securefirmware

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp"
)

type CentralSystemHandler interface {
	OnSignedFirmwareStatusNotification(ctx context.Context, chargingStationID string, request *SignedFirmwareStatusNotificationRequest) (response *SignedFirmwareStatusNotificationResponse, err error)
}

// Needs to be implemented by Charging stations for handling messages part of the OCPP 1.6j security extension.
type ChargePointHandler interface {
	OnSignedUpdateFirmware(ctx context.Context, request *SignedUpdateFirmwareRequest) (response *SignedUpdateFirmwareResponse, err error)
}

const ProfileName = "SecureFirmwareUpdate"

var Profile = ocpp.NewProfile(
	ProfileName,
	SignedFirmwareStatusNotificationFeature{},
	SignedUpdateFirmwareFeature{},
)
