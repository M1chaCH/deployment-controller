package rest

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/pages"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/M1chaCH/deployment-controller/mail"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"regexp"
)

func InitOpenEndpoints(router gin.IRouter) {
	router.GET("/login", getCurrentUser)
	router.POST("/login", postLogin)
	router.PUT("/login", putUserPassword)
	router.GET("/login/onboard/img", getOnboardingTokenImg)
	router.POST("/login/onboard", postCompleteOnboarding)
	router.GET("/pages", getOverviewPages)
	router.POST("/contact", postContact)
}

var digitRegex = regexp.MustCompile(`\d`)
var smallLetterRegex = regexp.MustCompile(`[a-z]`)
var largeLetterRegex = regexp.MustCompile(`[A-Z]`)

type loginDto struct {
	Mail     string `json:"mail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func postLogin(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "missing user info"})
		return
	}

	// TODO, do we really want this? this will let a user through even if his password is incorrect.
	if idToken.LoginState == auth.LoginStateLoggedIn || idToken.LoginState == auth.LoginStateOnboardingWaiting {
		auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": auth.LoginStateLoggedIn})
		return
	}

	if idToken.LoginState == auth.LoginStateTwofactorWaiting {
		auth.RespondWithCookie(c, http.StatusNotImplemented, gin.H{"message": "feature not yet implemented..."})
		return
	}

	var dto loginDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "login form invalid"})
		return
	}

	if dto.Mail == "" || dto.Password == "" {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "login form invalid"})
		return
	}

	user, ok := users.LoadUserByMail(framework.GetTx(c), dto.Mail)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	if user.Blocked {
		logs.Info(fmt.Sprintf("blocked user tryed to login: %s", user.Id))
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	hashedPassword := framework.SecureHash(dto.Password, user.Salt)
	if hashedPassword != user.Password {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	state := auth.HandleLoginWithValidCredentials(c, user)
	if state != auth.LoginStateLoggedOut {
		auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": state})
	}
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

	body := gin.H{
		"userId":     user.Id,
		"mail":       user.Mail,
		"admin":      user.Admin,
		"onboard":    user.Onboard,
		"loginState": idToken.LoginState,
	}
	auth.RespondWithCookie(c, http.StatusOK, body)
}

type changePasswordDto struct {
	UserId      string `json:"userId" binding:"required"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword" binding:"required"`
	Token       string `json:"token"`
}

func putUserPassword(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	var dto changePasswordDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		logs.Info(fmt.Sprintf("failed to bind data from change password request: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	ok = changePasswordHandler(c, dto, idToken, false)
	if ok {
		auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "password updated"})
	}
}

func getOnboardingTokenImg(c *gin.Context) {
	totpEntity, ok := handleGetOnboardingToken(c)
	if !ok {
		return
	}

	auth.AppendJwtToken(c)
	c.Header("Content-Type", "image/png")
	_, err := c.Writer.Write(totpEntity.Image)
	if err != nil {
		auth.AbortWithCooke(c, http.StatusInternalServerError, "failed to write image")
		return
	}
}

func handleGetOnboardingToken(c *gin.Context) (auth.MfaTokenEntity, bool) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return auth.MfaTokenEntity{}, false
	}

	if idToken.LoginState != auth.LoginStateOnboardingWaiting {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "not correct timing"})
		return auth.MfaTokenEntity{}, false
	}

	totpEntity, err := auth.LoadTotpForUser(framework.GetTx(c), idToken.UserId)
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to load totp for user: %s - %v", idToken.UserId, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load token"})
		return auth.MfaTokenEntity{}, false
	}

	if totpEntity.Validated {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "already validated, you should be onboard"})
		return auth.MfaTokenEntity{}, false
	}

	return totpEntity, true
}

func postCompleteOnboarding(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	var dto changePasswordDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		logs.Info(fmt.Sprintf("failed to bind data from onboarding request: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	valid, err := auth.InitiallyValidateToken(framework.GetTx(c), idToken.UserId, dto.Token)
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to validate token: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "invalid token"})
		return
	}
	if !valid {
		auth.RespondWithCookie(c, http.StatusForbidden, gin.H{"message": "invalid token"})
		return
	}

	ok = changePasswordHandler(c, dto, idToken, true)
	if ok {
		idToken.LoginState = auth.LoginStateLoggedIn
		auth.SetCurrentIdentityToken(c, idToken)
		auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "onboarding complete!"})
	}
}

type overviewPagesDto struct {
	PageTitle       string `json:"pageTitle" binding:"required"`
	PageDescription string `json:"pageDescription" binding:"required"`
	PageUrl         string `json:"pageUrl" binding:"required"`
	PrivatePage     bool   `json:"privatePage"`
	AccessAllowed   bool   `json:"accessAllowed"`
}

func getOverviewPages(c *gin.Context) {
	user, found := auth.GetCurrentUser(c)

	result := make([]overviewPagesDto, len(user.Pages))
	if found {
		for i, page := range user.Pages {
			result[i] = overviewPagesDto{
				PageTitle:       page.Title,
				PageDescription: page.Description,
				PageUrl:         page.Url,
				PrivatePage:     page.Private,
				AccessAllowed:   page.AccessAllowed,
			}
		}

		auth.RespondWithCookie(c, http.StatusOK, result)
		return
	}

	allPages, err := pages.LoadPages(framework.GetTx(c))
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to load all pages for overview: %v", err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "could not load pages"})
		return
	}

	result = make([]overviewPagesDto, len(allPages))
	for i, page := range allPages {
		result[i] = overviewPagesDto{
			PageTitle:       page.Title,
			PageDescription: page.Description,
			PageUrl:         page.Url,
			PrivatePage:     page.PrivatePage,
			AccessAllowed:   !page.PrivatePage,
		}
	}

	auth.RespondWithCookie(c, http.StatusOK, result)
}

func changePasswordHandler(c *gin.Context, dto changePasswordDto, idToken auth.IdentityToken, onboarding bool) bool {
	if onboarding && idToken.LoginState != auth.LoginStateOnboardingWaiting {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "onboarding failed, not logged in / already onboard?"})
		return false
	}

	if !onboarding && idToken.LoginState != auth.LoginStateLoggedIn {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return false
	}

	if !idToken.Admin && idToken.UserId != dto.UserId {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return false
	}

	user, ok := users.LoadUserById(framework.GetTx(c), dto.UserId)
	if !ok {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "user does not exist"})
		return false
	}

	if !idToken.Admin {
		hashedOldPassword := framework.SecureHash(dto.OldPassword, user.Salt)
		if hashedOldPassword != user.Password {
			auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "old password incorrect!"})
			return false
		}
	}

	if len(dto.NewPassword) < 8 || !digitRegex.MatchString(dto.NewPassword) || !smallLetterRegex.MatchString(dto.NewPassword) || !largeLetterRegex.MatchString(dto.NewPassword) {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "new password must contain at least 1 large letter, 1 small letter and one digit. Also the password must be at least 8 chars long."})
		return false
	}

	hashedNewPassword, salt, err := framework.SecureHashWithSalt(dto.NewPassword)
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to hash new password for user id: %s -> %v", dto.UserId, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "could not convert password!"})
		return false
	}

	_, err = users.UpdateUser(framework.GetTx(c), user.Id, user.Mail, hashedNewPassword, salt, user.Admin, user.Blocked, true, user.LastLogin, make([]string, 0), make([]string, 0))
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to save new password for user: %s -> %v", dto.UserId, err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "could not save changes to user!"})
		return false
	}

	return true
}

type contactDto struct {
	Mail    string `json:"mail"`
	Message string `json:"message"`
}

func postContact(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	var dto contactDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		logs.Info(fmt.Sprintf("failed to bind data from contact request: %v", err))
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	if len(dto.Message) > 1000 {
		auth.RespondWithCookie(c, http.StatusRequestEntityTooLarge, gin.H{"message": "message too long"})
		return
	}

	deviceId, err := clients.LookupDeviceId(idToken.Issuer, idToken.OriginIp, idToken.OriginAgent)
	if err != nil {
		logs.Warn(fmt.Sprintf("device of request not found: %v -- clientId:%s ip:%s agent:%s", err, idToken.Issuer, idToken.OriginIp, idToken.OriginAgent))
		deviceId = "not found: " + err.Error()
	}

	err = mail.SendMailToAdmin(idToken.Issuer, "michu-tech Contact request", func(writer io.WriteCloser) error {
		return mail.ParseContactRequestTemplate(writer, mail.ContactRequestMailData{
			ClientId: idToken.Issuer,
			DeviceId: deviceId,
			Sender:   dto.Mail,
			Message:  dto.Message,
		})
	})

	if err != nil {
		if errors.Is(err, mail.TooManyMailsError) {
			logs.Warn(fmt.Sprintf("mail threshold was reached by client: %s", idToken.Issuer))
			auth.RespondWithCookie(c, http.StatusTooManyRequests, gin.H{"message": "too many mails sent, wait some time and try again"})
			return
		}

		logs.Warn(fmt.Sprintf("failed to send mail: %v", err))
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to send mail"})
		return
	}

	auth.RespondWithCookie(c, http.StatusAccepted, gin.H{"message": "sent mail"})
}
