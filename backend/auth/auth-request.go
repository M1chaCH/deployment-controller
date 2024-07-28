package auth

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/pages"
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

	user, found := GetCurrentUser(c)

	if found {
		checkTargetAccess(c, technicalName, user.Pages)
		return
	}

	allPages, err := pages.LoadPages(framework.GetTx(c))
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to load pages, auth request failed: %v", err))
		RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "page not found"})
		return
	}

	checkTargetAccess(c, technicalName, allPages)
}

type assertablePage interface {
	GetTechnicalName() string
	GetAccessAllowed() bool
}

func checkTargetAccess[T assertablePage](c *gin.Context, target string, allPages []T) {
	for _, page := range allPages {
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
