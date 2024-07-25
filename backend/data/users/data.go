package users

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/caches"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/jmoiron/sqlx"
	"strings"
)

var cache = caches.GetCache[UserCacheItem]()

func InitCache() {
	logs.Info("Initializing cache for users")

	initial, err := selectAllUsers()
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize user cache: %v", err))
	}

	err = cache.Initialize(initial)
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize user cache: %v", err))
	}
	logs.Info("Initialized cache for users")
}

func LoadUsers() ([]UserCacheItem, error) {
	if cache.IsInitialized() {
		return cache.GetAllAsArray()
	}

	logs.Info("user cache not initialized, selecting all users")
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

	logs.Info("user cache not initialized, selecting user by mail")
	db := framework.DB()
	var result []User
	err := db.Select(&result, "select * from users where mail = $1", mail)
	if err != nil {
		logs.Info("failed to select user by mail: " + err.Error())
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
		Active:    user.Active,
		Pages:     pages,
	}, false
}

func LoadUserById(id string) (UserCacheItem, bool) {
	if cache.IsInitialized() {
		return cache.Get(id)
	}

	logs.Info("user cache not initialized, selecting user by id")
	db := framework.DB()
	var result []User
	err := db.Select(&result, "select * from users where id = $1", id)
	if err != nil {
		logs.Info("failed to select user by id: " + err.Error())
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
		Active:    user.Active,
		Pages:     pages,
	}, false
}

func InsertNewUser(tx *sqlx.Tx, user User, pageIds []string) (UserCacheItem, error) {
	_, err := tx.NamedExec(`
INSERT INTO users (id, mail, password, salt, admin, created_at, last_login, active) 
VALUES (:id, :mail, :password, :salt, :admin, current_time, current_time, true)
`, user)
	if err != nil {
		return UserCacheItem{}, err
	}

	err = insertUserPages(tx, user.Id, pageIds)
	if err != nil {
		return UserCacheItem{}, err
	}

	userPages, err := selectPagesByUserId(user.Id)
	if err != nil {
		return UserCacheItem{}, err
	}

	result := UserCacheItem{
		Id:        user.Id,
		Mail:      user.Mail,
		Password:  user.Password,
		Salt:      user.Salt,
		Admin:     user.Admin,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		Active:    user.Active,
		Pages:     userPages,
	}

	cache.StoreSafeBackground(result)
	return result, nil
}

func UpdateUser(tx *sqlx.Tx, user User, pageIdsToRemove []string, pageIdsToAdd []string) (UserCacheItem, error) {
	existingUser, ok := cache.Get(user.Id)
	if !ok {
		return UserCacheItem{}, errors.New("user not found")
	}

	_, err := tx.NamedExec(`
UPDATE users
SET mail = :mail, password = :password, salt = :salt, admin = :admin, last_login = :last_login, active = :active
WHERE id = :id
`, user)
	if err != nil {
		return UserCacheItem{}, err
	}

	err = deleteUserPages(tx, user.Id, pageIdsToRemove)
	if err != nil {
		return UserCacheItem{}, err
	}

	err = insertUserPages(tx, user.Id, pageIdsToAdd)
	if err != nil {
		return UserCacheItem{}, err
	}

	existingUser.Id = user.Id
	existingUser.Mail = user.Mail
	existingUser.Password = user.Password
	existingUser.Salt = user.Salt
	existingUser.Admin = user.Admin
	existingUser.LastLogin = user.LastLogin
	existingUser.Active = user.Active

	newPages, err := selectPagesByUserId(user.Id)
	if err != nil {
		return UserCacheItem{}, err
	}
	existingUser.Pages = newPages

	cache.StoreSafeBackground(existingUser)
	return existingUser, nil
}

func DeleteUser(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec("DELETE FROM users WHERE id = $1", id)

	if err == nil {
		err = cache.Remove(id)
	}

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

	var result []User
	err := db.Select(&result, "select * from users where id = $1 or mail = $2", id, mail)
	if err != nil || len(result) == 0 {
		return false
	}
	return true
}

func selectAllUsers() ([]UserCacheItem, error) {
	result := make([]UserCacheItem, 0)
	users := make([]User, 0)
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
			Active:    user.Active,
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
