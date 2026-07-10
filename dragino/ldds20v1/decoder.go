// Package ldds20v1 decodes Dragino LDDS20 v1 uplinks.
package ldds20v1

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "ldds20", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	BatV            float64 `json:"bat_v"`
	DistanceMM      *int    `json:"distance_mm,omitempty"`
	DistanceStatus  string  `json:"distance_status,omitempty"`
	InterruptStatus int     `json:"interrupt_status"`
}

func ptr[T any](v T) *T { return &v }

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 2 {
		return nil, fmt.Errorf("ldds20v1: payload too short: %d bytes", len(b))
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
