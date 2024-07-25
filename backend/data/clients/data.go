package clients

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/caches"
	"github.com/M1chaCH/deployment-controller/framework/logs"
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

var cache = caches.GetCache[ClientCacheItem]()

func InitCache() {
	logs.Info("Initializing cache for clients")

	initial, err := selectAllClients()
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize client cache: %v", err))
	}

	err = cache.Initialize(initial)
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize client cache: %v", err))
	}
	logs.Info("Initialized cache for clients")
}

func LoadClientInfo(clientId string) (ClientCacheItem, error) {
	if cache.IsInitialized() {
		cachedItem, _ := cache.Get(clientId)
		return cachedItem, nil
	}

	logs.Info("client cache not initialized, selecting client info")
	db := framework.DB()
	var client KnownClient
	var devices []ClientDevice
	knownClientError := make(chan error)
	devicesError := make(chan error)

	go func() {
		err := db.Select(&client, "SELECT * FROM clients WHERE id=$1", clientId)
		knownClientError <- err
	}()
	go func() {
		err := db.Select(&devices, "SELECT * FROM client_devices WHERE client_id=$1", clientId)
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
		RealUserId: client.RealUserId.String,
		Devices:    devices,
	}

	return cacheItem, nil
}

func LoadExistingClient(ip, userAgent string) (ClientCacheItem, error) {
	if cache.IsInitialized() {
		resultChannel := make(chan ClientCacheItem)
		go cache.GetAll(resultChannel)
		for item := range resultChannel {
			for _, device := range item.Devices {
				if device.IpAddress == ip && device.UserAgent == userAgent {
					return item, nil
				}
			}
		}
		return ClientCacheItem{}, nil
	}

	logs.Info("client cache not initialized, selecting existing client")
	db := framework.DB()
	var devices []ClientDevice
	err := db.Select(&devices, "SELECT * FROM client_devices WHERE ip_address=$1 AND user_agent=$2 LIMIT 1", ip, userAgent)
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
		_, err := tx.Exec("INSERT INTO clients (id, created_at) VALUES ($1, $2)", clientId, createdAt)
		if err != nil {
			return ClientCacheItem{}, err
		}
	} else {
		_, err := tx.Exec("INSERT INTO clients (id, real_user_id, created_at) VALUES ($1,$2,$3)", clientId, realUserId, createdAt)
		if err != nil {
			return ClientCacheItem{}, err
		}
	}

	deviceId := uuid.NewString()
	_, err := tx.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent, created_at) VALUES ($1,$2,$3,$4,$5)", deviceId, clientId, ip, userAgent, createdAt)
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
	cache.StoreSafeBackground(cacheItem)

	return cacheItem, nil
}

func AddDeviceToClient(tx *sqlx.Tx, clientId string, ip string, userAgent string) (ClientCacheItem, error) {
	deviceId := uuid.NewString()
	_, err := tx.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent) VALUES ($1, $2, $3, $4)", deviceId, clientId, ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, err
	}

	err = cache.Remove(clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	return LoadClientInfo(clientId)
}

func AddUserToClient(tx *sqlx.Tx, clientId string, userId string) (ClientCacheItem, error) {
	_, err := tx.Exec("UPDATE clients SET real_user_id=$1 WHERE id=$2", userId, clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	err = cache.Remove(clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}
	return LoadClientInfo(clientId)
}

func selectAllClients() ([]ClientCacheItem, error) {
	db := framework.DB()
	var result []ClientCacheItem
	var clients []KnownClient
	var devices []ClientDevice
	knownClientsError := make(chan error)
	devicesError := make(chan error)

	go func() {
		err := db.Select(&clients, "SELECT * FROM clients")
		knownClientsError <- err
	}()
	go func() {
		err := db.Select(&devices, "SELECT * FROM client_devices ORDER BY client_id")
		devicesError <- err
	}()

	err := <-knownClientsError
	if err != nil {
		return result, err
	}
	err = <-devicesError
	if err != nil {
		return result, err
	}

	for _, client := range clients {
		clientDevices := make([]ClientDevice, 0)
		foundDevices := false
		for _, device := range devices {
			if device.ClientId == client.Id {
				clientDevices = append(clientDevices, device)
				foundDevices = true
			} else if foundDevices {
				// since is ordered, we know there won't be more for current client
				break
			}
		}

		result = append(result, ClientCacheItem{
			Id:         client.Id,
			RealUserId: client.RealUserId.String,
			Devices:    clientDevices,
		})
	}

	return result, err
}
