// Package llms01v1 decodes Dragino LLMS01 v1 uplinks (leaf moisture sensor).
package llms01v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "llms01", "v1", decoders.New(Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.DSTemperature, decoders.Celsius),
		decoders.Offer(decoders.LeafMoisture, decoders.Percent),
		decoders.Offer(decoders.LeafTemp, decoders.Celsius),
		decoders.Offer(decoders.InterruptFlag, ""),
		decoders.Offer(decoders.MessageType, ""),
	))
}

type Data struct {
	BatV          float64 `json:"bat_v"`
	TempCDS18B20  float64 `json:"temp_c_ds18b20"`
	LeafMoisture  float64 `json:"leaf_moisture"`
	LeafTemp      float64 `json:"leaf_temp"`
	InterruptFlag int     `json:"interrupt_flag"`
	MessageType   int     `json:"message_type"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
		decoders.Float(decoders.DSTemperature, decoders.Celsius, d.TempCDS18B20),
		decoders.Float(decoders.LeafMoisture, decoders.Percent, d.LeafMoisture),
		decoders.Float(decoders.LeafTemp, decoders.Celsius, d.LeafTemp),
		decoders.Int(decoders.InterruptFlag, "", d.InterruptFlag),
		decoders.Int(decoders.MessageType, "", d.MessageType),
	}
}

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func signedSoilTemp(hi, lo byte, div float64) float64 {
	raw := uint16(hi)<<8 | uint16(lo)
	if raw&0x8000 != 0 {
		return round((float64(raw)-0xFFFF)/div, 2)
	}
	return round(float64(raw)/div, 2)
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("llms01v1: payload too short: %d bytes (want >= 11)", len(b))
	}
	raw := uint16(b[2])<<8 | uint16(b[3])
	var rawVal int32
	if b[2]&0x80 != 0 {
		rawVal = int32(raw) | int32(-1&^0xFFFF)
	} else {
		rawVal = int32(raw)
	}
	return &Data{
		BatV:          float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		TempCDS18B20:  round(float64(rawVal)/10, 2),
		LeafMoisture:  round(float64(uint16(b[4])<<8|uint16(b[5]))/10, 2),
		LeafTemp:      signedSoilTemp(b[6], b[7], 10),
		InterruptFlag: int(b[8]),
		MessageType:   int(b[10]),
	}, nil
}
