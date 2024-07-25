package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/jmoiron/sqlx"
)

const addedAgentContextKey = "agent-changed"

// TODO needs to be synchronized across all requests
func processNewClient(tx *sqlx.Tx, ip, userAgent string) (IdentityToken, error) {
	client, err := clients.LoadExistingClient(ip, userAgent)
	if err != nil {
		logs.Warn(fmt.Sprintf("no existing client found for %s:%s due to db error: %v", ip, userAgent, err))
		return IdentityToken{}, err
	}

	// client does not exist
	if client.Id == "" {
		client, err = clients.CreateNewClient(tx, "", ip, userAgent)
		if err != nil {
			return IdentityToken{}, err
		}
	}

	return createIdentityToken(client.Id, client.RealUserId, "", false, LoginStateLoggedOut, "", ip, userAgent), nil
}
