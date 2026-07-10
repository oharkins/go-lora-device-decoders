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
	BatV         float64 `json:"bat_v"`
	TempCDS18B20 float64 `json:"temp_c_ds18b20"`
	WaterSoil    float64 `json:"water_soil"`
	TempSoil     float64 `json:"temp_soil"`
	ConductSoil  float64 `json:"conduct_soil"`
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
		BatV:         float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		TempCDS18B20: round(float64(ds)/10, 2),
		WaterSoil:    round(float64(uint16(b[4])<<8|uint16(b[5]))/100, 2),
		TempSoil:     soilTemp,
		ConductSoil:  round(float64(uint16(b[8])<<8|uint16(b[9]))/100, 2),
	}, nil
}
