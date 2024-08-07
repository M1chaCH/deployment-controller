package location

import "github.com/M1chaCH/deployment-controller/framework"

type DbEntity struct {
	DeviceId           string  `db:"device_id"`
	CityId             int     `db:"city_id"`
	CityName           string  `db:"city_name"`
	CityPlz            string  `db:"city_plz"`
	SubdivisionId      int     `db:"subdivision_id"`
	SubdivisionCode    string  `db:"subdivision_code"`
	CountryId          int     `db:"country_id"`
	CountryCode        string  `db:"country_code"`
	ContinentId        int     `db:"continent_id"`
	ContinentCode      string  `db:"continent_code"`
	AccuracyRadius     int     `db:"accuracy_radius"`
	Latitude           float32 `db:"latitude"`
	Longitude          float32 `db:"longitude"`
	TimeZone           string  `db:"time_zone"`
	SystemNumber       int     `db:"system_number"`
	SystemOrganisation string  `db:"system_organisation"`
	Network            string  `db:"network"`
	IpAddress          string  `db:"ip_address"`
}

func InsertLocation(entity DbEntity) error {
	db := framework.DB()

	_, err := db.NamedExec(`
INSERT INTO ip_locations (device_id, city_id, city_name, city_plz, subdivision_id, subdivision_code, country_id, country_code, continent_id, continent_code, accuracy_radius, latitude, longitude, time_zone, system_number, system_organisation, network, ip_address) 
VALUES (:device_id, :city_id, :city_name, :city_plz, :subdivision_id, :subdivision_code, :country_id, :country_code, :continent_id, :continent_code, :accuracy_radius, :latitude, :longitude, :time_zone, :system_number, :system_organisation, :network, :ip_address)
`, entity)

	cache.StoreSafeBackground(CacheItem{
		CityId:             entity.CityId,
		CityName:           entity.CityName,
		CityPlz:            entity.CityPlz,
		SubdivisionId:      entity.SubdivisionId,
		SubdivisionCode:    entity.SubdivisionCode,
		CountryId:          entity.CountryId,
		CountryCode:        entity.CountryCode,
		ContinentId:        entity.ContinentId,
		ContinentCode:      entity.ContinentCode,
		AccuracyRadius:     entity.AccuracyRadius,
		Latitude:           entity.Latitude,
		Longitude:          entity.Longitude,
		TimeZone:           entity.TimeZone,
		SystemNumber:       entity.SystemNumber,
		SystemOrganisation: entity.SystemOrganisation,
		Network:            entity.Network,
		IpAddress:          entity.IpAddress,
	})

	return err
}
