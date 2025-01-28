package rest

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/auth/mfa"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/pageaccess"
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
	UserId     string                      `json:"userId"`
	Mail       string                      `json:"mail"`
	Admin      bool                        `json:"admin"`
	Blocked    bool                        `json:"blocked"`
	Onboard    bool                        `json:"onboard"`
	CreatedAt  time.Time                   `json:"createdAt"`
	LastLogin  time.Time                   `json:"lastLogin"`
	PageAccess []pageaccess.PageAccessPage `json:"pageAccess"`
	Devices    []UserDevicesDto            `json:"devices"`
}

type UserDevicesDto struct {
	UserId             string    `json:"userId" db:"user_id"`
	ClientId           string    `json:"clientId" db:"client_id"`
	DeviceId           string    `json:"deviceId" db:"device_id"`
	Ip                 string    `json:"ip" db:"ip_address"`
	Agent              string    `json:"userAgent" db:"user_agent"`
	City               string    `json:"city" db:"city"`
	Subdivision        string    `json:"subdivision" db:"subdivision"`
	Country            string    `json:"country" db:"country"`
	SystemOrganisation string    `json:"systemOrganisation" db:"system_organisation"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
}

func getUsers(c *gin.Context) {
	result, err := users.LoadUsers(framework.GetTx(c))
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to select all users: %v", err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load users"})
		return
	}

	userIds := make([]string, len(result))
	for i, user := range result {
		userIds[i] = user.Id
	}
	userDevices, err := clients.SelectDevicesByUsers(userIds)
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to select devices of users: %v", err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load users"})
		return
	}

	// don't want so send salt and password
	dtos := make([]AdminUserDto, len(result))
	for i, user := range result {
		pageAccess, err := pageaccess.LoadUserPageAccess(framework.GetTx(c), user.Id)
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to load user page access: %v", err))
			auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load users"})
			return
		}

		devices := make([]UserDevicesDto, 0)
		for _, userDevice := range userDevices {
			if userDevice.UserId == user.Id {
				devices = append(devices, UserDevicesDto{
					UserId:             userDevice.UserId,
					ClientId:           userDevice.ClientId,
					DeviceId:           userDevice.DeviceId,
					Ip:                 userDevice.Ip,
					Agent:              userDevice.Agent,
					City:               userDevice.City.String,
					Subdivision:        userDevice.Subdivision.String,
					Country:            userDevice.Country.String,
					SystemOrganisation: userDevice.SystemOrganisation.String,
					CreatedAt:          time.Time{},
				})
			}
		}

		dtos[i] = AdminUserDto{
			UserId:     user.Id,
			Mail:       user.Mail,
			Admin:      user.Admin,
			Blocked:    user.Blocked,
			Onboard:    user.Onboard,
			CreatedAt:  user.CreatedAt,
			LastLogin:  user.LastLogin,
			PageAccess: pageAccess.Pages,
			Devices:    devices,
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
	MfaType     string   `json:"mfaType"`
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

	err = users.InsertNewUser(framework.GetTx(c), dto.UserId, dto.Mail, hashedPassword, salt, dto.Admin, dto.Blocked, dto.MfaType, dto.AddPages)
	if err != nil {
		logs.Warn(fmt.Sprintf("could not insert new user: %v -> %v", dto.Mail, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to create user"})
		return
	}

	err = mfa.Prepare(framework.GetTx(c), dto.UserId, dto.MfaType)
	if err != nil {
		logs.Warn(fmt.Sprintf("could not prepare token for new user: %v -> %v", dto.Mail, err))
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

	if !currentUser.Onboard && dto.Onboard {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "can't 'admin onboard' user"})
		return
	}

	if currentUser.Onboard && !dto.Onboard {
		err := mfa.ClearTokenOfUser(framework.GetTx(c), dto.UserId)
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to remove token for user: %v -> %v", dto.UserId, err))
			auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to remove token for user"})
			return
		}

		err = mfa.Prepare(framework.GetTx(c), dto.UserId, dto.MfaType)
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to prepare token for user: %v -> %v", dto.UserId, err))
			auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to prepare token for user"})
			return
		}
	}

	err := users.UpdateUser(framework.GetTx(c), dto.UserId, dto.Mail, existingUser.Password, existingUser.Salt, dto.Admin, dto.Blocked, dto.Onboard, existingUser.LastLogin, dto.MfaType, dto.RemovePages, dto.AddPages)
	if err != nil {
		logs.Warn(fmt.Sprintf("could not update user: %v -> %v", dto.Mail, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to update user"})
		return
	}

	pageaccess.DeleteUserPageAccessCache(dto.UserId)
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
