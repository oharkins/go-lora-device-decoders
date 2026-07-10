// Package lds02v1 decodes Dragino LDS02 v1 uplinks.
package lds02v1

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lds02", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	BatV                  float64 `json:"bat_v"`
	Mod                   int     `json:"mod"`
	DoorOpenStatus        *int    `json:"door_open_status,omitempty"`
	WaterLeakStatus       *int    `json:"water_leak_status,omitempty"`
	DoorOpenTimes         *int    `json:"door_open_times,omitempty"`
	LastDoorOpenDuration  *int    `json:"last_door_open_duration,omitempty"`
	WaterLeakTimes        *int    `json:"water_leak_times,omitempty"`
	LastWaterLeakDuration *int    `json:"last_water_leak_duration,omitempty"`
	Alarm                 *int    `json:"alarm,omitempty"`
}

func ptr[T any](v T) *T { return &v }

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 10 {
		return nil, fmt.Errorf("lds02v1: payload too short: %d bytes (want >= 10)", len(b))
	}
	bat := float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000
	doorStatus := int((b[0] & 0x80) >> 7)
	waterStatus := int((b[0] & 0x40) >> 6)
	mod := int(b[2])
	alarm := int(b[9] & 0x01)
	d := &Data{BatV: bat, Mod: mod}
	switch mod {
	case 1:
		openTimes := int(uint32(b[3])<<16 | uint32(b[4])<<8 | uint32(b[5]))
		openDur := int(uint32(b[6])<<16 | uint32(b[7])<<8 | uint32(b[8]))
		d.DoorOpenStatus = ptr(doorStatus)
		d.DoorOpenTimes = ptr(openTimes)
		d.LastDoorOpenDuration = ptr(openDur)
		d.Alarm = ptr(alarm)
	case 2:
		leakTimes := int(uint32(b[3])<<16 | uint32(b[4])<<8 | uint32(b[5]))
		leakDur := int(uint32(b[6])<<16 | uint32(b[7])<<8 | uint32(b[8]))
		d.WaterLeakStatus = ptr(waterStatus)
		d.WaterLeakTimes = ptr(leakTimes)
		d.LastWaterLeakDuration = ptr(leakDur)
	case 3:
		d.DoorOpenStatus = ptr(doorStatus)
		d.WaterLeakStatus = ptr(waterStatus)
		d.Alarm = ptr(alarm)
	}
	return d, nil
}
