// Package lsnpk01v1 decodes Dragino LSNPK01 v1 uplinks (soil NPK sensor).
package lsnpk01v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lsnpk01", "v1", decoders.New(Decode,
		decoders.Offer("bat_v", "V"),
		decoders.Offer("temp_c_ds18b20", "C"),
		decoders.Offer("n_soil", "mg/kg"),
		decoders.Offer("p_soil", "mg/kg"),
		decoders.Offer("k_soil", "mg/kg"),
		decoders.Offer("interrupt_flag", ""),
		decoders.Offer("message_type", ""),
	))
}

type Data struct {
	BatV          float64 `json:"bat_v"`
	TempCDS18B20  float64 `json:"temp_c_ds18b20"`
	NSoil         int     `json:"n_soil"`
	PSoil         int     `json:"p_soil"`
	KSoil         int     `json:"k_soil"`
	InterruptFlag int     `json:"interrupt_flag"`
	MessageType   int     `json:"message_type"`
}

func (d *Data) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float("bat_v", "V", d.BatV),
		decoders.Float("temp_c_ds18b20", "C", d.TempCDS18B20),
		decoders.Int("n_soil", "mg/kg", d.NSoil),
		decoders.Int("p_soil", "mg/kg", d.PSoil),
		decoders.Int("k_soil", "mg/kg", d.KSoil),
		decoders.Int("interrupt_flag", "", d.InterruptFlag),
		decoders.Int("message_type", "", d.MessageType),
	}
}

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("lsnpk01v1: payload too short: %d bytes (want >= 11)", len(b))
	}
	rawDS := uint16(b[2])<<8 | uint16(b[3])
	var ds int32
	if b[2]&0x80 != 0 {
		ds = int32(rawDS) | int32(-1&^0xFFFF)
	} else {
		ds = int32(rawDS)
	}
	return &Data{
		BatV:          float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		TempCDS18B20:  round(float64(ds)/10, 2),
		NSoil:         int(uint16(b[4])<<8 | uint16(b[5])),
		PSoil:         int(uint16(b[6])<<8 | uint16(b[7])),
		KSoil:         int(uint16(b[8])<<8 | uint16(b[9])),
		MessageType:   int(b[10] >> 4),
		InterruptFlag: int(b[10] & 0x0F),
	}, nil
}
