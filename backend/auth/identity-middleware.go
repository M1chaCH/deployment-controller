package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/config"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"net/http"
)

// IdentityJwtMiddleware gets and parses the Identity cookie and adds it to the context as a struct
// does no app specific validation
func IdentityJwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenFailed := false
		tokenString, err := c.Cookie(idJwtCookieName)
		if err != nil || tokenString == "" {
			tokenFailed = true
		}

		var token IdentityToken
		if !tokenFailed {
			token, err = parseIdentityToken(c, tokenString)
			if err != nil {
				tokenFailed = true
			}
		}

		if tokenFailed {
			c.Set(idJwtContextKey, nil)
		} else {
			c.Set(idJwtContextKey, token)
			c.Set(updatedIdJwtContextKey, token)

			logs.AddApmLabels(c, logs.ApmLabels{
				"ctl.user.id":    token.UserId,
				"ctl.user.login": token.LoginState,
				"ctl.user.admin": token.Admin,
				"ctl.client.id":  token.Issuer,
				"ctl.tk.issued":  token.IssuedAt,
				"ctl.tk.expires": token.ExpiresAt,
			})
		}

		c.Next()
	}
}

func AppendJwtToken(c *gin.Context) {
	newToken, ok := getIdentityToken(c, updatedIdJwtContextKey)
	if !ok {
		logs.Warn(c, "found request with invalid updated ID tokens")
		return
	}

	newTokenString, err := newToken.ToJwtString()
	if err != nil {
		logs.Error(c, fmt.Sprintf("could not parse ID tokens, %v", err))
		return
	}

	// keep the cookie as long as possible, don't want to lose the clientId
	cnf := config.Config()
	maxAge := 60 * 60 * 24 * 400 // 400 days in seconds
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(idJwtCookieName, newTokenString, maxAge, "/", config.JWT.Domain, true, true)
}
