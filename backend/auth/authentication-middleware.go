package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"time"
)

// AuthenticationMiddleware is called for every request
// makes sure that the request has a valid cookie with a clientId, doesn't do any authorisation specific validations
// after this Middleware context.Get(updatedIdJwtContextKey) can be used as the source of truth regarding *authentication*
//
// flow:
//  1. check if cookie is here
//     no -> add new one to request
//     add new ID Token to context
//  2. check if has new IP or Agent -- TODO needs to match this
//     yes -> check if combination of IP and Agent are known
//     -- no -> add new agent and ip
//     -- & if new agent store new agent in context (so login knows if agent was known before this request)
//  3. check if cookie expired or in future
//     yes -> create new, keep client & user id, but logout
//  4. check if agent changed to unknown agent and logged in
//     yes -> logout
//  5. check if logged in and user blocked
//     yes -> logout
func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. check if cookie is here
		requestToken, ok := getIdentityToken(c, idJwtContextKey)
		if !ok || requestToken.Issuer == "" {
			newIdToken, err := processNewClient(uuid.NewString(), c.ClientIP(), c.Request.UserAgent())
			if err != nil {
				logs.Warn(fmt.Sprintf("could not create new client, %v", err))
				AbortWithCooke(c, 500, "failed to process request")
				return
			}

			c.Set(updatedIdJwtContextKey, newIdToken)
			c.Next()
			return
		}

		client, found, err := clients.LoadClientInfo(requestToken.Issuer)
		if err != nil {
			logs.Warn(fmt.Sprintf("client from cookie was not found due to internal error!, %v", err))
			AbortWithCooke(c, 404, "some required data was not found")
			return
		}
		if !found {
			logs.Warn(fmt.Sprintf("client from cookie was not found, %s", requestToken.Issuer))
			newIdToken, err := processNewClient(requestToken.Issuer, c.ClientIP(), c.Request.UserAgent())
			if err != nil {
				logs.Warn(fmt.Sprintf("could not create new client, %v", err))
				AbortWithCooke(c, 500, "failed to process request")
				return
			}

			c.Set(updatedIdJwtContextKey, newIdToken)
			c.Next()
			return
		}

		// 2. check if has new IP or Agent
		userAgent := c.Request.UserAgent()
		ip := c.ClientIP()
		if userAgent != requestToken.OriginAgent || ip != requestToken.OriginIp {
			if client.IsDeviceKnown(ip, userAgent) {
				token := createIdentityToken(requestToken.Issuer,
					requestToken.Subject,
					requestToken.Mail,
					requestToken.Admin,
					requestToken.LoginState,
					requestToken.PrivatePages,
					ip,
					userAgent)

				c.Set(updatedIdJwtContextKey, token)
				logs.Info(fmt.Sprintf("client changed to other known device: client:%s agent:%s ip:%s", requestToken.Issuer, userAgent, ip))
			} else {
				_, err := clients.AddDeviceToClient(framework.GetTx(c), requestToken.Issuer, ip, userAgent)
				if err != nil {
					logs.Warn(fmt.Sprintf("failed to add device to client, %v", err))
					AbortWithCooke(c, 500, "failed to process request")
					return
				}

				// token is updated later

				if userAgent != requestToken.OriginAgent {
					c.Set(addedAgentContextKey, userAgent)
				}
				logs.Info(fmt.Sprintf("client changed to new device: client:%s agent:%s ip:%s", requestToken.Issuer, userAgent, ip))
			}
		}

		// 3. check if cookie expired or in future
		expiredAt, err := requestToken.GetExpirationTime()
		issuedAt, err := requestToken.GetIssuedAt()
		now := time.Now()
		if err != nil || expiredAt.Before(now) || issuedAt.Before(now) {
			token := createIdentityToken(requestToken.Issuer,
				requestToken.Subject,
				requestToken.Mail,
				false,
				LoginStateLoggedOut,
				requestToken.PrivatePages,
				ip,
				userAgent)

			c.Set(updatedIdJwtContextKey, token)
			c.Next()
			return
		}

		// 4. check if agent changed and logged in
		// 5. check if logged in and user blocked
		user, userExists := GetCurrentUser(c)
		if requestToken.LoginState != LoginStateLoggedOut && (userAgent != requestToken.OriginAgent || !userExists || !user.Blocked) {
			token := createIdentityToken(requestToken.Issuer,
				requestToken.Subject,
				requestToken.Mail,
				requestToken.Admin,
				LoginStateLoggedOut,
				requestToken.PrivatePages,
				ip,
				userAgent)

			c.Set(updatedIdJwtContextKey, token)
			c.Next()
			logs.Warn(fmt.Sprintf("changed login state of client due to agent change or blocked: client:%s agent:%s ip:%s", requestToken.Issuer, userAgent, ip))
			return
		}

		// while waiting for two factor, ip can't change
		if ip != requestToken.OriginIp && requestToken.LoginState == LoginStateTwofactorWaiting {
			token := createIdentityToken(requestToken.Issuer,
				requestToken.Subject,
				requestToken.Mail,
				requestToken.Admin,
				LoginStateLoggedOut,
				requestToken.PrivatePages,
				ip,
				userAgent)

			c.Set(updatedIdJwtContextKey, token)
			c.Next()
			logs.Info(fmt.Sprintf("ip changed while waiting for MFA: client:%s agent:%s ip:%s", requestToken.Issuer, userAgent, ip))
			return
		}

		c.Next()
	}
}
