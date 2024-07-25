package rest

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

func InitAdminEndpoints(router *gin.RouterGroup) {
	router.POST("/users", postUser)
}

const emailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

var emailRegex = regexp.MustCompile(emailPattern)

type editUserDto struct {
	UserId       string   `json:"userId" binding:"required"`
	Mail         string   `json:"mail"`
	Password     string   `json:"password"`
	Admin        bool     `json:"admin"`
	Blocked      bool     `json:"blocked"`
	PrivatePages []string `json:"privatePages"`
}

func postUser(c *gin.Context) {
	var dto editUserDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		logs.Info(fmt.Sprintf("failed to bind user from request: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	if dto.Mail == "" || !emailRegex.MatchString(dto.Mail) || dto.Password == "" {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "required data is missing or has wrong format"})
		return
	}

	if users.SimilarUserExists(dto.UserId, dto.Mail) {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "user already exists"})
		return
	}

	hashedPassword, salt, err := framework.SecureHashWithSalt(dto.Password)
	if err != nil {
		logs.Warn(fmt.Sprintf("failed hash password: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "failed to encode password"})
		return
	}

	_, err = users.InsertNewUser(framework.GetTx(c), dto.UserId, dto.Mail, hashedPassword, salt, dto.Admin, dto.Blocked, dto.PrivatePages)
	if err != nil {
		logs.Warn(fmt.Sprintf("could not insert new user: %v -> %v", dto.Mail, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to create user"})
		return
	}

	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "user created"})
}
