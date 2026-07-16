// Package ldds45v1 decodes Dragino LDDS45 v1 uplinks.
package ldds45v1

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "ldds45", "v1", decoders.New(Decode,
		decoders.Offer("battery_voltage", "V"),
		decoders.Offer("distance_mm", "mm"),
		decoders.Offer("interrupt_status", ""),
	))
}

type Data struct {
	BatV            float64 `json:"bat_v"`
	DistanceMM      *int    `json:"distance_mm,omitempty"`
	DistanceStatus  string  `json:"distance_status,omitempty"`
	InterruptStatus int     `json:"interrupt_status"`
}

func (d *Data) Measurements() []decoders.Measurement {
	measurements := []decoders.Measurement{
		decoders.Float("battery_voltage", "V", d.BatV),
	}
	measurements = decoders.AppendInt(measurements, "distance_mm", "mm", d.DistanceMM)
	measurements = append(measurements, decoders.Int("interrupt_status", "", d.InterruptStatus))
	return measurements
}

func ptr[T any](v T) *T { return &v }

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 2 {
		return nil, fmt.Errorf("ldds45v1: payload too short: %d bytes", len(b))
	}
	d := &Data{
		BatV:            float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		InterruptStatus: int(b[len(b)-1]),
	}
	if len(b) == 5 {
		v := int(uint16(b[2])<<8 | uint16(b[3]))
		if v < 20 {
			d.DistanceStatus = "Invalid Reading"
		} else {
			d.DistanceMM = ptr(v)
		}
	} else {
		d.DistanceStatus = "No Sensor"
	}
	return d, nil
}
