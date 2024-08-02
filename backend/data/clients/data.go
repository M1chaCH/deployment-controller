package clients

import (
	"database/sql"
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

//
// DB Access in this file is not transactional, because changes always need to be saved. Changes should never be rolled back.
//

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

func LoadClientInfo(clientId string) (ClientCacheItem, bool, error) {
	if cache.IsInitialized() {
		cachedItem, found := cache.Get(clientId)
		if found {
			return cachedItem, found, nil
		}
	}

	logs.Info(fmt.Sprintf("client not found in cache, selecting client info: %s", clientId))
	db := framework.DB()
	var clientDataItem []knownClientEntity
	var devices []ClientDevice
	knownClientError := make(chan error)
	devicesError := make(chan error)

	go func() {
		err := db.Select(&clientDataItem, "SELECT * FROM clients WHERE id=$1", clientId)
		knownClientError <- err
	}()
	go func() {
		err := db.Select(&devices, "SELECT * FROM client_devices WHERE client_id=$1", clientId)
		devicesError <- err
	}()

	err := <-knownClientError
	if err != nil {
		return ClientCacheItem{}, false, err
	}
	if len(clientDataItem) != 1 {
		return ClientCacheItem{}, false, fmt.Errorf("no client found by id: %s", clientId)
	}
	err = <-devicesError
	if err != nil {
		return ClientCacheItem{}, false, err
	}

	client := clientDataItem[0]
	cacheItem := ClientCacheItem{
		Id:         client.Id,
		RealUserId: client.RealUserId.String,
		Devices:    devices,
	}

	cache.StoreSafeBackground(cacheItem)
	return cacheItem, true, nil
}

type UserDevicesDto struct {
	UserId    string    `json:"userId" db:"user_id"`
	ClientId  string    `json:"clientId" db:"client_id"`
	DeviceId  string    `json:"deviceId" db:"device_id"`
	Ip        string    `json:"ip" db:"ip_address"`
	Agent     string    `json:"userAgent" db:"user_agent"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

func SelectDevicesByUsers(userIds []string) ([]UserDevicesDto, error) {
	if len(userIds) == 0 {
		return []UserDevicesDto{}, nil
	}

	statement, args, err := sqlx.In(`
SELECT c.real_user_id as user_id, d.client_id, d.id as device_id, d.user_agent, d.ip_address, d.created_at
FROM client_devices as d
    LEFT JOIN public.clients c on c.id = d.client_id
WHERE c.real_user_id in (?)
ORDER BY d.created_at DESC
`, userIds)

	if err != nil {
		return nil, err
	}

	db := framework.DB()
	statement = db.Rebind(statement)
	var result []UserDevicesDto
	err = db.Select(&result, statement, args...)
	return result, err
}

func TryFindExistingClient(ip, userAgent string) (ClientCacheItem, bool, error) {
	if cache.IsInitialized() {
		resultChannel := make(chan ClientCacheItem)
		go cache.GetAll(resultChannel)
		foundClients := make([]ClientCacheItem, 0)
		for item := range resultChannel {
			for _, device := range item.Devices {
				if device.IpAddress == ip && device.UserAgent == userAgent {
					foundClients = append(foundClients, item)
				}
			}
		}

		if len(foundClients) == 1 {
			return foundClients[0], true, nil
		}

		return ClientCacheItem{}, false, nil
	}

	logs.Info("client cache not initialized, selecting existing client")
	db := framework.DB()
	var devices []ClientDevice
	err := db.Select(&devices, "SELECT * FROM client_devices WHERE ip_address=$1 AND user_agent=$2 LIMIT 1", ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, false, err
	}

	if len(devices) != 1 {
		return ClientCacheItem{}, false, nil
	}

	return LoadClientInfo(devices[0].ClientId)
}

func LookupDeviceId(clientId string, ip string, agent string) (string, error) {
	if cache.IsInitialized() {
		client, found := cache.Get(clientId)
		if found {
			for _, device := range client.Devices {
				if device.IpAddress == ip && device.UserAgent == agent {
					return device.Id, nil
				}
			}
		}
	}

	logs.Info("device found in client cache, searching in DB")
	db := framework.DB()
	var ids []string
	err := db.Select(&ids, "SELECT client_devices.id FROM client_devices WHERE ip_address=$1 AND user_agent=$2 AND client_id = $3 LIMIT 1", ip, agent, clientId)
	if err != nil {
		return "", err
	}

	if len(ids) > 1 {
		logs.Warn(fmt.Sprintf("found more than one device for client with same data in DB: clientId:%s ip:%s agent:%s", clientId, ip, agent))
	}
	if len(ids) < 1 {
		return "", sql.ErrNoRows
	}

	return ids[0], nil
}

func CreateNewClient(clientId string, realUserId string, ip string, userAgent string) (ClientCacheItem, error) {
	db := framework.DB()

	createdAt := time.Now()
	if realUserId == "" {
		_, err := db.Exec("INSERT INTO clients (id, created_at) VALUES ($1, $2)", clientId, createdAt)
		if err != nil {
			return ClientCacheItem{}, err
		}
	} else {
		_, err := db.Exec("INSERT INTO clients (id, real_user_id, created_at) VALUES ($1,$2,$3)", clientId, realUserId, createdAt)
		if err != nil {
			return ClientCacheItem{}, err
		}
	}

	deviceId := uuid.NewString()
	_, err := db.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent, created_at) VALUES ($1,$2,$3,$4,$5)", deviceId, clientId, ip, userAgent, createdAt)
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

	logs.Info(fmt.Sprintf("created new client: agent:%s ip:%s", userAgent, ip))
	return cacheItem, nil
}

func AddDeviceToClient(clientId string, ip string, userAgent string) (ClientCacheItem, error) {
	db := framework.DB()
	deviceId := uuid.NewString()
	_, err := db.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent) VALUES ($1, $2, $3, $4)", deviceId, clientId, ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, err
	}

	err = cache.Remove(clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	logs.Info(fmt.Sprintf("added new device to client: client:%s agent:%s ip:%s", clientId, userAgent, ip))
	client, found, err := LoadClientInfo(clientId)
	if !found && err == nil {
		logs.Warn(fmt.Sprintf("just inserted client, but was not found: id:%s", clientId))
	}
	return client, err
}

func AddUserToClient(clientId string, userId string) (ClientCacheItem, error) {
	db := framework.DB()
	_, err := db.Exec("UPDATE clients SET real_user_id=$1 WHERE id=$2", userId, clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	err = cache.Remove(clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	logs.Info(fmt.Sprintf("added user to clinet: user:%s client:%s", userId, clientId))
	client, found, err := LoadClientInfo(clientId)
	if !found && err == nil {
		logs.Warn(fmt.Sprintf("just added user to client, but client was not found: id:%s", clientId))
	}
	return client, err
}

func selectAllClients() ([]ClientCacheItem, error) {
	db := framework.DB()
	var result []ClientCacheItem
	var clients []knownClientEntity
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
