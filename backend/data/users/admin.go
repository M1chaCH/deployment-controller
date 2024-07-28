package users

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/google/uuid"
)

func MakeSureAdminExists() {
	config := framework.Config()

	tx, err := framework.DB().Beginx()
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to begin transcation: %v", err))
	}

	var users = make([]userEntity, 0)
	err = tx.Select(&users, "SELECT * FROM users WHERE admin = true and blocked = false and onboard = true")
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to check if at least one admin exists: %v", err))
	}

	if len(users) > 0 {
		return
	}

	userId := uuid.NewString()
	logs.Info(fmt.Sprintf("no admin user exists, creating %s...", userId))

	hashedPassword, salt, err := framework.SecureHashWithSalt(config.Root.Password)
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to hash password: %v", err))
	}

	_, err = tx.Exec("INSERT INTO users (id, mail, password, salt, admin, blocked, onboard) VALUES ($1, $2, $3, $4, $5, $6, $7)", userId, config.Root.Mail, hashedPassword, salt, true, false, true)
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to insert user: %v", err))
	}

	err = tx.Commit()
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to commit transaction to insert default user: %v", err))
	}
	logs.Info("created default admin user")
}
