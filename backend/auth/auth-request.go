package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/pageaccess"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
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

	token, found := GetCurrentIdentityToken(c)

	userId := token.UserId
	if !found || userId == "" {
		userId = pageaccess.AnonymousUserId
	}

	userPageAccess, err := pageaccess.LoadUserPageAccess(framework.GetTx(c), userId)
	if err != nil {
		logs.Warn(fmt.Sprintf("Failed to load user page access for user '%s': %v", token.UserId, err))
		RespondWithCookie(c, http.StatusForbidden, gin.H{"message": "access denied"})
		return
	}

	checkTargetAccess(c, technicalName, userPageAccess.Pages)
}

type assertablePage interface {
	GetTechnicalName() string
	GetAccessAllowed() bool
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
