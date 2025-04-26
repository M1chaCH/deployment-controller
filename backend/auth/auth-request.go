package auth

import (
	"github.com/M1chaCH/deployment-controller/data/pageaccess"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/M1chaCH/deployment-controller/location"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func InitAuthRequest(router gin.IRouter) {
	router.GET("/auth/:page_technical_name", authRequest)
}

func authRequest(c *gin.Context) {
	technicalName := c.Param("page_technical_name")
	if technicalName == "" {
		RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "page name invalid"})
		return
	}

	token, tokenFound := GetCurrentIdentityToken(c)

	var userId string
	if tokenFound && token.UserId != "" && token.LoginState == LoginStateLoggedIn {
		userId = token.UserId
	} else {
		userId = pageaccess.AnonymousUserId
	}

	userPageAccess, err := pageaccess.LoadUserPageAccess(framework.GetTx(c), userId)
	if err != nil {
		logs.Warn(c, "Failed to load user page access for user '%s': %v", token.UserId, err)
		RespondWithCookie(c, http.StatusForbidden, gin.H{"message": "access denied"})
		return
	}

	resolveAndLogData(c, token, technicalName)
	checkTargetAccess(c, technicalName, userPageAccess.Pages)
}

type assertablePage interface {
	GetTechnicalName() string
	GetAccessAllowed() bool
}

func resolveAndLogData(c *gin.Context, token IdentityToken, technicalName string) {
	loc, err := location.LoadLocation(token.OriginIp, true)
	var longitude, latitude float32
	if err == nil {
		longitude = loc.Longitude
		latitude = loc.Latitude
	}

	logs.AddApmLabels(c, logs.ApmLabels{
		"ctl.page.technical_name":   technicalName,
		"ctl.page.access.longitude": longitude,
		"ctl.page.access.latitude":  latitude,
	})
}

func checkTargetAccess[T assertablePage](c *gin.Context, target string, pages []T) {
	for _, page := range pages {
		if strings.ToLower(page.GetTechnicalName()) == strings.ToLower(target) {
			if page.GetAccessAllowed() {
				RespondWithCookie(c, http.StatusNoContent, nil)
				return
			} else {
				RespondWithCookie(c, http.StatusForbidden, gin.H{"message": "access denied"})
				return
			}
		}
	}

	RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "page not found"})
}
