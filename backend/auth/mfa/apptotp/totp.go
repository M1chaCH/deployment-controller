package apptotp

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/pquerna/otp/totp"
	"image/png"
	"time"
)

func PrepareTotp(loadableTx framework.LoadableTx, userId string, userEmail string) error {
	config := framework.Config()

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.JWT.Domain,
		AccountName: userEmail,
	})
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return err
	}
	err = png.Encode(&buf, img)
	if err != nil {
		return err
	}

	entity := totpEntity{
		UserId:      userId,
		Secret:      key.Secret(),
		AccountName: key.AccountName(),
		Image:       buf.Bytes(),
		Validated:   false,
	}

	return insertNewTotp(loadableTx, entity)
}

func LoadTotpImage(loadableTx framework.LoadableTx, userId string) ([]byte, error) {
	res, err := selectTotp(loadableTx, userId)
	if err != nil {
		return nil, err
	}

	return res.Image, nil
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
	entity, err := selectTotp(loadableTx, userId)
	if err != nil {
		return false, err
	}

	if checkValidated && !entity.Validated {
		return false, errors.New(framework.ErrNotValidated)
	}

	valid := totp.Validate(code, entity.Secret)
	return valid, nil
}

func RemoveTotpForUser(loadableTx framework.LoadableTx, userId string) error {
	tx, err := loadableTx()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM user_totp WHERE user_id = $1", userId)

	return err
}

type totpEntity struct {
	UserId      string       `db:"user_id"`
	Secret      string       `db:"secret"`
	AccountName string       `db:"account_name"`
	Image       []byte       `db:"image"`
	Validated   bool         `db:"validated"`
	CreatedAt   time.Time    `db:"created_at"`
	ValidatedAt sql.NullTime `db:"validated_at"`
}

func selectTotp(loadableTx framework.LoadableTx, userId string) (totpEntity, error) {
	tx, err := loadableTx()
	if err != nil {
		return totpEntity{}, err
	}

	var entities []totpEntity
	err = tx.Select(&entities, "SELECT * FROM user_totp WHERE user_id = $1", userId)
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

func insertNewTotp(txLoader framework.LoadableTx, entity totpEntity) error {
	tx, err := txLoader()
	if err != nil {
		return err
	}

	_, err = tx.NamedExec(`
INSERT INTO public.user_totp (user_id, secret, account_name, image, validated)
VALUES (:user_id, :secret, :account_name, :image, false)
`, entity)

	return err
}

func markTotpAsValid(txLoader framework.LoadableTx, userId string) error {
	tx, err := txLoader()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
UPDATE user_totp SET validated = true, validated_at = $1 WHERE user_id = $2
`, time.Now(), userId)

	return err
}
