package mfa

import (
	"errors"
	"github.com/M1chaCH/deployment-controller/auth/mfa/apptotp"
	"github.com/M1chaCH/deployment-controller/auth/mfa/mailtotp"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
)

const TypeApp = "mfa-apptotp"
const TypeMail = "mfa-mailtotp"
const ErrMfaTypeUnknown = "unknown MFA type"

func Prepare(loadableTx framework.LoadableTx, userId string, mfaType string) error {
	user, found := users.LoadUserById(loadableTx, userId)
	if !found {
		return errors.New("unknown user")
	}

	if mfaType == TypeApp {
		err := apptotp.PrepareTotp(loadableTx, userId, user.Mail)
		return err
	}
	if mfaType == TypeMail {
		err := mailtotp.PrepareTotp(loadableTx, userId, user.Mail)
		return err
	}

	return errors.New(ErrMfaTypeUnknown)
}

func InitialValidate(loadableTx framework.LoadableTx, userId string, mfaType string, code string) (bool, error) {
	if mfaType == TypeApp {
		return apptotp.InitiallyValidateTotp(loadableTx, userId, code)
	}
	if mfaType == TypeMail {
		return mailtotp.InitiallyValidateTotp(loadableTx, userId, code)
	}

	return false, errors.New(ErrMfaTypeUnknown)
}

func Validate(loadableTx framework.LoadableTx, userId string, mfaType string, code string, checkValidated bool) (bool, error) {
	if mfaType == TypeApp {
		return apptotp.ValidateTotp(loadableTx, userId, code, checkValidated)
	}
	if mfaType == TypeMail {
		return mailtotp.ValidateTotp(loadableTx, userId, code, checkValidated)
	}

	return false, errors.New(ErrMfaTypeUnknown)
}

func ClearTokenOfUser(loadableTx framework.LoadableTx, userId string) error {
	err := apptotp.RemoveTotpForUser(loadableTx, userId)
	err = mailtotp.RemoveTotpForUser(loadableTx, userId)
	return err
}

func GetQrImage(loadableTx framework.LoadableTx, userId string) ([]byte, error) {
	return apptotp.LoadTotpImage(loadableTx, userId)
}

func HandleChangedTotpType(loadableTx framework.LoadableTx, userId string, newType string) error {
	if newType != TypeApp && newType != TypeMail {
		return errors.New(ErrMfaTypeUnknown)
	}

	err := ClearTokenOfUser(loadableTx, userId)
	if err != nil {
		return err
	}

	err = Prepare(loadableTx, userId, newType)
	return err
}

func SendMailTotp(loadableTx framework.LoadableTx, userId string, mail string, checkValidated bool) error {
	return mailtotp.SendTotp(loadableTx, userId, mail, checkValidated)
}
