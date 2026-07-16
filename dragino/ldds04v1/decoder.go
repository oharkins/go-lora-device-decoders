// Package ldds04v1 decodes Dragino LDDS04 v1 uplinks.
package ldds04v1

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "ldds04", "v1", decoders.New(Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.Distance1CM, decoders.Centimeter),
		decoders.Offer(decoders.Distance2CM, decoders.Centimeter),
		decoders.Offer(decoders.Distance3CM, decoders.Centimeter),
		decoders.Offer(decoders.Distance4CM, decoders.Centimeter),
		decoders.Offer(decoders.MessageType, ""),
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

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
		decoders.Float(decoders.Distance1CM, decoders.Centimeter, d.Distance1CM),
		decoders.Float(decoders.Distance2CM, decoders.Centimeter, d.Distance2CM),
		decoders.Float(decoders.Distance3CM, decoders.Centimeter, d.Distance3CM),
		decoders.Float(decoders.Distance4CM, decoders.Centimeter, d.Distance4CM),
		decoders.Int(decoders.MessageType, "", d.MesType),
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
