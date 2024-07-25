package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
)

const addedAgentContextKey = "agent-changed"

// TODO needs to be synchronized across all requests
// processNewClient checks if an other client can be found by the ip and agent. if not creates the client with the given ID
// if client was found -> clientId will change
// not transactional -> will always save
func processNewClient(clientId, ip, userAgent string) (IdentityToken, error) {
	client, found, err := clients.TryFindExistingClient(ip, userAgent)
	if err != nil {
		logs.Warn(fmt.Sprintf("no existing client found for %s:%s due to db error: %v", ip, userAgent, err))
		return IdentityToken{}, err
	}

	// client does not exist
	if !found {
		tx, err := framework.DB().Beginx()
		if err != nil {
			return IdentityToken{}, err
		}

		client, err = clients.CreateNewClient(tx, clientId, "", ip, userAgent)
		if err != nil {
			return IdentityToken{}, err
		}

		err = tx.Commit()
		if err != nil {
			return IdentityToken{}, err
		}
	}

	return createIdentityToken(client.Id, client.RealUserId, "", false, LoginStateLoggedOut, "", ip, userAgent), nil
}

const userContextKey = "current-user"

func addUserToRequest(c *gin.Context) (users.UserCacheItem, bool) {
	token, ok := GetCurrentIdentityToken(c)
	if !ok {
		return users.UserCacheItem{}, false
	}

	if token.UserId != "" {
		user, ok := users.LoadUserById(token.UserId)
		if ok {
			c.Set(userContextKey, user)
			return user, true
		}
	}

	return users.UserCacheItem{}, false
}

func GetCurrentUser(c *gin.Context) (users.UserCacheItem, bool) {
	userValue, ok := c.Get(userContextKey)
	if ok {
		user, castOk := userValue.(users.UserCacheItem)
		if castOk {
			return user, true
		}
	}

	user, ok := addUserToRequest(c)
	if ok {
		return user, true
	}

	return users.UserCacheItem{}, false
}
