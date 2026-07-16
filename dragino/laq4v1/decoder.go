// Package laq4v1 decodes Dragino LAQ4 v1 uplinks.
package laq4v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "laq4", "v1", decoders.New(
		Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.TVOCPPB, decoders.PPB),
		decoders.Offer(decoders.CO2PPM, decoders.PPM),
		decoders.Offer(decoders.Temperature, decoders.Celsius),
		decoders.Offer(decoders.Humidity, decoders.Percent),
	))
}

type Data struct {
	BatV       float64  `json:"bat_v"`
	WorkMode   string   `json:"work_mode"`
	Alarm      *bool    `json:"alarm_status,omitempty"`
	TVOCPPB    *int     `json:"tvoc_ppb,omitempty"`
	CO2PPM     *int     `json:"co2_ppm,omitempty"`
	TempCSHT   *float64 `json:"temp_c_sht,omitempty"`
	HumSHT     *float64 `json:"hum_sht,omitempty"`
	SHTTempMin *int8    `json:"sht_temp_min,omitempty"`
	SHTTempMax *int8    `json:"sht_temp_max,omitempty"`
	SHTHumMin  *uint8   `json:"sht_hum_min,omitempty"`
	SHTHumMax  *uint8   `json:"sht_hum_max,omitempty"`
	CO2Min     *int     `json:"co2_min,omitempty"`
	CO2Max     *int     `json:"co2_max,omitempty"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
	}
	ms = decoders.AppendInt(ms, decoders.TVOCPPB, decoders.PPB, d.TVOCPPB)
	ms = decoders.AppendInt(ms, decoders.CO2PPM, decoders.PPM, d.CO2PPM)
	ms = decoders.AppendFloat(ms, decoders.Temperature, decoders.Celsius, d.TempCSHT)
	ms = decoders.AppendFloat(ms, decoders.Humidity, decoders.Percent, d.HumSHT)
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
		return nil, fmt.Errorf("laq4v1: want 11 bytes, got %d", len(b))
	}
	mode := (b[2] & 0x7C) >> 2
	d := &Data{
		BatV: float64(uint16(b[0])<<8|uint16(b[1])) / 1000,
	}
	switch mode {
	case 1:
		d.WorkMode = "CO2"
		alarm := b[2]&0x01 != 0
		tvoc := int(uint16(b[3])<<8 | uint16(b[4]))
		co2 := int(uint16(b[5])<<8 | uint16(b[6]))
		temp := round(float64(int16(uint16(b[7])<<8|uint16(b[8])))/10, 2)
		hum := round(float64(uint16(b[9])<<8|uint16(b[10]))/10, 1)
		d.Alarm = ptr(alarm)
		d.TVOCPPB = ptr(tvoc)
		d.CO2PPM = ptr(co2)
		d.TempCSHT = ptr(temp)
		d.HumSHT = ptr(hum)
	case 31:
		d.WorkMode = "ALARM"
		d.SHTTempMin = ptr(int8(b[3]))
		d.SHTTempMax = ptr(int8(b[4]))
		d.SHTHumMin = ptr(b[5])
		d.SHTHumMax = ptr(b[6])
		d.CO2Min = ptr(int(uint16(b[7])<<8 | uint16(b[8])))
		d.CO2Max = ptr(int(uint16(b[9])<<8 | uint16(b[10])))
	}
	return d, nil
}
