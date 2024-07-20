package users

import (
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

type User struct {
	Id        string    `db:"id"`
	Mail      string    `db:"mail"`
	Password  string    `db:"password"`
	Salt      []byte    `db:"salt"`
	Admin     bool      `db:"admin"`
	CreatedAt time.Time `db:"created_at"`
	LastLogin time.Time `db:"last_login"`
	Active    bool      `db:"active"`
}

type UserCacheItem struct {
	Id        string     `json:"id"`
	Mail      string     `json:"mail"`
	Password  string     `json:"password"`
	Salt      []byte     `json:"salt"`
	Admin     bool       `json:"admin"`
	CreatedAt time.Time  `json:"createdAt"`
	LastLogin time.Time  `json:"lastLogin"`
	Active    bool       `json:"active"`
	Pages     []PageItem `json:"pages"`
}

type PageItem struct {
	PageId        string `json:"pageId"`
	Url           string `json:"url"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	AccessAllowed bool   `json:"accessAllowed"`
}

func (u UserCacheItem) GetCacheKey() string {
	return u.Id
}

func (u UserCacheItem) GetPrivatePagesString() string {
	var pageIds []string
	for _, page := range u.Pages {
		pageIds = append(pageIds, page.PageId)
	}

	return strings.Join(pageIds, auth.PrivatePagesDelimiter)
}

var cache framework.ItemsCache[UserCacheItem] = &framework.LocalItemsCache[UserCacheItem]{}

// TODO update all of this for cache, eager cache, dont cache pages.db
func LaodUsers(tx *sqlx.Tx) ([]User, error) {
	if !cache.IsInitialized() {
		users := make([]User, 0)
		err := tx.Select(&users, "SELECT * FROM users")
		if err != nil {
			return nil, err
		}

		cache.Initialize(users)
		return users, nil
	}

	return cache.GetAll(), nil
}

func InsertUser(tx *sqlx.Tx, user User) error {
	_, err := tx.NamedExec(`
INSERT INTO users (id, mail, password, salt, admin, created_at, last_login, active) 
VALUES (:id, :mail, :password, :salt, :admin, current_time, current_time, true)
`, user)

	if err == nil {
		go cache.Store(user)
	}

	return err
}

func UpdateUser(tx *sqlx.Tx, user User) error {
	_, err := tx.NamedExec(`
UPDATE users
SET mail = :mail, password = :password, salt = :salt, admin = :admin, last_login = :last_login, active = :active
WHERE id = :id
`, user)

	if err == nil {
		go cache.Store(user)
	}

	return err
}

func DeleteUser(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec("DELETE FROM users WHERE id = $1", id)

	if err == nil {
		cache.Remove(id)
	}

	return err
}

func UserExists(id string) bool {
	_, ok := cache.Get(id)
	return ok
}

func SimilarUserExists(id string, mail string) bool {
	if UserExists(id) {
		return true
	}

	for _, user := range cache.GetAll() {
		if user.Mail == mail {
			return true
		}
	}

	return false
}

func SelectUserByMail(mail string) (User, error) {
	db := framework.DB()
	var user User
	err := db.Select(&user, "SELECT * FROM users WHERE mail = $1", mail)
	return user, err
}
