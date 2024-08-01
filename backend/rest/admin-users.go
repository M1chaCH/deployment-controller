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
	"time"
)

const emailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

var emailRegex = regexp.MustCompile(emailPattern)

type AdminUserDto struct {
	UserId     string          `json:"userId"`
	Mail       string          `json:"mail"`
	Admin      bool            `json:"admin"`
	Blocked    bool            `json:"blocked"`
	Onboard    bool            `json:"onboard"`
	CreatedAt  time.Time       `json:"createdAt"`
	LastLogin  time.Time       `json:"lastLogin"`
	PageAccess []PageAccessDto `json:"pageAccess"`
}

type PageAccessDto struct {
	PageId        string `json:"pageId"`
	TechnicalName string `json:"technicalName"`
	AccessAllowed bool   `json:"accessAllowed"`
	PagePrivate   bool   `json:"pagePrivate"`
}

func getUsers(c *gin.Context) {
	result, err := users.LoadUsers(framework.GetTx(c))
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to select all users: %v", err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load users"})
		return
	}

	// don't want so send salt and password
	dtos := make([]AdminUserDto, len(result))
	for i, user := range result {
		pageAccess := make([]PageAccessDto, 0)
		for _, page := range user.Pages {
			pageAccess = append(pageAccess, PageAccessDto{
				PageId:        page.PageId,
				TechnicalName: page.TechnicalName,
				AccessAllowed: page.AccessAllowed,
				PagePrivate:   page.Private,
			})
		}

		dtos[i] = AdminUserDto{
			UserId:     user.Id,
			Mail:       user.Mail,
			Admin:      user.Admin,
			Blocked:    user.Blocked,
			Onboard:    user.Onboard,
			CreatedAt:  user.CreatedAt,
			LastLogin:  user.LastLogin,
			PageAccess: pageAccess,
		}
	}

	auth.RespondWithCookie(c, http.StatusOK, dtos)
}

type editUserDto struct {
	UserId      string   `json:"userId" binding:"required"`
	Mail        string   `json:"mail" binding:"required"`
	Password    string   `json:"password,omitempty"`
	Admin       bool     `json:"admin"`
	Blocked     bool     `json:"blocked"`
	Onboard     bool     `json:"onboard,omitempty"`
	AddPages    []string `json:"addPages,omitempty"`
	RemovePages []string `json:"removePages,omitempty"`
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

	if users.SimilarUserExists(framework.GetTx(c), dto.UserId, dto.Mail) {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "user already exists"})
		return
	}

	hashedPassword, salt, err := framework.SecureHashWithSalt(dto.Password)
	if err != nil {
		logs.Warn(fmt.Sprintf("failed hash password: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "failed to encode password"})
		return
	}

	_, err = users.InsertNewUser(framework.GetTx(c), dto.UserId, dto.Mail, hashedPassword, salt, dto.Admin, dto.Blocked, dto.AddPages)
	if err != nil {
		logs.Warn(fmt.Sprintf("could not insert new user: %v -> %v", dto.Mail, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to create user"})
		return
	}

	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "user created"})
}

func putUser(c *gin.Context) {
	var dto editUserDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		logs.Info(fmt.Sprintf("failed to bind user from request: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	existingUser, found := users.LoadUserById(framework.GetTx(c), dto.UserId)
	if !found {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	if users.MailExists(framework.GetTx(c), dto.Mail, dto.UserId) {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "user with mail already exists"})
		return
	}

	currentUser, found := auth.GetCurrentUser(c)
	if !found {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "no user info provided"})
		return
	}

	if !dto.Admin && currentUser.Id == dto.UserId {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "can't remove your own admin access"})
		return
	}

	if dto.Blocked && currentUser.Id == dto.UserId {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "can't block your own access"})
		return
	}

	_, err := users.UpdateUser(framework.GetTx(c), dto.UserId, dto.Mail, existingUser.Password, existingUser.Salt, dto.Admin, dto.Blocked, dto.Onboard, existingUser.LastLogin, dto.RemovePages, dto.AddPages)
	if err != nil {
		logs.Warn(fmt.Sprintf("could not update user: %v -> %v", dto.Mail, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to update user"})
		return
	}

	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "user updated"})
}

func deleteUser(c *gin.Context) {
	tx := framework.GetTx(c)
	userId := c.Param("id")
	user, found := users.LoadUserById(tx, userId)
	if !found {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	if user.Admin && !users.DifferentAdminExists(tx, user.Id) {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "can't remove last admin"})
		return
	}

	currentUser, found := auth.GetCurrentUser(c)
	if !found {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "no user info provided"})
		return
	}
	if currentUser.Id == userId {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "can't remove your own user"})
		return
	}

	err := users.DeleteUser(tx, userId)
	if err != nil {
		logs.Warn(fmt.Sprintf("could not delete user: %v -> %v", userId, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to delete user"})
		return
	}

	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "user deleted"})
}
