package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
	"time"
)

const addedAgentContextKey = "agent-changed"

// HandleLoginWithValidCredentials handles some more steps regarding the authentication of a user
// 1. get token and client
// 2. make sure user and client are linked
// 3. handle mfa
// 4. generate success token
// returns true: all success & cookie is updated, false: something failed, response is prepared
func HandleLoginWithValidCredentials(c *gin.Context, user users.UserCacheItem) bool {
	// 1. get token and client
	currentRequestToken, ok := GetCurrentIdentityToken(c)
	if !ok {
		AbortWithCooke(c, http.StatusInternalServerError, "no user info found")
		return false
	}
	client, ok, err := clients.LoadClientInfo(framework.GetTx(c), currentRequestToken.Issuer)
	if err != nil {
		logs.Warn("request has no client")
		AbortWithCooke(c, http.StatusInternalServerError, "no user info found")
		return false
	}

	// 2. make sure user and client are linked
	if client.RealUserId == "" {
		_, err = clients.AddUserToClient(framework.GetTx(c), client.Id, user.Id)
		if err != nil {
			logs.Warn(fmt.Sprintf("could not link user with client (%s -> %s): %v", client.Id, user.Id, err))
			AbortWithCooke(c, http.StatusInternalServerError, "login failed")
			return false
		}
	} else if client.RealUserId != user.Id {
		// client has two users -> duplicate client for new user but only keep current device
		client, err = clients.CreateNewClient(framework.GetTx(c), client.Id, user.Id, currentRequestToken.OriginIp, currentRequestToken.OriginAgent)
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to create new user for existing client: %v", err))
			AbortWithCooke(c, http.StatusInternalServerError, "login failed")
			return false
		}
	}

	user, err = users.UpdateUser(framework.GetTx(c), user.Id, user.Mail, user.Password, user.Salt, user.Admin, user.Blocked, user.Onboard, time.Now(), make([]string, 0), make([]string, 0))
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to update last login of user: %v", err))
		AbortWithCooke(c, http.StatusInternalServerError, "login failed")
		return false
	}

	var userPagesString string
	for i, page := range user.Pages {
		if !page.Private {
			continue
		}

		userPagesString += page.TechnicalName
		if i != len(user.Pages)-1 {
			userPagesString += JwtListDelimiter
		}
	}

	// 3. handle mfa
	// if currently waiting for currentRequestToken, validate currentRequestToken

	// TODO
	// if user agent is not known -> request mfa currentRequestToken
	/*	newAgentInThisRequest, newAgentCreated := c.Get(AddedAgentContextKey)
		// first query should never be true, second should be true if agent is unknown and third should never be false
		if !client.IsDeviceKnown(currentRequestToken.OriginIp, currentRequestToken.OriginAgent) || (newAgentCreated && newAgentInThisRequest == currentRequestToken.OriginAgent) {
			// TODO get private pages
			tokenForTwoFactorAuth := createIdentityToken(client.Id,
				user.Id,
				user.Mail,
				user.Admin,
				LoginStateTwofactorWaiting,
				make([]string, 0),
				currentRequestToken.OriginIp,
				currentRequestToken.OriginAgent)
			c.Set(updatedIdJwtContextKey, tokenForTwoFactorAuth)
			c.JSON(http.StatusOK, gin.H{"message": "require mfa"})
			return
		}*/

	loginState := LoginStateOnboardingWaiting
	if user.Onboard {
		loginState = LoginStateLoggedIn
	}

	// 4. generate success token
	newToken := createIdentityToken(client.Id,
		user.Id,
		user.Mail,
		user.Admin,
		loginState,
		userPagesString,
		currentRequestToken.OriginIp,
		currentRequestToken.OriginAgent)
	SetCurrentIdentityToken(c, newToken)
	return true
}

// TODO needs to be synchronized across all requests
// processNewClient checks if an other client can be found by the ip and agent. if not creates the client with the given ID
// if client was found -> clientId will change
// not transactional -> will always save
func processNewClient(tx *sqlx.Tx, clientId, ip, userAgent string) (IdentityToken, error) {
	client, found, err := clients.TryFindExistingClient(tx, ip, userAgent)
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
		user, ok := users.LoadUserById(framework.GetTx(c), token.UserId)
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
