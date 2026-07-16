// Package doorleak decodes the shared Dragino LDS/LWL door / water-leak uplink layout.
package doorleak

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

// Data is the shared decoded door / water-leak uplink.
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

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
		decoders.Int(decoders.Mode, "", d.Mod),
	}
	ms = decoders.AppendInt(ms, decoders.DoorOpenStatus, "", d.DoorOpenStatus)
	ms = decoders.AppendInt(ms, decoders.WaterLeakStatus, "", d.WaterLeakStatus)
	ms = decoders.AppendInt(ms, decoders.DoorOpenTimes, decoders.Count, d.DoorOpenTimes)
	ms = decoders.AppendInt(ms, decoders.DoorOpenDuration, decoders.Second, d.LastDoorOpenDuration)
	ms = decoders.AppendInt(ms, decoders.WaterLeakTimes, decoders.Count, d.WaterLeakTimes)
	ms = decoders.AppendInt(ms, decoders.WaterLeakDuration, decoders.Second, d.LastWaterLeakDuration)
	ms = decoders.AppendInt(ms, decoders.Alarm, "", d.Alarm)
	return ms
}

// Offers are the measurements LDS/LWL devices can produce.
func Offers() []decoders.Offering {
	return []decoders.Offering{
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.Mode, ""),
		decoders.Offer(decoders.DoorOpenStatus, ""),
		decoders.Offer(decoders.WaterLeakStatus, ""),
		decoders.Offer(decoders.DoorOpenTimes, decoders.Count),
		decoders.Offer(decoders.DoorOpenDuration, decoders.Second),
		decoders.Offer(decoders.WaterLeakTimes, decoders.Count),
		decoders.Offer(decoders.WaterLeakDuration, decoders.Second),
		decoders.Offer(decoders.Alarm, ""),
	}
}

func ptr[T any](v T) *T { return &v }

// Decode parses a shared LDS/LWL payload. name is used in error messages.
func Decode(name string, u decoders.Uplink) (*Data, error) {
	b := u.Payload
	if len(b) < 10 {
		return nil, fmt.Errorf("%s: payload too short: %d bytes (want >= 10)", name, len(b))
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
