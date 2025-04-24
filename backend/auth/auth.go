package auth

import (
	"github.com/M1chaCH/deployment-controller/auth/mfa"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

// HandleAndCompleteLogin handles some more steps regarding the authentication of a user
// 1. get token and client
// 2. make sure user and client are linked
// 3. generate success token
func HandleAndCompleteLogin(c *gin.Context, user users.UserEntity) {
	// 1. get token and client
	currentRequestToken, ok := GetCurrentIdentityToken(c)
	if !ok {
		AbortWithCooke(c, http.StatusInternalServerError, "no user info found")
		return
	}
	client, ok, err := clients.LoadClientInfo(c, currentRequestToken.Issuer)
	if err != nil || !ok {
		logs.Warn(c, "request has no client")
		AbortWithCooke(c, http.StatusInternalServerError, "no user info found")
		return
	}

	// 2. make sure user and client are linked
	if client.RealUserId == "" {
		existingUserClient, found, err := clients.TryFindClientOfUser(c, user.Id)
		if err != nil {
			logs.Warn(c, "could not select clients of user due to server error (user:%s): %v", user.Id, err)
			AbortWithCooke(c, http.StatusInternalServerError, "login failed")
			return
		}

		// if the user already has a client -> merge else -> add user to client
		if found {
			client, err = clients.MergeDevicesAndDelete(c, existingUserClient, client)
			if err != nil {
				logs.Warn(c, "could not merge devices of two clients (1 : %s - 2 : %s): %v", currentRequestToken.Issuer, existingUserClient.Id, err)
				AbortWithCooke(c, http.StatusInternalServerError, "login failed")
				return
			}
		} else {
			client, err = clients.AddUserToClient(c, client.Id, user.Id)
			if err != nil {
				logs.Warn(c, "could not link user with client (%s -> %s): %v", currentRequestToken.Issuer, user.Id, err)
				AbortWithCooke(c, http.StatusInternalServerError, "login failed")
				return
			}
		}
	} else if client.RealUserId != user.Id {
		// client has two users -> duplicate client for new user but only keep current device
		client, err = clients.CreateNewClient(c, uuid.NewString(), user.Id, currentRequestToken.OriginIp, currentRequestToken.OriginAgent)
		if err != nil {
			logs.Warn(c, "failed to create new user for existing client: %v", err)
			AbortWithCooke(c, http.StatusInternalServerError, "login failed")
			return
		}
	}

	device, found := clients.GetCurrentDevice(client, currentRequestToken.OriginIp, currentRequestToken.OriginAgent)
	if !found {
		logs.Warn(c, "device not found for current user, during login...")
		RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "login failed due to server error"})
		return
	}

	err = users.UpdateUser(c, user.Id, user.Mail, user.Password, user.Salt, user.Admin, user.Blocked, user.Onboard, time.Now(), user.MfaType, make([]string, 0), make([]string, 0))
	if err != nil {
		logs.Warn(c, "failed to update last login of user: %v", err)
		AbortWithCooke(c, http.StatusInternalServerError, "login failed")
		return
	}

	loginState := LoginStateLoggedIn
	if !user.Onboard {
		loginState = LoginStateOnboardingWaiting
	} else if !device.Validated {
		loginState = LoginStateTwofactorWaiting

		if user.MfaType == mfa.TypeMail {
			err = mfa.SendMailTotp(c, user.Id, user.Mail, true)
			if err != nil {
				logs.Warn(c, "failed to send mail to user (%v) (for MFA)", err)
				AbortWithCooke(c, http.StatusInternalServerError, "login failed - could not send mfa mail")
				return
			}
		}
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
	RespondWithCookie(c, http.StatusOK, gin.H{"state": loginState})
}

func HandleAndCompleteMfaVerification(c *gin.Context, idToken IdentityToken, mfaToken string) {
	user, found := users.LoadUserById(c, idToken.UserId)
	if !found {
		RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	ok, err := mfa.Validate(framework.GetTx(c), idToken.UserId, user.MfaType, mfaToken, true)
	if err != nil {
		if err.Error() == framework.ErrNotValidated {
			logs.Error(c, "failed to validate token - user is not onboard - this must be a bug, this case should not happen: %v", err)
			idToken.LoginState = LoginStateOnboardingWaiting
			SetCurrentIdentityToken(c, idToken)
		} else {
			logs.Info(c, "failed to validate token: %v", err)
		}

		RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	if !ok {
		RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	// mark device as validated
	client, found, err := clients.LoadClientInfo(c, idToken.Issuer)
	if err != nil || !found {
		logs.Warn(c, "could not load or did not find client: %v (%v)", err, found)
		RespondWithCookie(c, http.StatusInternalServerError, "verification failed - server error")
		return
	}
	device, found := clients.GetCurrentDevice(client, idToken.OriginIp, idToken.OriginAgent)
	if !found {
		logs.Warn(c, "did not find current device in client: %v", found)
		RespondWithCookie(c, http.StatusInternalServerError, "verification failed - server error")
		return
	}

	err = clients.MarkDeviceAsValidated(c, device.ClientId, device.Id)
	if err != nil {
		logs.Warn(c, "failed to mark device as validated: %v", err)
		RespondWithCookie(c, http.StatusInternalServerError, "verification failed - server error")
		return
	}

	idToken.LoginState = LoginStateLoggedIn
	SetCurrentIdentityToken(c, idToken)
	RespondWithCookie(c, http.StatusOK, gin.H{"message": "token valid"})
}

var processNewClientMutex sync.Mutex

// processNewClient checks if another client can be found by the ip and agent. if not creates the client with the given ID
// if client was found -> clientId will change
// not transactional -> will always save
//
// this function can only be called once at a time per entire server -> race conditions
// locking mechanism could later be improved with sync.Cond to only lock for specific ip & agent, but only if we hit > 1'000 concurrent users...
func processNewClient(c *gin.Context, clientId, ip, userAgent string) (IdentityToken, error) {
	processNewClientMutex.Lock()
	defer processNewClientMutex.Unlock()

	client, found, err := clients.TryFindExistingClient(c, ip, userAgent)
	if err != nil {
		logs.Warn(c, "no existing client found for %s:%s due to db error: %v", ip, userAgent, err)
		return IdentityToken{}, err
	}

	// client does not exist
	if !found {
		client, err = clients.CreateNewClient(c, clientId, "", ip, userAgent)
		if err != nil {
			return IdentityToken{}, err
		}
	}

	return createIdentityToken(client.Id, client.RealUserId, "", false, LoginStateLoggedOut, ip, userAgent), nil
}

const userContextKey = "current-user"

func addUserToRequest(c *gin.Context) (users.UserEntity, bool) {
	token, ok := GetCurrentIdentityToken(c)
	if !ok {
		return users.UserEntity{}, false
	}

	if token.UserId != "" {
		user, ok := users.LoadUserById(c, token.UserId)
		if ok {
			c.Set(userContextKey, user)
			return user, true
		}
	}

	return users.UserEntity{}, false
}

func GetCurrentUser(c *gin.Context) (users.UserEntity, bool) {
	userValue, ok := c.Get(userContextKey)
	if ok {
		user, castOk := userValue.(users.UserEntity)
		if castOk {
			return user, true
		}
	}

	user, ok := addUserToRequest(c)
	if ok {
		return user, true
	}

	return users.UserEntity{}, false
}
