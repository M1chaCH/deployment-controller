package users

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/caches"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

var cache = caches.GetCache[UserCacheItem]()

func InitCache() {
	logs.Info("Initializing cache for users")

	initial, err := selectAllUsers()
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize userEntity cache: %v", err))
	}

	err = cache.Initialize(initial)
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize userEntity cache: %v", err))
	}
	logs.Info("Initialized cache for users")
}

func LoadUsers() ([]UserCacheItem, error) {
	if cache.IsInitialized() {
		return cache.GetAllAsArray()
	}

	logs.Info("userEntity cache not initialized, selecting all users")
	return selectAllUsers()
}

func LoadUserByMail(mail string) (UserCacheItem, bool) {
	if cache.IsInitialized() {
		receiver := make(chan UserCacheItem)
		go cache.GetAll(receiver)
		for user := range receiver {
			if user.Mail == mail {
				return user, true
			}
		}
	}

	logs.Info("userEntity cache not initialized, selecting userEntity by mail")
	db := framework.DB()
	var result []userEntity
	err := db.Select(&result, "select * from users where mail = $1", mail)
	if err != nil {
		logs.Info("failed to select userEntity by mail: " + err.Error())
		return UserCacheItem{}, false
	}
	if len(result) == 0 {
		return UserCacheItem{}, false
	}

	user := result[0]
	pages, err := selectPagesByUserId(user.Id)
	return UserCacheItem{
		Id:        user.Id,
		Mail:      user.Mail,
		Password:  user.Password,
		Salt:      user.Salt,
		Admin:     user.Admin,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		Blocked:   user.Blocked,
		Pages:     pages,
	}, false
}

func LoadUserById(id string) (UserCacheItem, bool) {
	if cache.IsInitialized() {
		return cache.Get(id)
	}

	logs.Info("userEntity cache not initialized, selecting userEntity by id")
	db := framework.DB()
	var result []userEntity
	err := db.Select(&result, "select * from users where id = $1", id)
	if err != nil {
		logs.Info("failed to select userEntity by id: " + err.Error())
		return UserCacheItem{}, false
	}
	if len(result) == 0 {
		return UserCacheItem{}, false
	}

	user := result[0]
	pages, err := selectPagesByUserId(user.Id)
	return UserCacheItem{
		Id:        user.Id,
		Mail:      user.Mail,
		Password:  user.Password,
		Salt:      user.Salt,
		Admin:     user.Admin,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		Blocked:   user.Blocked,
		Pages:     pages,
	}, false
}

func InsertNewUser(tx *sqlx.Tx, id string, mail string, password string, salt []byte, admin bool, blocked bool, pageIds []string) (UserCacheItem, error) {
	now := time.Now()

	_, err := tx.Exec(`
INSERT INTO users (id, mail, password, salt, admin, blocked, created_at, last_login) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, id, mail, password, salt, admin, blocked, now, now)
	if err != nil {
		return UserCacheItem{}, err
	}

	err = insertUserPages(tx, id, pageIds)
	if err != nil {
		return UserCacheItem{}, err
	}

	userPages, err := selectPagesByUserId(id)
	if err != nil {
		return UserCacheItem{}, err
	}

	result := UserCacheItem{
		Id:        id,
		Mail:      mail,
		Password:  password,
		Salt:      salt,
		Admin:     admin,
		CreatedAt: now,
		LastLogin: now,
		Blocked:   blocked,
		Onboard:   false,
		Pages:     userPages,
	}

	cache.StoreSafeBackground(result)
	logs.Info(fmt.Sprintf("inserted new userEntity: id:%s mail:%s admin:%t pages:%d", id, mail, admin, len(pageIds)))
	return result, nil
}

func UpdateUser(tx *sqlx.Tx, id string, mail string, password string, salt []byte, admin bool, blocked bool, onboard bool, lastLogin time.Time, pageIdsToRemove []string, pageIdsToAdd []string) (UserCacheItem, error) {
	existingUser, ok := cache.Get(id)
	if !ok {
		return UserCacheItem{}, errors.New("userEntity not found")
	}

	_, err := tx.Exec(`
UPDATE users
SET mail = $1, password = $2, salt = $3, admin = $4, last_login = $5, blocked = $6, onboard = $7
WHERE id = $8
`, mail, password, salt, admin, lastLogin, blocked, onboard, id)
	if err != nil {
		return UserCacheItem{}, err
	}

	err = deleteUserPages(tx, id, pageIdsToRemove)
	if err != nil {
		return UserCacheItem{}, err
	}

	err = insertUserPages(tx, id, pageIdsToAdd)
	if err != nil {
		return UserCacheItem{}, err
	}

	existingUser.Id = id
	existingUser.Mail = mail
	existingUser.Password = password
	existingUser.Salt = salt
	existingUser.Admin = admin
	existingUser.LastLogin = lastLogin
	existingUser.Blocked = blocked

	newPages, err := selectPagesByUserId(id)
	if err != nil {
		return UserCacheItem{}, err
	}
	existingUser.Pages = newPages

	cache.StoreSafeBackground(existingUser)
	logs.Info(fmt.Sprintf("updated user: id:%s mail:%s admin:%t newPages:%d", id, mail, admin, len(pageIdsToAdd)-len(pageIdsToRemove)))
	return existingUser, nil
}

func DeleteUser(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec("DELETE FROM users WHERE id = $1", id)

	if err == nil {
		err = cache.Remove(id)
	}

	logs.Info(fmt.Sprintf("deleted user: id:%s", id))
	return err
}

func UserExists(id string) bool {
	db := framework.DB()

	var userId string
	err := db.Select(&userId, "select id from users where id = $1", id)
	if err != nil || userId == "" {
		return false
	}

	return true
}

func SimilarUserExists(id string, mail string) bool {
	db := framework.DB()

	var result []userEntity
	err := db.Select(&result, "select * from users where id = $1 or mail = $2", id, mail)
	if err != nil || len(result) == 0 {
		return false
	}
	return true
}

func selectAllUsers() ([]UserCacheItem, error) {
	result := make([]UserCacheItem, 0)
	users := make([]userEntity, 0)
	usersError := make(chan error)
	allPageItems := make([]UserPageCacheItem, 0)
	pageItemsError := make(chan error)
	db := framework.DB()

	go func() {
		err := db.Select(&users, "SELECT * FROM users")
		usersError <- err
	}()
	go func() {
		err := db.Select(&allPageItems, `
SELECT p.id, p.technical_name, p.url, p.title, p.description, p.private_page, up.user_id, 
       CASE WHEN up.user_id IS NOT NULL THEN TRUE ELSE FALSE END AS has_access
FROM pages AS p
LEFT JOIN user_page up ON p.id = up.page_id
`)
		pageItemsError <- err
	}()

	err := <-usersError
	if err != nil {
		return result, err
	}

	err = <-pageItemsError
	if err != nil {
		return result, err
	}

	for _, user := range users {
		userPageItems := make([]UserPageCacheItem, 0)
		for _, page := range allPageItems {
			if page.UserId == user.Id {
				userPageItems = append(userPageItems, page)
			}
		}

		result = append(result, UserCacheItem{
			Id:        user.Id,
			Mail:      user.Mail,
			Password:  user.Password,
			Salt:      user.Salt,
			Admin:     user.Admin,
			CreatedAt: user.CreatedAt,
			LastLogin: user.LastLogin,
			Blocked:   user.Blocked,
			Pages:     userPageItems,
		})
	}

	return result, nil
}

func selectPagesByUserId(userId string) ([]UserPageCacheItem, error) {
	db := framework.DB()
	var userPages []UserPageCacheItem
	err := db.Select(&userPages, `
SELECT p.id, p.url, p.title, p.description, p.private_page, up.user_id, 
       CASE WHEN up.user_id IS NOT NULL THEN TRUE ELSE FALSE END AS has_access
FROM pages AS p
LEFT JOIN user_page up ON p.id = up.page_id
WHERE up.user_id = $1
`, userId)

	return userPages, err
}

func insertUserPages(tx *sqlx.Tx, userId string, pageIds []string) error {
	if len(pageIds) < 1 {
		return nil
	}

	query := "INSERT INTO user_page (user_id, page_id) VALUES "
	var values []string
	var args []interface{}
	argIndex := 1

	for _, pageId := range pageIds {
		values = append(values, fmt.Sprintf("($%d, $%d)", argIndex, argIndex+1))
		args = append(args, userId, pageId)
		argIndex += 2
	}

	query += strings.Join(values, ",")

	_, err := tx.Exec(query, args...)
	return err
}

func deleteUserPages(tx *sqlx.Tx, userId string, pageIds []string) error {
	pageIdsString := "(" + strings.Join(pageIds, ",") + ")"
	_, err := tx.Exec("DELETE FROM user_page WHERE user_id = $1 AND page_id in $2", userId, pageIdsString)
	return err
}
