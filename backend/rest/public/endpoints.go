package public

import (
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/users"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

func Init(router *gin.Engine) {
	router.GET("/login", getLoggedInUser)
	router.POST("/login", postLogin)
	router.PUT("/login", putChangePassword)

	router.GET("/overview")
}

type userInfoDto struct {
	Mail         string   `json:"mail"`
	Admin        bool     `json:"admin"`
	PrivatePages []string `json:"privatePages"`
	LoginState   string   `json:"loginStatus"`
}

func getLoggedInUser(c *gin.Context) {
	token, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "no user info found"})
		return
	}

	dto := userInfoDto{
		Mail:         token.Mail,
		Admin:        token.Admin,
		PrivatePages: strings.Split(token.PrivatePages, auth.PrivatePagesDelimiter),
		LoginState:   token.LoginState,
	}
	c.JSON(http.StatusOK, dto)
}

func postLogin(c *gin.Context) {
	var dto auth.LoginDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		log.Printf("failed to bind dto from request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "login form invalid"})
		return
	}

	if dto.Mail == "" || dto.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "login form invalid"})
		return
	}

	user, err := users.SelectUserByMail(dto.Mail)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	hashedPassword := framework.SecureHash(dto.Password, user.Salt)
	if hashedPassword != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	auth.HandleLoginWithValidCredentials(c, user)
}
