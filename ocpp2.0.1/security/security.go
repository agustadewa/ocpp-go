// The security functional block contains OCPP 2.0 features aimed at providing E2E security between a CSMS and a Charging station.
package security

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp"
)

// Needs to be implemented by a CSMS for handling messages part of the OCPP 2.0 Security profile.
type CSMSHandler interface {
	// OnSecurityEventNotification is called on the CSMS whenever a SecurityEventNotificationRequest is received from a charging station.
	OnSecurityEventNotification(ctx context.Context, chargingStationID string, request *SecurityEventNotificationRequest) (response *SecurityEventNotificationResponse, err error)
	// OnSignCertificate is called on the CSMS whenever a SignCertificateRequest is received from a charging station.
	OnSignCertificate(ctx context.Context, chargingStationID string, request *SignCertificateRequest) (response *SignCertificateResponse, err error)
}

// Needs to be implemented by Charging stations for handling messages part of the OCPP 2.0 Security profile.
type ChargingStationHandler interface {
	// OnCertificateSigned is called on a charging station whenever a CertificateSignedRequest is received from the CSMS.
	OnCertificateSigned(ctx context.Context, request *CertificateSignedRequest) (response *CertificateSignedResponse, err error)
}

const ProfileName = "Security"

var Profile = ocpp.NewProfile(
	ProfileName,
	CertificateSignedFeature{},
	SecurityEventNotificationFeature{},
	SignCertificateFeature{},
)
