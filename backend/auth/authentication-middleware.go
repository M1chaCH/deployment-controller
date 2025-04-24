package auth

import (
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthenticationMiddleware is called for every request
// makes sure that the request has a valid cookie with a clientId, doesn't do any authorisation specific validations
// after this Middleware context.Get(updatedIdJwtContextKey) can be used as the source of truth regarding *authentication*
//
// Docs: https://lucid.app/lucidspark/72ac4259-9ff9-4927-972f-d3c4758ab0a6/edit?beaconFlowId=6B32CEF8DD9B4049&page=0_0#
func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIp := parseIp(c)
		requestAgent := c.Request.UserAgent()

		// 1. check if valid token is available
		requestToken, ok := getIdentityToken(c, idJwtContextKey)
		if !ok || requestToken.Issuer == "" {
			newIdToken, err := processNewClient(c, uuid.NewString(), requestIp, requestAgent)
			if err != nil {
				logs.Warn(c, "could not create new client, %v", err)
				AbortWithCooke(c, 500, "failed to process request")
				return
			}

			c.Set(updatedIdJwtContextKey, newIdToken)
			c.Next()
			return
		}

		// 2. check if issuer exists
		client, found, err := clients.LoadClientInfo(c, requestToken.Issuer)
		if err != nil {
			logs.Warn(c, "client from cookie was not found due to internal error!, %v", err)
			AbortWithCooke(c, 500, "failed to process request")
			return
		}
		if !found {
			logs.Warn(c, "client from cookie was not found, %s", requestToken.Issuer)
			newIdToken, err := processNewClient(c, requestToken.Issuer, requestIp, requestAgent)
			if err != nil {
				logs.Warn(c, "could not create new client, %v", err)
				AbortWithCooke(c, 500, "failed to process request")
				return
			}

			c.Set(updatedIdJwtContextKey, newIdToken)
			c.Next()
			return
		}

		// 3. Apply Authentication Rules depending on login State
		newDeviceIsKnown := client.IsDeviceKnown(requestIp, requestAgent)
		switch requestToken.LoginState {
		case LoginStateTwofactorWaiting:
			// LoginStateTwofactorWaiting: Nothing can change
			if didDeviceChange(requestIp, requestAgent, requestToken) {
				if !newDeviceIsKnown {
					requestToken.LoginState = LoginStateLoggedOut
					addDeviceAndComplete(c, requestToken, requestIp, requestAgent)
				} else {
					requestToken.LoginState = LoginStateLoggedOut
					complete(c, requestToken, requestIp, requestAgent)
				}
			} else {
				complete(c, requestToken, requestIp, requestAgent)
			}
			return
		case LoginStateLoggedIn, LoginStateOnboardingWaiting:
			// normal logged in state: Agent can't change
			if requestToken.OriginAgent != requestAgent {
				requestToken.LoginState = LoginStateLoggedOut
			}

			if newDeviceIsKnown {
				complete(c, requestToken, requestIp, requestAgent)
			} else {
				addDeviceAndComplete(c, requestToken, requestIp, requestAgent)
			}
			return
		default:
		case LoginStateLoggedOut:
			// logged out: just keep track of the device
			if newDeviceIsKnown {
				complete(c, requestToken, requestIp, requestAgent)
			} else {
				addDeviceAndComplete(c, requestToken, requestIp, requestAgent)
			}

			return
		}

		logs.Error(c, "login state could not be handled properly for client: %v (this line should not be reachable) -- failing", client.Id)
		AbortWithCooke(c, 500, "failed to process request")
	}
}

func didDeviceChange(requestIp, requestAgent string, token IdentityToken) bool {
	return requestIp != token.OriginIp || requestAgent != token.OriginAgent
}

func addDeviceAndComplete(c *gin.Context, token IdentityToken, requestIp, requestAgent string) {
	_, err := clients.AddDeviceToClient(c, token.Issuer, requestIp, requestAgent)
	if err != nil {
		logs.Warn(c, "failed to add device to client, %v", err)
		AbortWithCooke(c, 500, "failed to process request")
		return
	}

	logs.Info(c, "client changed to other known device: client:%s agent:%s ip:%s", token.Issuer, requestAgent, requestIp)
	complete(c, token, requestIp, requestAgent)
}

func complete(c *gin.Context, token IdentityToken, requestIp, requestAgent string) {
	token = createIdentityToken(token.Issuer, token.UserId, token.Mail, token.Admin, token.LoginState, requestIp, requestAgent)
	c.Set(updatedIdJwtContextKey, token)
	c.Next()
}

func parseIp(c *gin.Context) string {
	realIp := c.GetHeader("X-Real-Ip")
	if realIp != "" {
		return realIp
	}

	return c.ClientIP()
}
