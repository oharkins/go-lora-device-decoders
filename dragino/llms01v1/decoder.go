// Package llms01v1 decodes Dragino LLMS01 v1 uplinks (leaf moisture sensor).
package llms01v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "llms01", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	BatteryVoltage     float64 `json:"battery_voltage"`
	TemperatureDS18B20 float64 `json:"temperature_ds18b20"`
	MoistureLeaf       float64 `json:"moisture_leaf"`
	TemperatureLeaf    float64 `json:"temperature_leaf"`
	InterruptFlag      int     `json:"interrupt_flag"`
	MessageType        int     `json:"message_type"`
}

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func signedSoilTemp(hi, lo byte, div float64) float64 {
	raw := uint16(hi)<<8 | uint16(lo)
	if raw&0x8000 != 0 {
		return round((float64(raw)-0xFFFF)/div, 2)
	}
	return round(float64(raw)/div, 2)
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("llms01v1: payload too short: %d bytes (want >= 11)", len(b))
	}
	raw := uint16(b[2])<<8 | uint16(b[3])
	var rawVal int32
	if b[2]&0x80 != 0 {
		rawVal = int32(raw) | int32(-1&^0xFFFF)
	} else {
		rawVal = int32(raw)
	}
	return &Data{
		BatteryVoltage:     float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		TemperatureDS18B20: round(float64(rawVal)/10, 2),
		MoistureLeaf:       round(float64(uint16(b[4])<<8|uint16(b[5]))/10, 2),
		TemperatureLeaf:    signedSoilTemp(b[6], b[7], 10),
		InterruptFlag:      int(b[8]),
		MessageType:        int(b[10]),
	}, nil
}
