package rest

import (
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitOpenEndpoints(router *gin.RouterGroup) {
	router.GET("/user", getCurrentUser)
	router.GET("/pages", getPages)
}

func getCurrentUser(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "no user info found"})
		return
	}

	user, ok := auth.GetCurrentUser(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "no user info found"})
		return
	}

	pageStrings := make([]string, len(user.Pages))
	for _, p := range user.Pages {
		pageStrings = append(pageStrings, p.TechnicalName)
	}

	body := gin.H{
		"id":           user.Id,
		"mail":         user.Mail,
		"admin":        user.Admin,
		"privatePages": pageStrings,
		"loginState":   idToken.LoginState,
	}
	auth.RespondWithCookie(c, http.StatusOK, body)
}

func getPages(c *gin.Context) {

}
