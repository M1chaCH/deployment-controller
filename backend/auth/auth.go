package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

const addedAgentContextKey = "agent-changed"

// HandleLoginWithValidCredentials handles some more steps regarding the authentication of a user
// 1. get token and client
// 2. make sure user and client are linked
// 3. handle mfa
// 4. generate success token
// returns true: all success & cookie is updated, false: something failed, response is prepared
func HandleLoginWithValidCredentials(c *gin.Context, user users.UserCacheItem) string {
	// 1. get token and client
	currentRequestToken, ok := GetCurrentIdentityToken(c)
	if !ok {
		AbortWithCooke(c, http.StatusInternalServerError, "no user info found")
		return LoginStateLoggedOut
	}
	client, ok, err := clients.LoadClientInfo(currentRequestToken.Issuer)
	if err != nil {
		logs.Warn("request has no client")
		AbortWithCooke(c, http.StatusInternalServerError, "no user info found")
		return LoginStateLoggedOut
	}

	// 2. make sure user and client are linked
	if client.RealUserId == "" {
		_, err = clients.AddUserToClient(client.Id, user.Id)
		if err != nil {
			logs.Warn(fmt.Sprintf("could not link user with client (%s -> %s): %v", client.Id, user.Id, err))
			AbortWithCooke(c, http.StatusInternalServerError, "login failed")
			return LoginStateLoggedOut
		}
	} else if client.RealUserId != user.Id {
		// client has two users -> duplicate client for new user but only keep current device
		client, err = clients.CreateNewClient(client.Id, user.Id, currentRequestToken.OriginIp, currentRequestToken.OriginAgent)
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to create new user for existing client: %v", err))
			AbortWithCooke(c, http.StatusInternalServerError, "login failed")
			return LoginStateLoggedOut
		}
	}

	user, err = users.UpdateUser(framework.GetTx(c), user.Id, user.Mail, user.Password, user.Salt, user.Admin, user.Blocked, user.Onboard, time.Now(), make([]string, 0), make([]string, 0))
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to update last login of user: %v", err))
		AbortWithCooke(c, http.StatusInternalServerError, "login failed")
		return LoginStateLoggedOut
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
			return LoginStateWaitingMfa
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
		currentRequestToken.OriginIp,
		currentRequestToken.OriginAgent)
	SetCurrentIdentityToken(c, newToken)
	return loginState
}

var processNewClientMutex sync.Mutex

// processNewClient checks if another client can be found by the ip and agent. if not creates the client with the given ID
// if client was found -> clientId will change
// not transactional -> will always save
//
// this function can only be called once at a time per entire server -> race conditions
// locking mechanism could later be improved with sync.Cond to only lock for specific ip & agent, but only if we hit > 10'000 concurrent users...
func processNewClient(clientId, ip, userAgent string) (IdentityToken, error) {
	processNewClientMutex.Lock()
	defer processNewClientMutex.Unlock()

	client, found, err := clients.TryFindExistingClient(ip, userAgent)
	if err != nil {
		logs.Warn(fmt.Sprintf("no existing client found for %s:%s due to db error: %v", ip, userAgent, err))
		return IdentityToken{}, err
	}

	// client does not exist
	if !found {
		client, err = clients.CreateNewClient(clientId, "", ip, userAgent)
		if err != nil {
			return IdentityToken{}, err
		}
	}

	return createIdentityToken(client.Id, client.RealUserId, "", false, LoginStateLoggedOut, ip, userAgent), nil
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
