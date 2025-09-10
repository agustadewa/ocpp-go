// The diagnostics functional block contains OCPP 2.0 features than enable remote diagnostics of problems with a charging station.
package certificates

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp"
)

// Needs to be implemented by Charging stations for handling messages part of the OCPP 1.6j security extension.
type ChargePointHandler interface {
	// OnDeleteCertificate is called on a charging station whenever a DeleteCertificateRequest is received from the CSMS.
	OnDeleteCertificate(ctx context.Context, request *DeleteCertificateRequest) (response *DeleteCertificateResponse, err error)
	// OnGetInstalledCertificateIds is called on a charging station whenever a GetInstalledCertificateIdsRequest is received from the CSMS.
	OnGetInstalledCertificateIds(ctx context.Context, request *GetInstalledCertificateIdsRequest) (response *GetInstalledCertificateIdsResponse, err error)
	// OnInstallCertificate is called on a charging station whenever an InstallCertificateRequest is received from the CSMS.
	OnInstallCertificate(ctx context.Context, request *InstallCertificateRequest) (response *InstallCertificateResponse, err error)
}

const ProfileName = "Certificates"

var Profile = ocpp.NewProfile(
	ProfileName,
	InstallCertificateFeature{},
	DeleteCertificateFeature{},
	GetInstalledCertificateIdsFeature{},
)
