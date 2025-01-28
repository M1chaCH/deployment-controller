package mailtotp

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/mail"
	"github.com/pquerna/otp/totp"
	"io"
	"time"
)

type totpEntity struct {
	UserId      string        `db:"user_id"`
	Secret      string        `db:"secret"`
	AccountName string        `db:"account_name"`
	Tries       int           `db:"tries"`
	LastCode    sql.NullInt16 `db:"last_code"`
	Validated   bool          `db:"validated"`
	ValidatedAt sql.NullTime  `db:"validated_at"`
	CreatedAt   time.Time     `db:"created_at"`
}

/*
TODO
current idea:
Does not matter if user MFAs with app or email, both must take the same steps.
1. Prepare
	Mail & App: Generate and store Secret
2. Send
	Mail: GenerateCode from Secret and store code and send in mail to user
	App: Done by the app
3. Initial validate
	Mail & App: Validate the code, if valid set validated
4. General usage
	Mail: Send and later validate
	App: Validate
*/

func PrepareTotp(loadableTx framework.LoadableTx, userId string, userEmail string) error {
	config := framework.Config()

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.JWT.Domain,
		AccountName: "michu-tech - " + userEmail,
		Period:      5 * 60,
	})
	if err != nil {
		return err
	}

	return insertNewMailTotp(loadableTx, totpEntity{
		UserId:      userId,
		Secret:      key.Secret(),
		AccountName: key.AccountName(),
	})
}

func SendTotp(loadableTx framework.LoadableTx, userId string, userEmail string, checkValidated bool) error {
	config := framework.Config()

	token, err := selectMailTotp(loadableTx, userId)
	if err != nil {
		return err
	}

	if checkValidated && !token.Validated {
		return errors.New("token has never been validated")
	}

	code, err := totp.GenerateCode(token.Secret, time.Now().UTC())
	if err != nil {
		return err
	}

	err = updateCurrentToken(loadableTx, userId, code, 0)
	if err != nil {
		return err
	}

	err = mail.SendMail(fmt.Sprintf("totp-%s", userId), userEmail, "Authentication Code", func(writer io.WriteCloser) error {
		return mail.ParseMfaCodeTemplate(writer, mail.MfaCodeMailData{
			AdminMail: config.Mail.Receiver,
			MfaCode:   code,
		})
	})
	return err
}

func InitiallyValidateTotp(loadableTx framework.LoadableTx, userId string, code string) (bool, error) {

}

func ValidateTotp(loadableTx framework.LoadableTx, userId string, code string, checkValidated bool) (bool, error) {

}

func selectMailTotp(loadableTx framework.LoadableTx, userId string) (totpEntity, error) {
	tx, err := loadableTx()
	if err != nil {
		return totpEntity{}, err
	}

	var entities []totpEntity
	err = tx.Select(&entities, "SELECT * FROM user_mail_totp WHERE user_id = $1", userId)
	if err != nil {
		return totpEntity{}, err
	}
	if len(entities) == 0 {
		return totpEntity{}, sql.ErrNoRows
	}
	if len(entities) > 1 {
		return totpEntity{}, errors.New(framework.ErrTooManyTokens)
	}
	return entities[0], nil
}

func insertNewMailTotp(loadableTx framework.LoadableTx, entity totpEntity) error {
	tx, err := loadableTx()
	if err != nil {
		return err
	}

	_, err = tx.NamedExec(`
INSERT INTO public.user_mail_totp (user_id, secret, account_name) 
VALUES (:user_id, :secret, :account_name)
`, entity)

	return err
}

func updateCurrentToken(loadableTx framework.LoadableTx, userId string, code string, tries int) error {
	tx, err := loadableTx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
UPDATE public.user_mail_totp SET last_code = $1, tries = $2 where user_id = $3
`, code, tries, userId)
	return err
}

func removeMailOtp() {

}
