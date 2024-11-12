package users

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/pageaccess"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/jmoiron/sqlx"
	"time"
)

func LoadUsers(txFunc framework.LoadableTx) ([]UserEntity, error) {
	tx, err := txFunc()
	if err != nil {
		return nil, err
	}

	users := make([]UserEntity, 0)

	err = tx.Select(&users, "SELECT * FROM users")
	return users, err
}

func LoadUserByMail(txFunc framework.LoadableTx, mail string) (UserEntity, bool) {
	tx, err := txFunc()
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to check db: %v", err))
		return UserEntity{}, false
	}

	var result []UserEntity
	err = tx.Select(&result, "select * from users where mail = $1", mail)
	if err != nil {
		logs.Info("failed to select UserEntity by mail: " + err.Error())
		return UserEntity{}, false
	}
	if len(result) == 0 {
		return UserEntity{}, false
	}

	// E-Mail is unique in DB, so this is always either empty or one
	user := result[0]
	return user, true
}

func LoadUserById(txFunc framework.LoadableTx, id string) (UserEntity, bool) {
	tx, err := txFunc()
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to check db: %v", err))
		return UserEntity{}, false
	}

	var result []UserEntity
	err = tx.Select(&result, "select * from users where id = $1", id)
	if err != nil {
		logs.Info("failed to select UserEntity by id: " + err.Error())
		return UserEntity{}, false
	}
	if len(result) == 0 {
		return UserEntity{}, false
	}

	user := result[0]
	return user, true
}

func InsertNewUser(txFunc framework.LoadableTx, id string, mail string, password string, salt []byte, admin bool, blocked bool, pageIds []string) error {
	tx, err := txFunc()
	if err != nil {
		return err
	}

	now := time.Now()

	_, err = tx.Exec(`
INSERT INTO users (id, mail, password, salt, admin, blocked, created_at, last_login) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, id, mail, password, salt, admin, blocked, now, now)
	if err != nil {
		return err
	}

	_, err = insertUserPages(txFunc, id, pageIds)
	if err != nil {
		return err
	}

	logs.Info(fmt.Sprintf("inserted new UserEntity: id:%s mail:%s admin:%t pages:%d", id, mail, admin, len(pageIds)))
	return nil
}

func UpdateUser(txFunc framework.LoadableTx, id string, mail string, password string, salt []byte, admin bool, blocked bool, onboard bool, lastLogin time.Time, pageIdsToRemove []string, pageIdsToAdd []string) error {
	tx, err := txFunc()
	if err != nil {
		return err
	}

	res, err := tx.Exec(`
UPDATE users
SET mail = $1, password = $2, salt = $3, admin = $4, last_login = $5, blocked = $6, onboard = $7
WHERE id = $8
`, mail, password, salt, admin, lastLogin, blocked, onboard, id)
	if err != nil {
		return err
	}

	affectedRows, err := res.RowsAffected()

	pagesDeleted, err := deleteUserPages(txFunc, id, pageIdsToRemove)
	if err != nil {
		return err
	}

	affectedRows += pagesDeleted

	pagesInserted, err := insertUserPages(txFunc, id, pageIdsToAdd)
	if err != nil {
		return err
	}

	affectedRows += pagesInserted

	if affectedRows < 1 {
		return errors.New("user not found")
	}

	logs.Info(fmt.Sprintf("updated user: id:%s mail:%s admin:%t newPages:%d", id, mail, admin, len(pageIdsToAdd)-len(pageIdsToRemove)))
	return nil
}

func DeleteUser(txFunc framework.LoadableTx, id string) error {
	tx, err := txFunc()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM users WHERE id = $1", id)

	if err == nil {
		logs.Info(fmt.Sprintf("deleted user: id:%s", id))
	}

	return err
}

func UserExists(txFunc framework.LoadableTx, id string) bool {
	tx, err := txFunc()
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to get transaction for UserExists: %v", err))
		return false
	}

	var userId string
	err = tx.Select(&userId, "select id from users where id = $1", id)
	if err != nil || userId == "" {
		return false
	}

	return true
}

func SimilarUserExists(txFunc framework.LoadableTx, id string, mail string) bool {
	tx, err := txFunc()
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to get transaction for SimilarUserExists: %v", err))
		return false
	}

	var result []UserEntity
	err = tx.Select(&result, "select * from users where id = $1 or mail = $2", id, mail)
	if err != nil || len(result) == 0 {
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to select user by mail or id: %s, %s -> %v", id, mail, err))
		}
		return false
	}
	return true
}

func MailExists(txFunc framework.LoadableTx, mail string, excludedUserId string) bool {
	tx, err := txFunc()
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to get transaction for MailExists: %v", err))
		return false
	}

	var result []UserEntity
	err = tx.Select(&result, "select * from users where mail = $1 and id != $2", mail, excludedUserId)
	if err != nil || len(result) == 0 {
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to select users by mail: %s -> %v", mail, err))
		}
		return false
	}
	return true
}

func DifferentAdminExists(txFunc framework.LoadableTx, excludedUserId string) bool {
	tx, err := txFunc()
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to get transaction for DifferentAdminExists: %v", err))
		return false
	}

	var result []UserEntity
	err = tx.Select(&result, "select * from users where admin = true and id != $1", excludedUserId)
	if err != nil || len(result) == 0 {
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to select users for admin check: %v", err))
		}
		return false
	}
	return true
}

func insertUserPages(txFunc framework.LoadableTx, userId string, pageIds []string) (int64, error) {
	if len(pageIds) < 1 {
		return 0, nil
	}

	tx, err := txFunc()
	if err != nil {
		return 0, err
	}

	userPageAccess := make([]userPageAccessEntity, len(pageIds))
	for i, id := range pageIds {
		userPageAccess[i] = userPageAccessEntity{
			UserId: userId,
			PageId: id,
		}
	}

	res, err := tx.NamedExec("INSERT INTO user_page (user_id, page_id) VALUES (:user_id, :page_id)", userPageAccess)
	if err != nil {
		return 0, err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affectedRows > 0 {
		pageaccess.DeleteUserPageAccessCache(userId)
	}

	return affectedRows, err
}

func deleteUserPages(txFunc framework.LoadableTx, userId string, pageIds []string) (int64, error) {
	if len(pageIds) < 1 {
		return 0, nil
	}

	tx, err := txFunc()
	if err != nil {
		return 0, err
	}

	statement, args, err := sqlx.In("delete from user_page where user_id = ? and page_id in (?)", userId, pageIds)
	if err != nil {
		return 0, err
	}
	statement = tx.Rebind(statement)

	res, err := tx.Exec(statement, args...)
	if err != nil {
		return 0, err
	}

	affectedRows, err := res.RowsAffected()
	return affectedRows, err
}
