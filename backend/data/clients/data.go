package clients

import (
	"database/sql"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/caches"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
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
- add agent to user Y and update client id in cookie.
*/

//
// DB Access in this file is not transactional, because changes always need to be saved. Changes should never be rolled back.
//

var cache = caches.GetCache[ClientCacheItem]()

func InitCache() {
	logs.Info(nil, "Initializing cache for clients")

	initial, err := selectAllClients()
	if err != nil {
		logs.Panic(nil, "could not initialize client cache: %v", err)
	}

	err = cache.Initialize(initial)
	if err != nil {
		logs.Panic(nil, "could not initialize client cache: %v", err)
	}
	logs.Info(nil, "Initialized cache for clients")
}

func LoadClientInfo(c *gin.Context, clientId string) (ClientCacheItem, bool, error) {
	if cache.IsInitialized() {
		cachedItem, found := cache.Get(clientId)
		if found {
			return cachedItem, found, nil
		}
	}

	logs.Info(c, "client not found in cache, selecting client info: %s", clientId)
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
		return ClientCacheItem{}, false, nil
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

type UserDevicesEntity struct {
	UserId             string         `json:"userId" db:"user_id"`
	ClientId           string         `json:"clientId" db:"client_id"`
	DeviceId           string         `json:"deviceId" db:"device_id"`
	Ip                 string         `json:"ip" db:"ip_address"`
	Agent              string         `json:"userAgent" db:"user_agent"`
	City               sql.NullString `json:"city" db:"city"`
	Subdivision        sql.NullString `json:"subdivision" db:"subdivision"`
	Country            sql.NullString `json:"country" db:"country"`
	SystemOrganisation sql.NullString `json:"systemOrganisation" db:"system_organisation"`
	CreatedAt          time.Time      `json:"createdAt" db:"created_at"`
}

func SelectDevicesByUsers(userIds []string) ([]UserDevicesEntity, error) {
	if len(userIds) == 0 {
		return []UserDevicesEntity{}, nil
	}

	statement, args, err := sqlx.In(`
SELECT c.real_user_id as user_id, d.client_id, d.id as device_id, d.user_agent, d.ip_address, d.created_at, 
       il.city_name as city, il.subdivision_code as subdivision, il.country_code as country, il.system_organisation as system_organisation
FROM client_devices as d
    LEFT JOIN public.clients c on c.id = d.client_id
    LEFT JOIN public.ip_locations il on d.id = il.device_id
WHERE c.real_user_id in (?)
ORDER BY d.created_at DESC
`, userIds)

	if err != nil {
		return nil, err
	}

	db := framework.DB()
	statement = db.Rebind(statement)
	var result []UserDevicesEntity
	err = db.Select(&result, statement, args...)
	return result, err
}

type DeviceWithNoLocation struct {
	DeviceId  string `json:"deviceId" db:"device_id"`
	IpAddress string `json:"ipAddress" db:"ip_address"`
}

func SelectDevicesWithNoLocation() ([]DeviceWithNoLocation, error) {
	db := framework.DB()

	var devices []DeviceWithNoLocation
	err := db.Select(&devices, `
SELECT d.id as device_id, d.ip_address
FROM client_devices as d
         LEFT JOIN public.ip_locations il on d.id = il.device_id
WHERE CASE WHEN il.device_id IS NULL THEN TRUE ELSE FALSE END
  AND (d.ip_location_check_error IS NULL OR d.ip_location_check_error = '')
`)

	return devices, err
}

func UpdateDeviceAfterLocationCheck(deviceId string, otherError string) error {
	db := framework.DB()
	_, err := db.Exec(`
UPDATE client_devices SET ip_location_check_error = $1 WHERE id = $2
`, otherError, deviceId)
	return err
}

func TryFindClientOfUser(c *gin.Context, userId string) (ClientCacheItem, bool, error) {
	if cache.IsInitialized() {
		allCached, err := cache.GetAllAsArray()
		if err != nil {
			return ClientCacheItem{}, false, err
		}
		for _, item := range allCached {
			if item.RealUserId == userId {
				return item, true, nil
			}
		}
	}

	logs.Info(c, "failed to find client of user in cache checkin db, user: %s", userId)
	db := framework.DB()
	var clients []knownClientEntity
	err := db.Select(&clients, "SELECT * FROM clients WHERE real_user_id = $1", userId)
	if err != nil {
		return ClientCacheItem{}, false, err
	}

	if len(clients) == 0 {
		return ClientCacheItem{}, false, nil
	}

	return LoadClientInfo(c, clients[0].Id)
}

func TryFindExistingClient(c *gin.Context, ip, userAgent string) (ClientCacheItem, bool, error) {
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

	logs.Info(c, "client cache not initialized, selecting existing client")
	db := framework.DB()
	var devices []ClientDevice
	err := db.Select(&devices, "SELECT * FROM client_devices WHERE ip_address=$1 AND user_agent=$2 LIMIT 1", ip, userAgent)
	if err != nil {
		return ClientCacheItem{}, false, err
	}

	if len(devices) != 1 {
		return ClientCacheItem{}, false, nil
	}

	return LoadClientInfo(c, devices[0].ClientId)
}

func GetCurrentDevice(client ClientCacheItem, ip string, agent string) (ClientDevice, bool) {
	for _, device := range client.Devices {
		if device.IpAddress == ip && device.UserAgent == agent {
			return device, true
		}
	}

	return ClientDevice{}, false
}

func LookupDeviceId(c *gin.Context, clientId string, ip string, agent string) (string, error) {
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

	logs.Info(c, "device found in client cache, searching in DB")
	db := framework.DB()
	var ids []string
	err := db.Select(&ids, "SELECT client_devices.id FROM client_devices WHERE ip_address=$1 AND user_agent=$2 AND client_id = $3 LIMIT 1", ip, agent, clientId)
	if err != nil {
		return "", err
	}

	if len(ids) > 1 {
		logs.Warn(c, "found more than one device for client with same data in DB: clientId:%s ip:%s agent:%s", clientId, ip, agent)
	}
	if len(ids) < 1 {
		return "", sql.ErrNoRows
	}

	return ids[0], nil
}

func CreateNewClient(c *gin.Context, clientId string, realUserId string, ip string, userAgent string) (ClientCacheItem, error) {
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
			Validated: false,
		}},
	}
	cache.StoreSafeBackground(cacheItem)

	logs.Info(c, "created new client: agent:%s ip:%s", userAgent, ip)
	return cacheItem, nil
}

func AddDeviceToClient(c *gin.Context, clientId string, ip string, userAgent string) (ClientCacheItem, error) {
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

	logs.Info(c, "added new device to client: client:%s agent:%s ip:%s", clientId, userAgent, ip)
	client, found, err := LoadClientInfo(c, clientId)
	if !found && err == nil {
		logs.Warn(c, "just inserted client, but was not found: id:%s", clientId)
	}
	return client, err
}

func AddUserToClient(c *gin.Context, clientId string, userId string) (ClientCacheItem, error) {
	db := framework.DB()
	_, err := db.Exec("UPDATE clients SET real_user_id=$1 WHERE id=$2", userId, clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	err = cache.Remove(clientId)
	if err != nil {
		return ClientCacheItem{}, err
	}

	logs.Info(c, "added user to clinet: user:%s client:%s", userId, clientId)
	client, found, err := LoadClientInfo(c, clientId)
	if !found && err == nil {
		logs.Warn(c, "just added user to client, but client was not found: id:%s", clientId)
	}
	return client, err
}

func MarkDeviceAsValidated(c *gin.Context, clientId string, deviceId string) error {
	tx, err := framework.GetTx(c)()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE client_devices SET validated = true WHERE id = $1", deviceId)
	if err != nil {
		return err
	}

	// update cache, needs to be like this, because LoadClientInfo ignores the current transaction
	client, found, err := LoadClientInfo(c, clientId)
	if !found && err == nil {
		logs.Warn(c, "updated device but client was not found, client: "+clientId)
	}
	if err != nil {
		return err
	}
	for i, device := range client.Devices {
		if device.Id == deviceId {
			client.Devices[i].Validated = true
			break
		}
	}
	cache.StoreSafeBackground(client)

	return err
}

func MergeDevicesAndDelete(c *gin.Context, target ClientCacheItem, toMerge ClientCacheItem) (ClientCacheItem, error) {
	// 1. find devices that are new to target
	// 2. insert found device to target
	// 3. if toMerge has no user, remove client and devices else keep client

	devicesAlreadyInTarget := target.Devices
	var devicesToAdd []ClientDevice
	for _, device := range toMerge.Devices {
		foundIndex := -1
		for j, existingDevice := range devicesAlreadyInTarget {
			if device.IpAddress == existingDevice.IpAddress && device.UserAgent == existingDevice.UserAgent {
				foundIndex = j
				break
			}
		}

		if foundIndex == -1 {
			device.Id = uuid.NewString()
			device.ClientId = target.Id
			devicesToAdd = append(devicesToAdd, device)
		}
	}

	tx, err := framework.GetTx(c)()
	if err != nil {
		return ClientCacheItem{}, err
	}

	if len(devicesToAdd) > 0 {
		_, err = tx.NamedExec(`
INSERT INTO client_devices (id, client_id, ip_address, user_agent, ip_location_check_error, created_at, validated) 
VALUES (:id, :client_id, :ip_address, :user_agent, :ip_location_check_error, :created_at, :validated)
`, devicesToAdd)

		if err != nil {
			return ClientCacheItem{}, err
		}
	}

	if toMerge.RealUserId == "" {
		_, err = tx.Exec("DELETE FROM clients WHERE id = $1", toMerge.Id)
		if err != nil {
			return ClientCacheItem{}, err
		}

		err = cache.Remove(toMerge.Id)
		if err != nil {
			return ClientCacheItem{}, err
		}
	}

	logs.Info(c, "merged devices of client %s into client %s (%d devices) -- updating caches", toMerge.Id, target.Id, len(target.Devices))

	err = cache.Remove(target.Id)
	if err != nil {
		return ClientCacheItem{}, err
	}

	target.Devices = append(target.Devices, devicesToAdd...)
	err = cache.Store(target)
	if err != nil {
		return ClientCacheItem{}, err
	}

	return target, err
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
