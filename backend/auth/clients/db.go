package clients

import (
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

/*
Goal
- find a secure way to know that a device is a user
	- only required the user to do two factor auth no user can be assumed
- track not logged in users

Scenarios
1. User logged in
- UserId shows what user this is
- user stays logged in as long as ip & agent stay in a known range
- if user ip changes but agent not, still logged in, but IP is added to known client
- if user agent changes, user logged out, except agent is known for user

2. User not logged in
- receives a clientIp
- clientId refers to a user agent and ip
- if clientId exists and anything changes -> add to DB
- if clientId does not exist, but ip and agent is the same as some known clientId -> assign clientId
- if user logs in, add clientId to userId

3. Client X logs in with User X, logs out, logs in with User Y and Client Y
- TODO!?!
- add agent to user Y and update client id in cookie.
*/

type KnownClient struct {
	Id         string    `db:"id"`
	RealUserId string    `db:"real_user_id"`
	CreatedAt  time.Time `db:"created_at"`
}

type ClientDevice struct {
	Id        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	ClientId  string    `db:"client_id"`
	UserAgent string    `db:"user_agent"`
	IpAddress string    `db:"ip_address"`
}

/*
TODO, not sure if this noramlisation is of any advantage.

	type ClientDeviceIp struct {
		Id string `db:"id"`
		ClientDeviceId string `db:"client_device_id"`
		CreatedAt time.Time `db:"created_at"`
		IpAddress string `db:"ip_address"`
	}
*/
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

var cache framework.ItemsCache[ClientCacheItem] = &framework.LocalItemsCache[ClientCacheItem]{}

func LoadClientInfo(clientId string) (ClientCacheItem, error) {
	cachedItem, ok := cache.Get(clientId)
	if ok {
		return cachedItem, nil
	}

	db := framework.DB()
	var client KnownClient
	var devices []ClientDevice
	knownClientError := make(chan error)
	devicesError := make(chan error)

	go func() {
		err := db.Select(&client, "SELECT * FROM clients WHERE id=?", clientId)
		knownClientError <- err
	}()
	go func() {
		err := db.Select(&devices, "SELECT * FROM client_devices WHERE client_id=?", clientId)
		devicesError <- err
	}()

	err := <-knownClientError
	if err != nil {
		return ClientCacheItem{}, err
	}
	err = <-devicesError
	if err != nil {
		return ClientCacheItem{}, err
	}

	cacheItem := ClientCacheItem{
		Id:         client.Id,
		RealUserId: client.RealUserId,
		Devices:    devices,
	}
	go cache.Store(cacheItem)

	return cacheItem, nil
}

func LoadExistingClient(ip, userAgent string) (ClientCacheItem, error) {
	for _, item := range cache.GetAll() {
		for _, device := range item.Devices {
			if device.IpAddress == ip && device.UserAgent == userAgent {
				return item, nil
			}
		}
	}

	db := framework.DB()
	var devices []ClientDevice
	err := db.Select(&devices, "SELECT * FROM client_devices WHERE ip_address=? AND user_agent=?", ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, err
	}

	if len(devices) <= 0 {
		return ClientCacheItem{}, nil
	}

	return LoadClientInfo(devices[0].ClientId)
}

func CreateNewClient(tx *sqlx.Tx, realUserId string, ip string, userAgent string) (ClientCacheItem, error) {
	clientId := uuid.NewString()
	createdAt := time.Now()
	if realUserId == "" {
		_, err := tx.Exec("INSERT INTO clients (id, created_at) VALUES (?, ?)", clientId, createdAt)
		if err != nil {
			return ClientCacheItem{}, err
		}
	} else {
		_, err := tx.Exec("INSERT INTO clients (id, real_user_id, created_at) VALUES (?,?,?)", clientId, realUserId, createdAt)
		if err != nil {
			return ClientCacheItem{}, err
		}
	}

	deviceId := uuid.NewString()
	_, err := tx.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent, created_at) VALUES (?,?,?,?,?)", deviceId, clientId, ip, userAgent, createdAt)
	if err != nil {
		return ClientCacheItem{}, err
	}

	cacheItem := ClientCacheItem{
		Id:         clientId,
		RealUserId: realUserId,
		Devices: []ClientDevice{{
			Id:        deviceId,
			CreatedAt: createdAt,
			ClientId:  clientId,
			IpAddress: ip,
			UserAgent: userAgent,
		}},
	}
	go cache.Store(cacheItem)

	return cacheItem, nil
}

func AddDeviceToClient(tx *sqlx.Tx, clientId string, ip string, userAgent string) (ClientCacheItem, error) {
	deviceId := uuid.NewString()
	_, err := tx.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent) VALUES (?, ?, ?, ?)", deviceId, clientId, ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, err
	}

	cache.Remove(clientId)
	return LoadClientInfo(clientId)
}

func AddUserToClient(tx *sqlx.Tx, clientId string, userId string) (ClientCacheItem, error) {
	_, err := tx.Exec("UPDATE clients SET real_user_id=? WHERE id=?", userId, clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	cache.Remove(clientId)
	return LoadClientInfo(clientId)
}
