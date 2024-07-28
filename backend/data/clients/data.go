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
- TODO!?!, current logic needs to be tested
- add agent to user Y and update client id in cookie.
*/

var cache = caches.GetCache[ClientCacheItem]()

func InitCache() {
	logs.Info("Initializing cache for clients")

	tx, err := framework.DB().Beginx()
	txFunc := func() (*sqlx.Tx, error) {
		return tx, err
	}

	initial, err := selectAllClients(txFunc)
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize client cache: %v", err))
	}

	err = cache.Initialize(initial)
	if err != nil {
		logs.Panic(fmt.Sprintf("could not initialize client cache: %v", err))
	}

	err = tx.Commit()
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to commit client cache: %v", err))
	}
	logs.Info("Initialized cache for clients")
}

func LoadClientInfo(txFunc func() (*sqlx.Tx, error), clientId string) (ClientCacheItem, bool, error) {
	if cache.IsInitialized() {
		cachedItem, found := cache.Get(clientId)
		if found {
			return cachedItem, found, nil
		}
	}

	logs.Info(fmt.Sprintf("client not found in cache, selecting client info: %s", clientId))
	tx, err := txFunc()
	if err != nil {
		return ClientCacheItem{}, false, err
	}

	var clientDataItem []knownClientEntity
	err = tx.Select(&clientDataItem, "SELECT * FROM clients WHERE id=$1", clientId)
	if err != nil {
		return ClientCacheItem{}, false, err
	}
	if len(clientDataItem) != 1 {
		return ClientCacheItem{}, false, fmt.Errorf("no client found by id: %s", clientId)
	}

	var devices []ClientDevice
	err = tx.Select(&devices, "SELECT * FROM client_devices WHERE client_id=$1", clientId)
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

func TryFindExistingClient(txFunc func() (*sqlx.Tx, error), ip, userAgent string) (ClientCacheItem, bool, error) {
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
	tx, err := txFunc()
	if err != nil {
		return ClientCacheItem{}, false, err
	}

	var devices []ClientDevice
	err = tx.Select(&devices, "SELECT * FROM client_devices WHERE ip_address=$1 AND user_agent=$2 LIMIT 1", ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, false, err
	}

	if len(devices) != 1 {
		return ClientCacheItem{}, false, nil
	}

	return LoadClientInfo(txFunc, devices[0].ClientId)
}

func CreateNewClient(txFunc func() (*sqlx.Tx, error), clientId string, realUserId string, ip string, userAgent string) (ClientCacheItem, error) {
	tx, err := txFunc()
	if err != nil {
		return ClientCacheItem{}, err
	}

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
	_, err = tx.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent, created_at) VALUES ($1,$2,$3,$4,$5)", deviceId, clientId, ip, userAgent, createdAt)
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

func AddDeviceToClient(txFunc func() (*sqlx.Tx, error), clientId string, ip string, userAgent string) (ClientCacheItem, error) {
	tx, err := txFunc()
	if err != nil {
		return ClientCacheItem{}, err
	}

	deviceId := uuid.NewString()
	_, err = tx.Exec("INSERT INTO client_devices (id, client_id, ip_address, user_agent) VALUES ($1, $2, $3, $4)", deviceId, clientId, ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, err
	}

	err = cache.Remove(clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	logs.Info(fmt.Sprintf("added new device to client: client:%s agent:%s ip:%s", clientId, userAgent, ip))
	client, found, err := LoadClientInfo(txFunc, clientId)
	if !found && err == nil {
		logs.Warn(fmt.Sprintf("just inserted client, but was not found: id:%s", clientId))
	}
	return client, err
}

func AddUserToClient(txFunc func() (*sqlx.Tx, error), clientId string, userId string) (ClientCacheItem, error) {
	tx, err := txFunc()
	if err != nil {
		return ClientCacheItem{}, err
	}

	_, err = tx.Exec("UPDATE clients SET real_user_id=$1 WHERE id=$2", userId, clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	err = cache.Remove(clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	logs.Info(fmt.Sprintf("added user to clinet: user:%s client:%s", userId, clientId))
	client, found, err := LoadClientInfo(txFunc, clientId)
	if !found && err == nil {
		logs.Warn(fmt.Sprintf("just added user to client, but client was not found: id:%s", clientId))
	}
	return client, err
}

func selectAllClients(txFunc func() (*sqlx.Tx, error)) ([]ClientCacheItem, error) {
	tx, err := txFunc()
	if err != nil {
		return nil, err
	}

	var result []ClientCacheItem
	var clients []knownClientEntity
	err = tx.Select(&clients, "SELECT * FROM clients")
	if err != nil {
		return result, err
	}

	var devices []ClientDevice
	err = tx.Select(&devices, "SELECT * FROM client_devices ORDER BY client_id")
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
