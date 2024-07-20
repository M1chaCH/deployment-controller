package admin

import (
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/rest/admin/pages"
	"github.com/M1chaCH/deployment-controller/users"
	"github.com/gin-gonic/gin"
)

func Init(router *gin.Engine) {
	group := router.Group("/admin")
	group.Use(auth.AdminAuthorisationMiddleware())

	pages.Init(group)
	users.Init(group)
}
