package rest

import (
	"github.com/gin-gonic/gin"
)

func InitAdminEndpoints(router gin.IRouter) {
	router.GET("/users", getUsers)
	router.POST("/users", postUser)
	router.PUT("/users", putUser)
	router.DELETE("/users/:id", deleteUser)

	router.GET("/pages", getPages)
	router.POST("/pages", postPage)
	router.PUT("/pages", putPage)
	router.DELETE("/pages/:id", deletePage)
}
