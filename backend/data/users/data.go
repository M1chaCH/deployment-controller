package users

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/pages"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/caches"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/jmoiron/sqlx"
	"time"
)

var cache = caches.GetCache[UserCacheItem]()

func InitCache() {
	logs.Info("Initializing cache for users")

	tx, err := framework.DB().Beginx()
	if err != nil {
		logs.Panic(fmt.Sprintf("could not start transaction: %v", err))
	}

	err = RefreshCash(tx)
	if err != nil {
		panic(err)
	}
}

func RefreshCash(tx *sqlx.Tx) error {
	initial, err := selectAllUsers(tx)
	if err != nil {
		logs.Warn(fmt.Sprintf("failed to select all users, cache will not be up to date: %v", err))
		return err
	}

	err = cache.Initialize(initial)
	if err != nil {
		logs.Severe(fmt.Sprintf("could not initialize userEntity cache: %v", err))
		return err
	}

	logs.Info("refreshed user cache")
	return nil
}

func LoadUsers(tx *sqlx.Tx) ([]UserCacheItem, error) {
	if cache.IsInitialized() {
		users, err := cache.GetAllAsArray()
		if len(users) > 0 || err != nil {
			return users, err
		}
	}

	logs.Info("no users found in cache, selecting all users")
	users, err := selectAllUsers(tx)
	if err == nil && len(users) > 0 {
		err = cache.Initialize(users)
	}
	return users, err
}

func LoadUserByMail(tx *sqlx.Tx, mail string) (UserCacheItem, bool) {
	if cache.IsInitialized() {
		receiver := make(chan UserCacheItem)
		go cache.GetAll(receiver)
		for user := range receiver {
			if user.Mail == mail {
				return user, true
			}
		}
	}

	logs.Info(fmt.Sprintf("user by email not found in cache, checking db: %s", mail))
	var result []userEntity
	err := tx.Select(&result, "select * from users where mail = $1", mail)
	if err != nil {
		logs.Info("failed to select userEntity by mail: " + err.Error())
		return UserCacheItem{}, false
	}
	if len(result) == 0 {
		return UserCacheItem{}, false
	}

	user := result[0]
	userPages, err := selectPagesByUserId(tx, user.Id)
	return UserCacheItem{
		Id:        user.Id,
		Mail:      user.Mail,
		Password:  user.Password,
		Salt:      user.Salt,
		Admin:     user.Admin,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		Blocked:   user.Blocked,
		Onboard:   user.Onboard,
		Pages:     userPages,
	}, false
}

func LoadUserById(tx *sqlx.Tx, id string) (UserCacheItem, bool) {
	if cache.IsInitialized() {
		user, found := cache.Get(id)
		if found {
			return user, found
		}
	}

	logs.Info(fmt.Sprintf("user not found in cache, selecting userEntity by id: %s", id))
	var result []userEntity
	err := tx.Select(&result, "select * from users where id = $1", id)
	if err != nil {
		logs.Info("failed to select userEntity by id: " + err.Error())
		return UserCacheItem{}, false
	}
	if len(result) == 0 {
		return UserCacheItem{}, false
	}

	user := result[0]
	userPages, err := selectPagesByUserId(tx, user.Id)
	cacheItem := UserCacheItem{
		Id:        user.Id,
		Mail:      user.Mail,
		Password:  user.Password,
		Salt:      user.Salt,
		Admin:     user.Admin,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		Blocked:   user.Blocked,
		Onboard:   user.Onboard,
		Pages:     userPages,
	}
	cache.StoreSafeBackground(cacheItem)
	return cacheItem, true
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

	userPages, err := selectPagesByUserId(tx, id)
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
	existingUser.Onboard = onboard

	newPages, err := selectPagesByUserId(tx, id)
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

func UserExists(tx *sqlx.Tx, id string) bool {
	var userId string
	err := tx.Select(&userId, "select id from users where id = $1", id)
	if err != nil || userId == "" {
		return false
	}

	return true
}

func SimilarUserExists(tx *sqlx.Tx, id string, mail string) bool {
	var result []userEntity
	err := tx.Select(&result, "select * from users where id = $1 or mail = $2", id, mail)
	if err != nil || len(result) == 0 {
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to select user by mail or id: %s, %s -> %v", id, mail, err))
		}
		return false
	}
	return true
}

func MailExists(tx *sqlx.Tx, mail string, excludedUserId string) bool {
	var result []userEntity
	err := tx.Select(&result, "select * from users where mail = $1 and id != $2", mail, excludedUserId)
	if err != nil || len(result) == 0 {
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to select users by mail: %s -> %v", mail, err))
		}
		return false
	}
	return true
}

func DifferentAdminExists(tx *sqlx.Tx, excludedUserId string) bool {
	var result []userEntity
	err := tx.Select(&result, "select * from users where admin = true and id != $1", excludedUserId)
	if err != nil || len(result) == 0 {
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to select users for admin check: %v", err))
		}
		return false
	}
	return true
}

func selectAllUsers(tx *sqlx.Tx) ([]UserCacheItem, error) {
	result := make([]UserCacheItem, 0)

	users := make([]userEntity, 0)
	err := tx.Select(&users, "SELECT * FROM users")
	if err != nil {
		return result, err
	}

	allPages, err := pages.LoadPages(tx)
	if err != nil {
		return result, err
	}

	userPageAccess := make([]userPageAccessEntity, 0)
	err = tx.Select(&userPageAccess, "SELECT * FROM user_page")
	if err != nil {
		return result, err
	}

	for _, user := range users {
		userPageItems := make([]UserPageCacheItem, 0)
		for _, page := range allPages {
			page := UserPageCacheItem{
				PageId:        page.Id,
				TechnicalName: page.TechnicalName,
				Url:           page.Url,
				Title:         page.Title,
				Description:   page.Description,
				Private:       page.PrivatePage,
			}

			hasAccess := !page.Private
			if page.Private {
				for _, userPage := range userPageAccess {
					if userPage.UserId == user.Id && userPage.PageId == page.PageId {
						hasAccess = true
						break
					}
				}
			}

			page.AccessAllowed = hasAccess
			userPageItems = append(userPageItems, page)
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
			Onboard:   user.Onboard,
			Pages:     userPageItems,
		})
	}

	return result, nil
}

func selectPagesByUserId(tx *sqlx.Tx, userId string) ([]UserPageCacheItem, error) {
	var userPages []UserPageCacheItem
	err := tx.Select(&userPages, `
SELECT p.id, p.url, p.title, p.description, p.private_page, p.technical_name,
       CASE 
           WHEN up.user_id IS NOT NULL 
                    OR p.private_page IS NOT TRUE 
               THEN TRUE 
           ELSE FALSE 
       END AS has_access
FROM pages AS p
LEFT JOIN user_page up ON p.id = up.page_id AND up.user_id = $1
`, userId)

	return userPages, err
}

func insertUserPages(tx *sqlx.Tx, userId string, pageIds []string) error {
	if len(pageIds) < 1 {
		return nil
	}

	userPageAccess := make([]userPageAccessEntity, len(pageIds))
	for i, id := range pageIds {
		userPageAccess[i] = userPageAccessEntity{
			UserId: userId,
			PageId: id,
		}
	}

	_, err := tx.NamedExec("INSERT INTO user_page (user_id, page_id) VALUES (:user_id, :page_id)", userPageAccess)
	return err
}

func deleteUserPages(tx *sqlx.Tx, userId string, pageIds []string) error {
	if len(pageIds) < 1 {
		return nil
	}

	statement, args, err := sqlx.In("delete from user_page where user_id = ? and page_id in (?)", userId, pageIds)
	if err != nil {
		return err
	}
	statement = tx.Rebind(statement)

	_, err = tx.Exec(statement, args...)
	return err
}
