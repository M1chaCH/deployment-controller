package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/auth/clients"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/users"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LoginDto struct {
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Token    uint8  `json:"token,omitempty"`
}

// HandleLoginWithValidCredentials handles some more steps regarding the authentication of a user
// 1. get token and client
// 2. make sure user and client are linked
// 3. handle mfa
// 4. generate success token
func HandleLoginWithValidCredentials(c *gin.Context, user users.User) {
	// 1. get token and client
	currentRequestToken, ok := GetCurrentIdentityToken(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "no user info found"})
		return
	}
	client, err := clients.LoadClientInfo(currentRequestToken.Issuer)
	if err != nil {
		fmt.Println("WARN: request has no client")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "no user info found"})
		return
	}

	// 2. make sure user and client are linked
	if client.RealUserId == "" {
		_, err = clients.AddUserToClient(framework.GetTx(c), client.Id, user.Id)
		if err != nil {
			fmt.Printf("WARN: could not link user with client (%s -> %s)\n", client.Id, user.Id)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "login failed"})
			return
		}
	} else if client.RealUserId != user.Id {
		// client has two users -> duplicate client for new user but only keep current device
		client, err = clients.CreateNewClient(framework.GetTx(c), user.Id, currentRequestToken.OriginIp, currentRequestToken.OriginAgent)
		if err != nil {
			fmt.Println("WARN: failed to create new user for existing client")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "login failed"})
			return
		}
	}

	// 3. handle mfa
	// if currently waiting for currentRequestToken, validate currentRequestToken

	// if user agent is not known -> request mfa currentRequestToken
	newAgentInThisRequest, newAgentCreated := c.Get(AddedAgentContextKey)
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
	}

	// 4. generate success token
	newToken := createIdentityToken(client.Id,
		user.Id,
		user.Mail,
		user.Admin,
		LoginStateLoggedIn,
		make([]string, 0),
		currentRequestToken.OriginIp,
		currentRequestToken.OriginAgent)
	c.Set(updatedIdJwtContextKey, newToken)
	c.Status(http.StatusNoContent)
	return
}
