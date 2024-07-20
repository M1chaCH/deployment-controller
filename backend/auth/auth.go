package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/auth/clients"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strings"
	"time"
)

const AddedAgentContextKey = "agent-changed"

/*
 BIG TODO: fix following scenario
- new client comes and browser requests a lot of f.e. js files in parallel.
-> make sure, only one clientId is created and handed out.
  - created JWT probably needs to be cached somewhere (probably) for 1 second --> requests from same origin with no cookie from within one second always get the cached token.
*/

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
func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. check if cookie is here
		requestToken, ok := getIdentityToken(c, idJwtContextKey)
		if !ok || requestToken.Issuer == "" {
			newIdToken, err := processNewClient(framework.GetTx(c), c.ClientIP(), c.Request.UserAgent())
			if err != nil {
				fmt.Printf("ERROR: creating new client: %v\n", err)
				c.AbortWithStatusJSON(500, gin.H{"message": "failed to process request"})
				return
			}

			c.Set(updatedIdJwtContextKey, newIdToken)
			c.Next()
			return
		}

		client, err := clients.LoadClientInfo(requestToken.Issuer)
		if err != nil {
			fmt.Printf("ERROR: client from cookie was not found %v\n", err)
			c.AbortWithStatusJSON(404, gin.H{"message": "some required data was not found"})
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
					strings.Split(requestToken.PrivatePages, PrivatePagesDelimiter),
					ip,
					userAgent)

				c.Set(updatedIdJwtContextKey, token)
			} else {
				_, err := clients.AddDeviceToClient(framework.GetTx(c), requestToken.Issuer, ip, userAgent)
				if err != nil {
					fmt.Printf("ERROR: while adding device to client: %v\n", err)
					c.AbortWithStatusJSON(500, gin.H{"message": "failed to process request"})
					return
				}

				if userAgent != requestToken.OriginAgent {
					c.Set(AddedAgentContextKey, userAgent)
				}
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
				make([]string, 0),
				ip,
				userAgent)

			c.Set(updatedIdJwtContextKey, token)
			c.Next()
			return
		}

		// 4. check if agent changed and logged in
		if userAgent != requestToken.OriginAgent && requestToken.LoginState != LoginStateLoggedOut {
			token := createIdentityToken(requestToken.Issuer,
				requestToken.Subject,
				requestToken.Mail,
				false,
				LoginStateLoggedOut,
				make([]string, 0),
				ip,
				userAgent)

			c.Set(updatedIdJwtContextKey, token)
			c.Next()
			return
		}

		// while waiting for two factor, ip can't change
		if ip != requestToken.OriginIp && requestToken.LoginState == LoginStateTwofactorWaiting {
			token := createIdentityToken(requestToken.Issuer,
				requestToken.Subject,
				requestToken.Mail,
				false,
				LoginStateLoggedOut,
				make([]string, 0),
				ip,
				userAgent)

			c.Set(updatedIdJwtContextKey, token)
			c.Next()
			return
		}

		c.Next()
	}
}

// AdminAuthorisationMiddleware is called for every request where the user needs to be an admin
// validates that the user is an admin
func AdminAuthorisationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := getIdentityToken(c, updatedIdJwtContextKey)
		if !ok || token.LoginState != LoginStateLoggedIn {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
			return
		}

		if !token.Admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "access denied"})
			return
		}

		c.Next()
	}
}

// TODO needs to be synchronized across all requests
func processNewClient(tx *sqlx.Tx, ip, userAgent string) (IdentityToken, error) {
	client, err := clients.LoadExistingClient(ip, userAgent)
	if err != nil {
		fmt.Printf("WARN: no existing client found for %s:%s due to db error: %v\n", ip, userAgent, err)
	}

	// client does not exist
	if client.Id == "" {
		client, err = clients.CreateNewClient(tx, "", ip, userAgent)
		if err != nil {
			return IdentityToken{}, err
		}
	}

	return createIdentityToken(client.Id, client.RealUserId, "", false, LoginStateLoggedOut, make([]string, 0), ip, userAgent), nil
}
