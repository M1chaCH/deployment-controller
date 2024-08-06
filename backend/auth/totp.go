package auth

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/pquerna/otp/totp"
	"image/png"
	"time"
)

/*
TODO
- change admin page onboard edit
	- remove checkbox
	- add button "reset onboard"
		- remove totp registration
- integrate TOTP into login process
	- probably need to add "verified" flag to devices
- implement functionality "receive TOTP via mail"
- notify admin via mail when someone completed onboarding
*/

func PrepareToken(loadableTx framework.LoadableTx, userId string, userEmail string) (MfaTokenEntity, error) {
	config := framework.Config()

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.JWT.Domain,
		AccountName: "michu-tech - " + userEmail,
	})
	if err != nil {
		return MfaTokenEntity{}, err
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return MfaTokenEntity{}, err
	}
	err = png.Encode(&buf, img)
	if err != nil {
		return MfaTokenEntity{}, err
	}

	entity := MfaTokenEntity{
		UserId:      userId,
		Secret:      key.Secret(),
		AccountName: key.AccountName(),
		Image:       buf.Bytes(),
		Validated:   false,
	}

	err = insertNewToken(loadableTx, entity)
	return entity, err
}

func LoadTotpForUser(loadableTx framework.LoadableTx, userId string) (MfaTokenEntity, error) {
	return selectToken(loadableTx, userId)
}

const ErrTooManyTokens = "too many tokens found per user"
const ErrNotValidated = "token is not validated"

func InitiallyValidateToken(loadableTx framework.LoadableTx, userId string, code string) (bool, error) {
	codeValid, err := ValidateToken(loadableTx, userId, code, false)
	if err != nil || !codeValid {
		return codeValid, err
	}

	err = markAsValid(loadableTx, userId)
	return codeValid, err
}

func ValidateToken(loadableTx framework.LoadableTx, userId string, code string, checkValidated bool) (bool, error) {
	entity, err := selectToken(loadableTx, userId)
	if err != nil {
		return false, err
	}

	if checkValidated && !entity.Validated {
		return false, errors.New(ErrNotValidated)
	}

	valid := totp.Validate(code, entity.Secret)
	return valid, nil
}

type MfaTokenEntity struct {
	UserId       string       `db:"user_id"`
	Secret       string       `db:"secret"`
	Url          string       `db:"url"`
	AccountName  string       `db:"account_name"`
	Image        []byte       `db:"image"`
	RecoveryCode string       `db:"recovery_code"`
	Validated    bool         `db:"validated"`
	CreatedAt    time.Time    `db:"created_at"`
	ValidatedAt  sql.NullTime `db:"validated_at"`
}

func selectToken(loadableTx framework.LoadableTx, userId string) (MfaTokenEntity, error) {
	tx, err := loadableTx()
	if err != nil {
		return MfaTokenEntity{}, err
	}

	var entities []MfaTokenEntity
	err = tx.Select(&entities, "SELECT * FROM user_totp WHERE user_id = $1", userId)
	if err != nil {
		return MfaTokenEntity{}, err
	}
	if len(entities) == 0 {
		return MfaTokenEntity{}, sql.ErrNoRows
	}
	if len(entities) > 1 {
		return MfaTokenEntity{}, errors.New(ErrTooManyTokens)
	}
	return entities[0], nil
}

func insertNewToken(txLoader framework.LoadableTx, entity MfaTokenEntity) error {
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

func markAsValid(txLoader framework.LoadableTx, userId string) error {
	tx, err := txLoader()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
UPDATE user_totp SET validated = true, validated_at = $1 WHERE user_id = $2
`, time.Now(), userId)

	return err
}
