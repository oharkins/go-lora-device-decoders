// Package lse01v1 decodes Dragino LSE01 v1 uplinks (soil moisture/temp/conductivity).
package lse01v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lse01", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	BatteryVoltage     float64 `json:"battery_voltage"`
	TemperatureDS18B20 float64 `json:"temperature_ds18b20"`
	MoistureSoil       float64 `json:"moisture_soil"`
	TemperatureSoil    float64 `json:"temperature_soil"`
	ConductivitySoilUSCM float64 `json:"conductivity_soil_us_cm"`
}

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 10 {
		return nil, fmt.Errorf("lse01v1: payload too short: %d bytes (want >= 10)", len(b))
	}

	rawDS := uint16(b[2])<<8 | uint16(b[3])
	var ds int32
	if b[2]&0x80 != 0 {
		ds = int32(rawDS) | int32(-1&^0xFFFF)
	} else {
		ds = int32(rawDS)
	}

	rawSoil := uint16(b[6])<<8 | uint16(b[7])
	var soilTemp float64
	if rawSoil&0x8000 != 0 {
		soilTemp = round((float64(rawSoil)-0xFFFF)/100, 2)
	} else {
		soilTemp = round(float64(rawSoil)/100, 2)
	}

	return &Data{
		BatteryVoltage:       float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		TemperatureDS18B20:   round(float64(ds)/10, 2),
		MoistureSoil:         round(float64(uint16(b[4])<<8|uint16(b[5]))/100, 2),
		TemperatureSoil:      soilTemp,
		ConductivitySoilUSCM: round(float64(uint16(b[8])<<8|uint16(b[9]))/100, 2),
	}, nil
}
