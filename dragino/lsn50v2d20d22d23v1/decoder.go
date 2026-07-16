// Package lsn50v2d20d22d23v1 decodes Dragino LSN50v2-D20/D22/D23 v1 uplinks.
// Supports mode 3 (DS18B20 triple temp) and mode 31 (ALARM config).
package lsn50v2d20d22d23v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lsn50v2-d20-d22-d23", "v1", decoders.New(Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.TempRed, decoders.Celsius),
		decoders.Offer(decoders.TempWhite, decoders.Celsius),
		decoders.Offer(decoders.TempBlack, decoders.Celsius),
		decoders.Offer(decoders.TempRedMin, decoders.Celsius),
		decoders.Offer(decoders.TempRedMax, decoders.Celsius),
		decoders.Offer(decoders.TempWhiteMin, decoders.Celsius),
		decoders.Offer(decoders.TempWhiteMax, decoders.Celsius),
		decoders.Offer(decoders.TempBlackMin, decoders.Celsius),
		decoders.Offer(decoders.TempBlackMax, decoders.Celsius),
	))
}

type Data struct {
	WorkMode     string   `json:"work_mode"`
	BatV         *float64 `json:"bat_v,omitempty"`
	AlarmStatus  *bool    `json:"alarm_status,omitempty"`
	TempRed      any      `json:"temp_red,omitempty"`
	TempWhite    any      `json:"temp_white,omitempty"`
	TempBlack    any      `json:"temp_black,omitempty"`
	TempRedMin   *int8    `json:"temp_red_min,omitempty"`
	TempRedMax   *int8    `json:"temp_red_max,omitempty"`
	TempWhiteMin *int8    `json:"temp_white_min,omitempty"`
	TempWhiteMax *int8    `json:"temp_white_max,omitempty"`
	TempBlackMin *int8    `json:"temp_black_min,omitempty"`
	TempBlackMax *int8    `json:"temp_black_max,omitempty"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	var measurements []decoders.Measurement
	measurements = decoders.AppendFloat(measurements, decoders.BatteryVoltage, decoders.Volt, d.BatV)
	measurements = appendAnyFloat(measurements, decoders.TempRed, decoders.Celsius, d.TempRed)
	measurements = appendAnyFloat(measurements, decoders.TempWhite, decoders.Celsius, d.TempWhite)
	measurements = appendAnyFloat(measurements, decoders.TempBlack, decoders.Celsius, d.TempBlack)
	if d.TempRedMin != nil {
		measurements = append(measurements, decoders.Int(decoders.TempRedMin, decoders.Celsius, int(*d.TempRedMin)))
	}
	if d.TempRedMax != nil {
		measurements = append(measurements, decoders.Int(decoders.TempRedMax, decoders.Celsius, int(*d.TempRedMax)))
	}
	if d.TempWhiteMin != nil {
		measurements = append(measurements, decoders.Int(decoders.TempWhiteMin, decoders.Celsius, int(*d.TempWhiteMin)))
	}
	if d.TempWhiteMax != nil {
		measurements = append(measurements, decoders.Int(decoders.TempWhiteMax, decoders.Celsius, int(*d.TempWhiteMax)))
	}
	if d.TempBlackMin != nil {
		measurements = append(measurements, decoders.Int(decoders.TempBlackMin, decoders.Celsius, int(*d.TempBlackMin)))
	}
	if d.TempBlackMax != nil {
		measurements = append(measurements, decoders.Int(decoders.TempBlackMax, decoders.Celsius, int(*d.TempBlackMax)))
	}
	return measurements
}

func appendAnyFloat(dst []decoders.Measurement, name, unit string, v any) []decoders.Measurement {
	if f, ok := v.(float64); ok {
		return append(dst, decoders.Float(name, unit, f))
	}
	return dst
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func nullableTemp(hi, lo byte, div float64) any {
	if hi == 0xFF && lo == 0xFF {
		return nil
	}
	return round(float64(int16(uint16(hi)<<8|uint16(lo)))/div, 1)
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) != 11 {
		return nil, fmt.Errorf("lsn50v2d20d22d23v1: want 11 bytes, got %d", len(b))
	}
	mode := (b[6] & 0x7C) >> 2
	d := &Data{}
	switch mode {
	case 3:
		d.WorkMode = "DS18B20"
		batV := float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		d.BatV = ptr(batV)
		alarm := b[6]&0x01 != 0
		d.AlarmStatus = ptr(alarm)
		d.TempRed = nullableTemp(b[2], b[3], 10)
		d.TempWhite = nullableTemp(b[7], b[8], 10)
		d.TempBlack = nullableTemp(b[9], b[10], 10)
	case 31:
		d.WorkMode = "ALARM"
		d.TempRedMin = ptr(int8(b[4]))
		d.TempRedMax = ptr(int8(b[5]))
		d.TempWhiteMin = ptr(int8(b[7]))
		d.TempWhiteMax = ptr(int8(b[8]))
		d.TempBlackMin = ptr(int8(b[9]))
		d.TempBlackMax = ptr(int8(b[10]))
	default:
		return nil, fmt.Errorf("lsn50v2d20d22d23v1: unsupported mode %d", mode)
	}
	return d, nil
}
