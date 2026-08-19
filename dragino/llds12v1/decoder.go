// Package llds12v1 decodes Dragino LLDS12 v1 uplinks.
package llds12v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "llds12", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	BatteryVoltage          float64 `json:"battery_voltage"`
	TemperatureDS18B20      float64 `json:"temperature_ds18b20"`
	DistanceCM              float64 `json:"distance_cm"`
	DistanceSignalStrength  int     `json:"distance_signal_strength"`
	TemperatureLidar        int     `json:"temperature_lidar"`
	InterruptFlag           int     `json:"interrupt_flag"`
	MessageType             int     `json:"message_type"`
}

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("llds12v1: payload too short: %d bytes (want >= 11)", len(b))
	}
	if b[0] == 0x03 && b[10] == 0x02 {
		return nil, nil
	}
	return &Data{
		BatteryVoltage:         float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		TemperatureDS18B20:     round(float64(int16(uint16(b[2])<<8|uint16(b[3])))/10, 2),
		DistanceCM:             float64(uint16(b[4])<<8|uint16(b[5])) / 10,
		DistanceSignalStrength: int(uint16(b[6])<<8 | uint16(b[7])),
		TemperatureLidar:       int(int8(b[9])),
		InterruptFlag:          int(b[8]),
		MessageType:            int(b[10]),
	}, nil
}
