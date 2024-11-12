package users

import (
	"time"
)

type UserEntity struct {
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

type userPageAccessEntity struct {
	UserId string `db:"user_id"`
	PageId string `db:"page_id"`
}
