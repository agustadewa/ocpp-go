package main

import (
	"context"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/authorization"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/types"
)

func (c *CSMSHandler) OnAuthorize(ctx context.Context, chargingStationID string, request *authorization.AuthorizeRequest) (confirmation *authorization.AuthorizeResponse, err error) {
	logDefault(chargingStationID, request.GetFeatureName()).Infof("client with token %v authorized", request.IdToken)
	confirmation = authorization.NewAuthorizationResponse(*types.NewIdTokenInfo(types.AuthorizationStatusAccepted))
	return
}
