package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AdminAuthorisationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		idToken, ok := GetCurrentIdentityToken(c)
		if !ok {
			AbortWithCooke(c, http.StatusForbidden, "access denied")
			return
		}

		if !idToken.Admin {
			AbortWithCooke(c, http.StatusForbidden, "access denied")
			return
		}

		c.Next()
	}
}
