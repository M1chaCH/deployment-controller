package auth

import (
	"github.com/M1chaCH/deployment-controller/auth/mfa"
	"github.com/M1chaCH/deployment-controller/data/users"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/config"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func MakeSureAdminExists() {
	cnf := config.Config()

	tx, err := framework.DB().Beginx()
	if err != nil {
		logs.Panic(nil, "failed to begin transcation: %v", err)
	}

	var userList = make([]users.UserEntity, 0)
	err = tx.Select(&userList, "SELECT * FROM users WHERE admin = true and blocked = false")
	if err != nil {
		logs.Panic(nil, "failed to check if at least one admin exists: %v", err)
	}

	if len(userList) > 0 {
		return
	}

	userId := uuid.NewString()
	logs.Info(nil, "no admin user exists, creating %s...", userId)

	hashedPassword, salt, err := framework.SecureHashWithSalt(cnf.Root.Password)
	if err != nil {
		logs.Panic(nil, "failed to hash password: %v", err)
	}

	_, err = tx.Exec("INSERT INTO users (id, mail, password, salt, admin, blocked, onboard, mfa_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", userId, cnf.Root.Mail, hashedPassword, salt, true, false, false, mfa.TypeApp)
	if err != nil {
		logs.Panic(nil, "failed to insert user: %v", err)
	}

	err = mfa.PrepareOptionalLogging(nil, func() (*sqlx.Tx, error) { return tx, nil }, userId, mfa.TypeApp)
	if err != nil {
		logs.Panic(nil, "failed to prepare mfa for default user: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		logs.Panic(nil, "failed to commit transaction to insert default user: %v", err)
	}
	logs.Info(nil, "created default admin user")
}
