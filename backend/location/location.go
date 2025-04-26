package location

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/caches"
	"github.com/M1chaCH/deployment-controller/framework/config"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type CacheItem struct {
	CityId             int
	CityName           string
	CityPlz            string
	SubdivisionId      int
	SubdivisionCode    string
	CountryId          int
	CountryCode        string
	ContinentId        int
	ContinentCode      string
	AccuracyRadius     int
	Latitude           float32
	Longitude          float32
	TimeZone           string
	SystemNumber       int
	SystemOrganisation string
	Network            string
	IpAddress          string
}

func (item CacheItem) GetCacheKey() string {
	return item.IpAddress
}

var cache = caches.GetCache[CacheItem]()

func InitCache() {
	err := cache.Initialize(make([]CacheItem, 0))
	if err != nil {
		logs.Panic(nil, "Error initializing cache for location cache: %v", err)
	}

	cnf := config.Config().Location.CacheExpireHours
	cache.SetLifetime(time.Duration(cnf) + time.Hour)
	logs.Info(nil, "successfully initialized locations cache")
}

var privateIpRegexp = regexp.MustCompile(`(^192\.168\..*$)|(^172\.16\..*$)|(^10\..*$)`)

func LoadLocation(ip string, onlyCache bool) (CacheItem, error) {
	if len(ip) < 7 {
		return CacheItem{}, errors.New("ip invalid, too short")
	}

	// 172.18.0.1 & 172.31.0.1 is a docker ip found during development
	if privateIpRegexp.MatchString(ip) || strings.ToLower(ip) == "localhost" || ip == "127.0.0.1" || ip == "172.18.0.1" || ip == "172.31.0.1" {
		ip = config.Config().Location.LocalIp
	}

	if cache.IsInitialized() {
		item, ok := cache.Get(ip)
		if ok {
			return item, nil
		}
	}

	if onlyCache {
		return CacheItem{}, errors.New("ip not found in cache")
	}

	logs.Info(nil, "ip not found in cache, checking geoip: "+ip)
	cnf := config.Config()
	auth := cnf.Location.Account + ":" + cnf.Location.License
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://geolite.info/geoip/v2.1/city/%s", ip), nil)
	if err != nil {
		return CacheItem{}, err
	}
	request.Header.Add("Authorization", "Basic "+encodedAuth)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return CacheItem{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Warn(nil, "Error closing location request response body: %v", err)
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return CacheItem{}, errors.New("invalid geoip response status: " + response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return CacheItem{}, err
	}

	var data geoIpCityDto
	err = json.Unmarshal(body, &data)
	if err != nil {
		return CacheItem{}, err
	}

	var cacheItem CacheItem
	cacheItem.CityId = data.City.GeonameId
	cacheItem.CityName = data.City.Names.En
	cacheItem.CityPlz = data.Postal.Code
	if len(data.Subdivisions) > 0 {
		if len(data.Subdivisions) > 1 {
			logs.Warn(nil, "got %d subdevisions for ip %s", len(data.Subdivisions), ip)
		}

		cacheItem.SubdivisionId = data.Subdivisions[0].GeonameId
		cacheItem.SubdivisionCode = data.Subdivisions[0].IsoCode
	}
	cacheItem.CountryId = data.Country.GeonameId
	cacheItem.CountryCode = data.Country.IsoCode
	cacheItem.ContinentId = data.Continent.GeonameId
	cacheItem.ContinentCode = data.Continent.Code
	cacheItem.AccuracyRadius = data.Location.AccuracyRadius
	cacheItem.Latitude = data.Location.Latitude
	cacheItem.Longitude = data.Location.Longitude
	cacheItem.TimeZone = data.Location.TimeZone
	cacheItem.SystemNumber = data.Traits.AutonomousSystemNumber
	cacheItem.SystemOrganisation = data.Traits.AutonomousSystemOrganization
	cacheItem.Network = data.Traits.Network
	cacheItem.IpAddress = ip

	if data.Traits.IpAddress != ip {
		logs.Warn(nil, "requested location for ip '%s' but got for ip '%s'", ip, data.Traits.IpAddress)
	}

	cache.StoreSafeBackground(cacheItem)
	return cacheItem, nil
}

type geoIpCityDto struct {
	City struct {
		GeonameId int `json:"geoname_id"`
		Names     struct {
			En string `json:"en"`
		} `json:"names"`
	} `json:"city"`
	Continent struct {
		Code      string `json:"code"`
		GeonameId int    `json:"geoname_id"`
		Names     struct {
			En string `json:"en"`
		} `json:"names"`
	} `json:"continent"`
	Country struct {
		IsoCode   string `json:"iso_code"`
		GeonameId int    `json:"geoname_id"`
		Names     struct {
			En string `json:"en"`
		} `json:"names"`
	} `json:"country"`
	Location struct {
		AccuracyRadius int     `json:"accuracy_radius"`
		Latitude       float32 `json:"latitude"`
		Longitude      float32 `json:"longitude"`
		TimeZone       string  `json:"time_zone"`
	} `json:"location"`
	Postal struct {
		Code string `json:"code"`
	} `json:"postal"`
	RegisteredCountry struct {
		IsoCode   string `json:"iso_code"`
		GeonameId int    `json:"geoname_id"`
		Names     struct {
			En string `json:"en"`
		} `json:"names"`
	} `json:"registered_country"`
	Subdivisions []struct {
		IsoCode   string `json:"iso_code"`
		GeonameId int    `json:"geoname_id"`
		Names     struct {
			En string `json:"en"`
		} `json:"names"`
	} `json:"subdivisions"`
	Traits struct {
		AutonomousSystemNumber       int    `json:"autonomous_system_number"`
		AutonomousSystemOrganization string `json:"autonomous_system_organization"`
		IpAddress                    string `json:"ip_address"`
		Network                      string `json:"network"`
	} `json:"traits"`
}
