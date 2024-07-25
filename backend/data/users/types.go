package users

import (
	"time"
)

type userEntity struct {
	Id        string    `db:"id"`
	Mail      string    `db:"mail"`
	Password  string    `db:"password"`
	Salt      []byte    `db:"salt"`
	Admin     bool      `db:"admin"`
	CreatedAt time.Time `db:"created_at"`
	LastLogin time.Time `db:"last_login"`
	Blocked   bool      `db:"blocked"`
	Onboard   bool      `db:"onboard"`
}

type UserCacheItem struct {
	Id        string              `json:"id"`
	Mail      string              `json:"mail"`
	Password  string              `json:"password"`
	Salt      []byte              `json:"salt"`
	Admin     bool                `json:"admin"`
	CreatedAt time.Time           `json:"createdAt"`
	LastLogin time.Time           `json:"lastLogin"`
	Blocked   bool                `json:"blocked"`
	Onboard   bool                `json:"onboard"`
	Pages     []UserPageCacheItem `json:"pages"`
}

type UserPageCacheItem struct {
	PageId        string `json:"pageId" db:"id"`
	TechnicalName string `json:"technicalName" db:"technical_name"`
	Url           string `json:"url" db:"url"`
	Title         string `json:"title" db:"title"`
	Description   string `json:"description" db:"description"`
	Private       bool   `json:"private" db:"private_page"`
	AccessAllowed bool   `json:"accessAllowed" db:"has_access"`
	UserId        string `json:"userId,omitempty" db:"user_id"`
}

func (u UserCacheItem) GetCacheKey() string {
	return u.Id
}
