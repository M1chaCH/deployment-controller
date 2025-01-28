package mfa

import (
	"github.com/M1chaCH/deployment-controller/auth/mfa/apptotp"
	"github.com/M1chaCH/deployment-controller/auth/mfa/mailtotp"
	"github.com/M1chaCH/deployment-controller/framework"
)

const MfaTypeApp = "mfa-apptotp"
const MfaTypeMail = "mfa-mailtotp"

func Prepare(loadableTx framework.LoadableTx, userId string, mfaType string) error {
	if mfaType == MfaTypeApp {
		err := apptotp.PrepareTotp(loadableTx, userId, mfaType)
		return err
	}
	if mfaType == MfaTypeMail {
		err := mailtotp.PrepareTotp(loadableTx, userId, mfaType)
		return err
	}
}

func IntialValidate(loadableTx framework.LoadableTx, userId string, mfaType string, code string) (bool, error) {

}

func Validate(loadableTx framework.LoadableTx, userId string, mfaType string, code string, checkValidated bool) (bool, error) {

}

func ClearTokenOfUser(loadableTx framework.LoadableTx, userId string) error {

}

func GetQrImage(loadableTx framework.LoadableTx, userId string) ([]byte, error) {
	// needs to failed if not validated
}

func SendMailTotp(loadableTx framework.LoadableTx, userId string, mail string) {

}
