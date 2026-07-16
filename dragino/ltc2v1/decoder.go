// Package ltc2v1 decodes Dragino LTC2 v1 uplinks (dual-channel temp/resistance).
package ltc2v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "ltc2", "v1", decoders.New(
		Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.TempChannel1, decoders.Celsius),
		decoders.Offer(decoders.TempChannel2, decoders.Celsius),
		decoders.Offer(decoders.Resistance1, decoders.Ohm),
		decoders.Offer(decoders.Resistance2, decoders.Ohm),
	))
}

type Data struct {
	Ext          int      `json:"ext"`
	BatV         float64  `json:"bat_v"`
	TempChannel1 *float64 `json:"temp_channel1,omitempty"`
	TempChannel2 *float64 `json:"temp_channel2,omitempty"`
	ResChannel1  *float64 `json:"res_channel1,omitempty"`
	ResChannel2  *float64 `json:"res_channel2,omitempty"`
	SysTimestamp int64    `json:"sys_timestamp"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
	}
	ms = decoders.AppendFloat(ms, decoders.TempChannel1, decoders.Celsius, d.TempChannel1)
	ms = decoders.AppendFloat(ms, decoders.TempChannel2, decoders.Celsius, d.TempChannel2)
	ms = decoders.AppendFloat(ms, decoders.Resistance1, decoders.Ohm, d.ResChannel1)
	ms = decoders.AppendFloat(ms, decoders.Resistance2, decoders.Ohm, d.ResChannel2)
	return ms
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) != 11 {
		return nil, fmt.Errorf("ltc2v1: want 11 bytes, got %d", len(b))
	}
	pollStatus := (b[2] & 0x40) >> 6
	if pollStatus != 0 {
		return nil, decoders.ErrIgnored
	}
	ext := int(b[2] & 0x0F)
	d := &Data{
		Ext:          ext,
		BatV:         float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		SysTimestamp: int64(uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10])),
	}
	switch ext {
	case 0x01:
		d.TempChannel1 = ptr(round(float64(int16(uint16(b[3])<<8|uint16(b[4])))/100, 2))
		d.TempChannel2 = ptr(round(float64(int16(uint16(b[5])<<8|uint16(b[6])))/100, 2))
	case 0x02:
		d.TempChannel1 = ptr(round(float64(int16(uint16(b[3])<<8|uint16(b[4])))/10, 1))
		d.TempChannel2 = ptr(round(float64(int16(uint16(b[5])<<8|uint16(b[6])))/10, 1))
	case 0x03:
		d.ResChannel1 = ptr(round(float64(uint16(b[3])<<8|uint16(b[4]))/100, 2))
		d.ResChannel2 = ptr(round(float64(uint16(b[5])<<8|uint16(b[6]))/100, 2))
	}
	return d, nil
}
