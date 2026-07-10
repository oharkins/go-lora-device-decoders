// Package lsn50v2d20d22d23v1 decodes Dragino LSN50v2-D20/D22/D23 v1 uplinks.
// Supports mode 3 (DS18B20 triple temp) and mode 31 (ALARM config).
package lsn50v2d20d22d23v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lsn50v2-d20-d22-d23", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	WorkMode    string   `json:"work_mode"`
	BatV        *float64 `json:"bat_v,omitempty"`
	AlarmStatus *bool    `json:"alarm_status,omitempty"`
	TempRed     any      `json:"temp_red,omitempty"`
	TempWhite   any      `json:"temp_white,omitempty"`
	TempBlack   any      `json:"temp_black,omitempty"`
	TempRedMin  *int8    `json:"temp_red_min,omitempty"`
	TempRedMax  *int8    `json:"temp_red_max,omitempty"`
	TempWhiteMin *int8   `json:"temp_white_min,omitempty"`
	TempWhiteMax *int8   `json:"temp_white_max,omitempty"`
	TempBlackMin *int8   `json:"temp_black_min,omitempty"`
	TempBlackMax *int8   `json:"temp_black_max,omitempty"`
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
