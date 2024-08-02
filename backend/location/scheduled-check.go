package location

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
)

func InitScheduledLocationCheck() {
	logs.Info("running device location check on startup")
	if err := checkDevicesWithNoLocation(); err != nil { // run once initially to make sure it works at least on startup...
		logs.Panic(fmt.Sprintf("check 'device_with_no_location' failed initially: %v", err))
	}
	framework.RunScheduledTask("device_with_no_location", framework.Config().Location.CheckWaitTimeMinutes, checkDevicesWithNoLocation)
}

func checkDevicesWithNoLocation() error {
	devicesToCheck, err := clients.SelectDevicesWithNoLocation()
	if err != nil {
		return err
	}

	if len(devicesToCheck) == 0 {
		logs.Info("all devices have a location, nothing done")
		return nil
	}

	logs.Info(fmt.Sprintf("loading locations of %d devices", len(devicesToCheck)))

	for _, device := range devicesToCheck {
		location, err := LoadLocation(device.IpAddress)

		if errors.Is(err, PrivateIpError) {
			err = clients.UpdateDeviceAfterLocationCheck(device.DeviceId, true, "")
			if err != nil {
				return err
			}
		} else if err != nil {
			logs.Warn(fmt.Sprintf("error loading location for ip (%s): %v", device.IpAddress, err))
			err = clients.UpdateDeviceAfterLocationCheck(device.DeviceId, false, err.Error())
			if err != nil {
				return err
			}
		} else {
			err = InsertLocation(DbEntity{
				DeviceId:           device.DeviceId,
				CityId:             location.CityId,
				CityName:           location.CityName,
				CityPlz:            location.CityPlz,
				SubdivisionId:      location.SubdivisionId,
				SubdivisionCode:    location.SubdivisionCode,
				CountryId:          location.CountryId,
				CountryCode:        location.CountryCode,
				ContinentId:        location.ContinentId,
				ContinentCode:      location.ContinentCode,
				AccuracyRadius:     location.AccuracyRadius,
				Latitude:           location.Latitude,
				Longitude:          location.Longitude,
				TimeZone:           location.TimeZone,
				SystemNumber:       location.SystemNumber,
				SystemOrganisation: location.SystemOrganisation,
				Network:            location.Network,
				IpAddress:          device.IpAddress,
			})
			if err != nil {
				return err
			}

			logs.Info(fmt.Sprintf("inserted location for ip %s and device %s", device.IpAddress, device.DeviceId))
		}
	}

	return nil
}
