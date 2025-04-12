package mailtotp

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/mail"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"io"
	"time"
)

const totpPeriodeSeconds = 5 * 60

type totpEntity struct {
	UserId      string        `db:"user_id"`
	Secret      string        `db:"secret"`
	AccountName string        `db:"account_name"`
	Tries       int           `db:"tries"`
	LastCode    sql.NullInt32 `db:"last_code"`
	Validated   bool          `db:"validated"`
	ValidatedAt sql.NullTime  `db:"validated_at"`
	CreatedAt   time.Time     `db:"created_at"`
}

/*
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
		AccountName: userEmail,
		Period:      totpPeriodeSeconds,
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

	code, err := totp.GenerateCodeCustom(token.Secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    totpPeriodeSeconds,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return err
	}

	err = updateCurrentToken(userId, code, 0)
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
	codeValid, err := ValidateTotp(loadableTx, userId, code, false)
	if err != nil || !codeValid {
		return codeValid, err
	}

	err = markTotpAsValid(loadableTx, userId)
	return codeValid, err
}

func ValidateTotp(loadableTx framework.LoadableTx, userId string, code string, checkValidated bool) (bool, error) {
	entity, err := selectMailTotp(loadableTx, userId)
	if err != nil {
		return false, err
	}

	if entity.Tries+1 > 5 {
		return false, errors.New("totp expired")
	}

	if checkValidated && !entity.Validated {
		return false, errors.New(framework.ErrNotValidated)
	}

	valid, err := totp.ValidateCustom(code, entity.Secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    totpPeriodeSeconds,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if valid {
		err = updateCurrentToken(userId, code, 0)
	} else {
		err = updateCurrentToken(userId, code, entity.Tries+1)
	}

	return valid, err
}

func RemoveTotpForUser(loadableTx framework.LoadableTx, userId string) error {
	tx, err := loadableTx()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM user_mail_totp WHERE user_id = $1", userId)

	return err
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

func updateCurrentToken(userId string, code string, tries int) error {
	db := framework.DB()

	_, err := db.Exec(`
UPDATE public.user_mail_totp SET last_code = $1, tries = $2 where user_id = $3
`, code, tries, userId)
	return err
}

func markTotpAsValid(txLoader framework.LoadableTx, userId string) error {
	tx, err := txLoader()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
UPDATE user_mail_totp SET validated = true, validated_at = $1 WHERE user_id = $2
`, time.Now(), userId)

	return err
}
