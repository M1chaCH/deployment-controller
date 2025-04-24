package rest

import (
	"errors"
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/auth/mfa"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/pageaccess"
	"github.com/M1chaCH/deployment-controller/data/pages"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/config"
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
	router.GET("/login/onboard/url", getOnboardingTokenUrl)
	router.POST("/login/onboard", postCompleteOnboarding)
	router.POST("/login/mfa/mail", postSendMfaToken)
	router.POST("/login/mfa", postMfaCheck)
	router.PUT("/login/mfa/type", putChangeMfaType)
	router.GET("/pages", getOverviewPages)
	router.POST("/contact", postContact)
	router.POST("/logout", postLogout)
}

var digitRegex = regexp.MustCompile(`\d`)
var smallLetterRegex = regexp.MustCompile(`[a-z]`)
var largeLetterRegex = regexp.MustCompile(`[A-Z]`)

type loginDto struct {
	Mail     string `json:"mail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func postLogin(c *gin.Context) {
	var dto loginDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "login form invalid"})
		return
	}

	if dto.Mail == "" || dto.Password == "" {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "login form invalid"})
		return
	}

	user, ok := users.LoadUserByMail(c, dto.Mail)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	if user.Blocked {
		logs.Info(c, "blocked user tryed to login: %s", user.Id)
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	hashedPassword := framework.SecureHash(dto.Password, user.Salt)
	if hashedPassword != user.Password {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "login failed"})
		return
	}

	auth.HandleAndCompleteLogin(c, user)
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
		"mfaType":    user.MfaType,
		"loginState": idToken.LoginState,
	}
	auth.RespondWithCookie(c, http.StatusOK, body)
}

type changePasswordDto struct {
	UserId      string `json:"userId" binding:"required"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword" binding:"required"`
	Token       string `json:"token"`
	MfaType     string `json:"mfaType"`
}

func putUserPassword(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	var dto changePasswordDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		logs.Info(c, "failed to bind data from change password request: %v", err)
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	ok = changePasswordHandler(c, dto, idToken, false)
	if ok {
		auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "password updated"})
	}
}

func getOnboardingTokenImg(c *gin.Context) {
	image, _, ok := handleGetOnboardingToken(c)
	if !ok {
		return
	}

	auth.AppendJwtToken(c)
	c.Header("Content-Type", "image/png")
	_, err := c.Writer.Write(image)
	if err != nil {
		auth.AbortWithCooke(c, http.StatusInternalServerError, "failed to write image")
		return
	}
}

func getOnboardingTokenUrl(c *gin.Context) {
	_, url, ok := handleGetOnboardingToken(c)
	if !ok {
		return
	}

	auth.AppendJwtToken(c)
	auth.RespondWithCookie(c, http.StatusOK, gin.H{"url": url})
}

func handleGetOnboardingToken(c *gin.Context) ([]byte, string, bool) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return nil, "", false
	}

	if idToken.LoginState != auth.LoginStateOnboardingWaiting && idToken.LoginState != auth.LoginStateLoggedIn {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "not correct timing"})
		return nil, "", false
	}

	image, url, err := mfa.GetQrImageAndUrl(framework.GetTx(c), idToken.UserId)
	if err != nil {
		logs.Warn(c, "failed to load totp for user: %s - %v", idToken.UserId, err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load token"})
		return nil, "", false
	}

	return image, url, true
}

func postCompleteOnboarding(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	var dto changePasswordDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		logs.Info(c, "failed to bind data from onboarding request: %v", err)
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	tx := framework.GetTx(c)

	valid, err := mfa.InitialValidate(tx, idToken.UserId, dto.MfaType, dto.Token)
	if err != nil {
		logs.Warn(c, "failed to validate token: %v", err)
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "invalid token"})
		return
	}
	if !valid {
		auth.RespondWithCookie(c, http.StatusForbidden, gin.H{"message": "invalid token"})
		return
	}

	client, found, err := clients.LoadClientInfo(c, idToken.Issuer)
	if err != nil || !found {
		logs.Warn(c, "failed to load client info: %v / found: %v", err, found)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "onboarding failed"})
		return
	}

	device, found := clients.GetCurrentDevice(client, idToken.OriginIp, idToken.OriginAgent)
	err = clients.MarkDeviceAsValidated(c, idToken.Issuer, device.Id)
	if err != nil {
		logs.Warn(c, "failed to mark device as validated while onboarding: client: %s - device: %s", client.Id, device.Id)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "onboarding failed"})
		return
	}

	ok = changePasswordHandler(c, dto, idToken, true)
	if !ok {
		return
	}

	user, found := users.LoadUserById(c, idToken.UserId)
	if !found {
		logs.Warn(c, "failed to load user with id, not found: %s", idToken.UserId)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "onboarding failed"})
		return
	}

	pageAccess, err := pageaccess.LoadUserPageAccess(framework.GetTx(c), idToken.UserId)
	if err != nil {
		logs.Warn(c, "failed to load user page access: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "onboarding failed"})
		return
	}

	pagesString := ""
	for i, page := range pageAccess.Pages {
		if !page.GetAccessAllowed() {
			continue
		}

		pagesString += page.TechnicalName

		if i != len(pageAccess.Pages)-1 {
			pagesString += ", "
		}
	}

	err = mail.SendMailToAdmin(c, mail.TechnicalNoThrottle, "michu-tech Onboarding Complete", func(writer io.WriteCloser) error {
		return mail.ParseOnboardingCompleteTemplate(writer, mail.OnboardingCompleteMailData{
			UserMail:       user.Mail,
			UserPageAccess: pagesString,
		})
	})

	idToken.LoginState = auth.LoginStateLoggedIn
	auth.SetCurrentIdentityToken(c, idToken)
	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "onboarding complete!"})
}

func postSendMfaToken(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	if idToken.LoginState != auth.LoginStateTwofactorWaiting && idToken.LoginState != auth.LoginStateOnboardingWaiting {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "not correct timing"})
		return
	}

	user, found := users.LoadUserById(c, idToken.UserId)
	if !found {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "user does not exist"})
		return
	}

	if user.MfaType != mfa.TypeMail {
		auth.RespondWithCookie(c, http.StatusForbidden, gin.H{"message": "mfa not received via Mail"})
		return
	}

	checkValidated := true
	checkValidatedQuery := c.Query("onboarding")
	if checkValidatedQuery == "false" {
		checkValidated = false
	}

	err := mfa.SendMailTotp(c, user.Id, user.Mail, checkValidated)
	if err != nil {
		logs.Warn(c, "failed to send totp for user: %s - %v", idToken.UserId, err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to send totp"})
		return
	}

	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "sent"})
	return
}

type mfaCheckDto struct {
	Token string `json:"token"`
}

func postMfaCheck(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	if idToken.LoginState != auth.LoginStateTwofactorWaiting {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "invalid state"})
		return
	}

	var dto mfaCheckDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "mfa check form invalid"})
		return
	}

	auth.HandleAndCompleteMfaVerification(c, idToken, dto.Token)
}

type changeMfaTypeDto struct {
	UserId  string `json:"userId"`
	MfaType string `json:"mfaType"`
}

func putChangeMfaType(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)

	if !ok {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "request unauthorized"})
		return
	}

	if idToken.LoginState == auth.LoginStateLoggedOut || idToken.LoginState == auth.LoginStateTwofactorWaiting {
		auth.RespondWithCookie(c, http.StatusUnauthorized, gin.H{"message": "invalid state"})
		return
	}

	var dto changeMfaTypeDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "mfa check form invalid"})
		return
	}

	if dto.UserId != idToken.UserId && !idToken.Admin {
		auth.RespondWithCookie(c, http.StatusForbidden, gin.H{"message": "forbidden"})
		return
	}

	user, found := users.LoadUserById(c, dto.UserId)
	if !found {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "user does not exist"})
		return
	}

	if dto.MfaType != mfa.TypeMail && dto.MfaType != mfa.TypeApp {
		auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "mfa type unknown"})
		return
	}

	if user.MfaType == dto.MfaType {
		auth.RespondWithCookie(c, http.StatusNoContent, gin.H{"message": "mfa type the same, nothing changed"})
		return
	}

	err := users.UpdateUser(c, dto.UserId, user.Mail, user.Password, user.Salt, user.Admin, user.Blocked, user.Onboard, user.LastLogin, dto.MfaType, make([]string, 0), make([]string, 0))
	if err != nil {
		logs.Warn(c, "failed to update user: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	err = mfa.HandleChangedTotpType(c, dto.UserId, dto.MfaType)
	if err != nil {
		logs.Warn(c, "failed to change totp type: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "ok"})
}

type overviewPagesDto struct {
	PageTitle       string `json:"pageTitle" binding:"required"`
	PageDescription string `json:"pageDescription" binding:"required"`
	PageUrl         string `json:"pageUrl" binding:"required"`
	PrivatePage     bool   `json:"privatePage"`
	AccessAllowed   bool   `json:"accessAllowed"`
}

func getOverviewPages(c *gin.Context) {
	user, userFound := auth.GetCurrentUser(c)
	token, tokenFound := auth.GetCurrentIdentityToken(c)

	userId := user.Id
	if !userFound || !tokenFound || token.LoginState != auth.LoginStateLoggedIn {
		userId = pageaccess.AnonymousUserId
	}

	pageEntities, err := pages.LoadPages(framework.GetTx(c))
	if err != nil {
		logs.Warn(c, "failed to load pages: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load pages"})
		return
	}

	pageAccess, err := pageaccess.LoadUserPageAccess(framework.GetTx(c), userId)
	if err != nil {
		logs.Warn(c, "failed to load user page access: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to load page access"})
		return
	}

	result := make([]overviewPagesDto, len(pageEntities))
	for i, p := range pageEntities {
		hasAccess := !p.PrivatePage

		for _, pa := range pageAccess.Pages {
			if pa.PageId == p.Id {
				hasAccess = pa.Access
				break
			}
		}

		result[i] = overviewPagesDto{
			PageTitle:       p.Title,
			PageDescription: p.Description,
			PageUrl:         p.Url,
			PrivatePage:     p.PrivatePage,
			AccessAllowed:   hasAccess,
		}
	}

	auth.RespondWithCookie(c, http.StatusOK, result)
}

func changePasswordHandler(c *gin.Context, dto changePasswordDto, idToken auth.IdentityToken, onboarding bool) bool {
	if onboarding && idToken.LoginState != auth.LoginStateOnboardingWaiting && idToken.LoginState != auth.LoginStateLoggedIn {
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

	user, ok := users.LoadUserById(c, dto.UserId)
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
		logs.Warn(c, "failed to hash new password for user id: %s -> %v", dto.UserId, err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "could not convert password!"})
		return false
	}

	err = users.UpdateUser(c, user.Id, user.Mail, hashedNewPassword, salt, user.Admin, user.Blocked, true, user.LastLogin, user.MfaType, make([]string, 0), make([]string, 0))
	if err != nil {
		logs.Warn(c, "failed to save new password for user: %s -> %v", dto.UserId, err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "could not save changes to user!"})
		return false
	}

	pageaccess.DeleteUserPageAccessCache(user.Id)

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
		logs.Info(c, "failed to bind data from contact request: %v", err)
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "data has invalid format"})
		return
	}

	if len(dto.Message) > config.Config().Mail.MaxMessageLength {
		auth.RespondWithCookie(c, http.StatusRequestEntityTooLarge, gin.H{"message": "message too long"})
		return
	}

	deviceId, err := clients.LookupDeviceId(c, idToken.Issuer, idToken.OriginIp, idToken.OriginAgent)
	if err != nil {
		logs.Warn(c, "device of request not found: %v -- clientId:%s ip:%s agent:%s", err, idToken.Issuer, idToken.OriginIp, idToken.OriginAgent)
		deviceId = "not found: " + err.Error()
	}

	// todo cleanup this api, create the function in the mail package, not as the param
	err = mail.SendMailToAdmin(c, idToken.Issuer, "michu-tech Contact request", func(writer io.WriteCloser) error {
		return mail.ParseContactRequestTemplate(writer, mail.ContactRequestMailData{
			ClientId: idToken.Issuer,
			DeviceId: deviceId,
			Sender:   dto.Mail,
			Message:  dto.Message,
		})
	})

	if err != nil {
		if errors.Is(err, mail.TooManyMailsError) {
			logs.Warn(c, "mail threshold was reached by client: %s", idToken.Issuer)
			auth.RespondWithCookie(c, http.StatusTooManyRequests, gin.H{"message": "too many mails sent, wait some time and try again"})
			return
		}

		logs.Warn(c, "failed to send mail: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to send mail"})
		return
	}

	auth.RespondWithCookie(c, http.StatusAccepted, gin.H{"message": "sent mail"})
}

func postLogout(c *gin.Context) {
	idToken, ok := auth.GetCurrentIdentityToken(c)
	if !ok {
		auth.RespondWithCookie(c, http.StatusNoContent, gin.H{})
		return
	}

	idToken.LoginState = auth.LoginStateLoggedOut
	idToken.UserId = ""
	idToken.Mail = ""

	auth.SetCurrentIdentityToken(c, idToken)
	auth.RespondWithCookie(c, http.StatusNoContent, gin.H{})
}
