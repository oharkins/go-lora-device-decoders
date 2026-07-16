// Package lgt92v1 decodes Dragino LGT92 v1 uplinks (GPS tracker, firmware 1.6.4).
package lgt92v1

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lgt92", "v1", decoders.New(
		Decode,
		decoders.Offer("lat", "deg"),
		decoders.Offer("lon", "deg"),
		decoders.Offer("altitude", "m"),
		decoders.Offer("battery_voltage", "V"),
		decoders.Offer("roll", "deg"),
		decoders.Offer("pitch", "deg"),
		decoders.Offer("hdop", ""),
	))
}

type Data struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Altitude    float64 `json:"altitude"`
	Accuracy    int     `json:"accuracy"`
	Roll        float64 `json:"roll"`
	Pitch       float64 `json:"pitch"`
	BatV        float64 `json:"bat_v"`
	AlarmStatus string  `json:"alarm_status"`
	MotionMode  string  `json:"motion_mode"`
	LEDUpDown   string  `json:"led_updown"`
	Firmware    int     `json:"firmware"`
	HDOP        float64 `json:"hdop"`
}

func (d *Data) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float("lat", "deg", d.Latitude),
		decoders.Float("lon", "deg", d.Longitude),
		decoders.Float("altitude", "m", d.Altitude),
		decoders.Float("battery_voltage", "V", d.BatV),
		decoders.Float("roll", "deg", d.Roll),
		decoders.Float("pitch", "deg", d.Pitch),
		decoders.Float("hdop", "", d.HDOP),
	}
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 18 {
		return nil, fmt.Errorf("lgt92v1: payload too short: %d bytes (want >= 18)", len(b))
	}

	lat := float64(int32(uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]))) / 1000000
	lon := float64(int32(uint32(b[4])<<24|uint32(b[5])<<16|uint32(b[6])<<8|uint32(b[7]))) / 1000000

	alarm := "FALSE"
	if b[8]&0x40 != 0 {
		alarm = "TRUE"
	}

	batV := float64((uint16(b[8]&0x3F)<<8)|uint16(b[9])) / 1000

	var motionMode string
	switch b[10] & 0xC0 {
	case 0x40:
		motionMode = "Move"
	case 0x80:
		motionMode = "Collide"
	case 0xC0:
		motionMode = "User"
	default:
		motionMode = "Disable"
	}

	led := "OFF"
	if b[10]&0x20 != 0 {
		led = "ON"
	}

	firmware := 160 + int(b[10]&0x1F)
	roll := float64(int16(uint16(b[11])<<8|uint16(b[12]))) / 100
	pitch := float64(int16(uint16(b[13])<<8|uint16(b[14]))) / 100

	var hdop float64
	if b[15] > 0 {
		hdop = float64(b[15]) / 100
	}

	altitude := float64(int16(uint16(b[16])<<8|uint16(b[17]))) / 100

	return &Data{
		Latitude:    lat,
		Longitude:   lon,
		Altitude:    altitude,
		Accuracy:    3,
		Roll:        roll,
		Pitch:       pitch,
		BatV:        batV,
		AlarmStatus: alarm,
		MotionMode:  motionMode,
		LEDUpDown:   led,
		Firmware:    firmware,
		HDOP:        hdop,
	}, nil
}
