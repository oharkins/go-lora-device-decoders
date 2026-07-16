// Package ldds04v1 decodes Dragino LDDS04 v1 uplinks.
package ldds04v1

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "ldds04", "v1", decoders.New(Decode,
		decoders.Offer("bat_v", "V"),
		decoders.Offer("distance1_cm", "cm"),
		decoders.Offer("distance2_cm", "cm"),
		decoders.Offer("distance3_cm", "cm"),
		decoders.Offer("distance4_cm", "cm"),
		decoders.Offer("mes_type", ""),
	))
}

type Data struct {
	BatV        float64 `json:"bat_v"`
	EXTITrigger bool    `json:"exti_trigger"`
	Distance1CM float64 `json:"distance1_cm"`
	Distance2CM float64 `json:"distance2_cm"`
	Distance3CM float64 `json:"distance3_cm"`
	Distance4CM float64 `json:"distance4_cm"`
	MesType     int     `json:"mes_type"`
}

func (d *Data) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float("bat_v", "V", d.BatV),
		decoders.Float("distance1_cm", "cm", d.Distance1CM),
		decoders.Float("distance2_cm", "cm", d.Distance2CM),
		decoders.Float("distance3_cm", "cm", d.Distance3CM),
		decoders.Float("distance4_cm", "cm", d.Distance4CM),
		decoders.Int("mes_type", "", d.MesType),
	}
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("ldds04v1: payload too short: %d bytes (want >= 11)", len(b))
	}
	if b[0] == 0x03 && b[10] == 0x02 {
		return nil, decoders.ErrIgnored
	}
	return &Data{
		BatV:        float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		EXTITrigger: b[0]&0x80 != 0,
		Distance1CM: float64(uint16(b[2])<<8|uint16(b[3])) / 10,
		Distance2CM: float64(uint16(b[4])<<8|uint16(b[5])) / 10,
		Distance3CM: float64(uint16(b[6])<<8|uint16(b[7])) / 10,
		Distance4CM: float64(uint16(b[8])<<8|uint16(b[9])) / 10,
		MesType:     int(b[10]),
	}, nil
}
