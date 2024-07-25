package clients

import (
	"database/sql"
	"time"
)

type KnownClient struct {
	Id         string         `db:"id"`
	RealUserId sql.NullString `db:"real_user_id"`
	CreatedAt  time.Time      `db:"created_at"`
}

type ClientDevice struct {
	Id        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	ClientId  string    `db:"client_id"`
	UserAgent string    `db:"user_agent"`
	IpAddress string    `db:"ip_address"`
}

type ClientCacheItem struct {
	Id         string         `json:"id"`
	RealUserId string         `json:"real_user_id"`
	Devices    []ClientDevice `json:"devices"`
}

func (cacheItem ClientCacheItem) GetCacheKey() string {
	return cacheItem.Id
}

func (cacheItem ClientCacheItem) IsDeviceKnown(ip, agent string) bool {
	for _, device := range cacheItem.Devices {
		if device.IpAddress == ip && device.UserAgent == agent {
			return true
		}
	}

	return false
}
