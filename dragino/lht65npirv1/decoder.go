// Package lht65npirv1 decodes Dragino LHT65N-PIR v1 uplinks.
// Extends LHT65N with PIR motion sensor support (Ext=14).
package lht65npirv1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lht65n-pir", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	NodeType     string   `json:"node_type"`
	BatV         *float64 `json:"bat_v,omitempty"`
	BatStatus    string   `json:"bat_status,omitempty"`
	TempCSHT     *float64 `json:"temp_c_sht,omitempty"`
	HumSHT       *float64 `json:"hum_sht,omitempty"`
	NoConnect    string   `json:"no_connect,omitempty"`
	ExtSensor    string   `json:"ext_sensor,omitempty"`
	WorkMode     string   `json:"work_mode,omitempty"`
	TempCDS      *float64 `json:"temp_c_ds,omitempty"`
	TempCTMP117  *float64 `json:"temp_c_tmp117,omitempty"`
	ExtiPinLevel string   `json:"exti_pin_level,omitempty"`
	ExtiStatus   string   `json:"exti_status,omitempty"`
	ExitCount    *int64   `json:"exit_count,omitempty"`
	ExitDuration *int     `json:"exit_duration,omitempty"`
	MoveCount    *int     `json:"move_count,omitempty"`
	ILLLx        *int     `json:"ill_lx,omitempty"`
	ADCV         *float64 `json:"adc_v,omitempty"`
	SysTimestamp *int64   `json:"sys_timestamp,omitempty"`
	ExtTempCSHT  *float64 `json:"ext_temp_c_sht,omitempty"`
	ExtHumSHT    *float64 `json:"ext_hum_sht,omitempty"`
	ID           string   `json:"id,omitempty"`
}

type DeviceInfo struct {
	NodeType        string  `json:"node_type"`
	SensorModel     string  `json:"sensor_model"`
	FirmwareVersion string  `json:"firmware_version"`
	FrequencyBand   string  `json:"frequency_band"`
	SubBand         any     `json:"sub_band"`
	Bat             float64 `json:"bat"`
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func signed16(hi, lo byte) float64 {
	return float64(int16(uint16(hi)<<8 | uint16(lo)))
}

func hexID(bb ...byte) string {
	s := ""
	for _, b := range bb {
		s += fmt.Sprintf("%02x ", b)
	}
	return s
}

func freqBand(code byte) string {
	switch code {
	case 0x01:
		return "EU868"
	case 0x02:
		return "US915"
	case 0x03:
		return "IN865"
	case 0x04:
		return "AU915"
	case 0x05:
		return "KZ865"
	case 0x06:
		return "RU864"
	case 0x07:
		return "AS923"
	case 0x08:
		return "AS923_1"
	case 0x09:
		return "AS923_2"
	case 0x0A:
		return "AS923_3"
	case 0x0B:
		return "CN470"
	case 0x0C:
		return "EU433"
	case 0x0D:
		return "KR920"
	case 0x0E:
		return "MA869"
	default:
		return "Unknown"
	}
}

func batStatusStr(code int) string {
	switch code {
	case 3:
		return "Good"
	case 2:
		return "OK"
	case 1:
		return "Low"
	default:
		return "Ultra Low"
	}
}

// Decode decodes an LHT65N-PIR v1 uplink payload.
func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload

	if u.FPort == 5 {
		if len(b) < 7 {
			return nil, fmt.Errorf("lht65npirv1: port 5 payload too short: %d bytes", len(b))
		}
		var sensor string
		if b[0] == 0x0B {
			sensor = "LHT65N"
		} else if b[0] == 0x1A {
			sensor = "LHT65N-PIR"
		}
		var subBand any
		if b[4] == 0xFF {
			subBand = "NULL"
		} else {
			subBand = int(b[4])
		}
		firmVer := fmt.Sprintf("%d.%d.%d", b[1]&0x0F, (b[2]>>4)&0x0F, b[2]&0x0F)
		return &DeviceInfo{
			NodeType:        "LHT65N-PIR",
			SensorModel:     sensor,
			FirmwareVersion: firmVer,
			FrequencyBand:   freqBand(b[3]),
			SubBand:         subBand,
			Bat:             float64(uint16(b[5])<<8|uint16(b[6])) / 1000,
		}, nil
	}

	if len(b) < 11 {
		return nil, fmt.Errorf("lht65npirv1: payload too short: %d bytes (want >= 11)", len(b))
	}

	ext := b[6]
	pollStatus := (b[6] >> 6) & 0x03
	connect := (b[6] & 0x80) >> 7

	if pollStatus != 0 {
		return map[string]any{"node_type": "LHT65N-PIR", "datalog": true}, nil
	}

	d := &Data{NodeType: "LHT65N-PIR"}

	if ext == 0x09 {
		d.TempCDS = ptr(round(signed16(b[0], b[1])/100, 2))
		d.BatStatus = fmt.Sprintf("%d", int(b[4]>>6))
	} else {
		batV := float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000
		d.BatV = ptr(batV)
		d.BatStatus = batStatusStr(int(b[0] >> 6))
	}

	// SHT sensor — skipped for PIR (0x0E) and ID modes
	if ext != 0x0F && ext != 0x10 && ext != 0x20 && ext != 0x0E {
		d.TempCSHT = ptr(round(signed16(b[2], b[3])/100, 2))
		d.HumSHT = ptr(round(float64((uint16(b[4])<<8|uint16(b[5]))&0xFFF)/10, 1))
	}

	if connect == 1 {
		d.NoConnect = "Sensor no connection"
	}

	switch ext {
	case 0:
		d.ExtSensor = "No external sensor"
	case 1:
		d.ExtSensor = "Temperature Sensor"
		d.TempCDS = ptr(round(signed16(b[7], b[8])/100, 2))
	case 2:
		d.ExtSensor = "Temperature Sensor"
		d.TempCTMP117 = ptr(round(signed16(b[7], b[8])/100, 2))
	case 4:
		d.WorkMode = "Interrupt Sensor send"
		if b[7] != 0 {
			d.ExtiPinLevel = "High"
		} else {
			d.ExtiPinLevel = "Low"
		}
		if b[8] != 0 {
			d.ExtiStatus = "True"
		} else {
			d.ExtiStatus = "False"
		}
		if len(b) > 11 {
			c := int64(uint32(b[9])<<16 | uint32(b[10])<<8 | uint32(b[11]))
			d.ExitCount = ptr(c)
		}
		if len(b) > 14 {
			dur := int(uint32(b[12])<<16 | uint32(b[13])<<8 | uint32(b[14]))
			d.ExitDuration = ptr(dur)
		}
	case 5:
		d.WorkMode = "Illumination Sensor"
		v := int(uint16(b[7])<<8 | uint16(b[8]))
		d.ILLLx = ptr(v)
	case 6:
		d.WorkMode = "ADC Sensor"
		d.ADCV = ptr(float64(uint16(b[7])<<8|uint16(b[8])) / 1000)
	case 7:
		d.WorkMode = "Interrupt Sensor count"
		c := int64(uint16(b[7])<<8 | uint16(b[8]))
		d.ExitCount = ptr(c)
	case 8:
		d.WorkMode = "Interrupt Sensor count"
		c := int64(uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10]))
		d.ExitCount = ptr(c)
	case 9:
		d.WorkMode = "DS18B20 & timestamp"
		ts := int64(uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10]))
		d.SysTimestamp = ptr(ts)
	case 11:
		d.WorkMode = "SHT31 Sensor"
		d.ExtTempCSHT = ptr(round(signed16(b[7], b[8])/100, 2))
		d.ExtHumSHT = ptr(round(float64((uint16(b[9])<<8|uint16(b[10]))&0xFFF)/10, 1))
	case 0x0E:
		d.WorkMode = "PIR Sensor"
		if b[7]&0x01 != 0 {
			d.ExtiPinLevel = "Activity"
		} else {
			d.ExtiPinLevel = "No activity"
		}
		mc := int(uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10]))
		d.MoveCount = ptr(mc)
	case 0x10:
		d.WorkMode = "SHT31ID"
		d.ID = hexID(b[2], b[3], b[4], b[5])
		d.ExtTempCSHT = ptr(round(signed16(b[7], b[8])/100, 2))
		d.ExtHumSHT = ptr(round(float64((uint16(b[9])<<8|uint16(b[10]))&0xFFF)/10, 1))
	case 0x20:
		d.WorkMode = "NE117ID"
		d.ID = hexID(b[2], b[3], b[4], b[5], b[9], b[10])
		d.TempCTMP117 = ptr(round(signed16(b[7], b[8])/100, 2))
	case 15:
		d.WorkMode = "DS18B20ID"
		d.ID = hexID(b[2], b[3], b[4], b[5], b[7], b[8], b[9], b[10])
	}

	return d, nil
}
