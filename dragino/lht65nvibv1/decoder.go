// Package lht65nvibv1 decodes Dragino LHT65N-VIB v1 uplinks (vibration sensor).
package lht65nvibv1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lht65n-vib", "v1", decoders.New(Decode,
		decoders.Offer("battery_voltage", "V"),
		decoders.Offer("vibration_count", "count"),
		decoders.Offer("work_minutes", "min"),
		decoders.Offer("temperature", "C"),
		decoders.Offer("humidity", "%"),
		decoders.Offer("acceleration_x", "g"),
		decoders.Offer("acceleration_y", "g"),
		decoders.Offer("acceleration_z", "g"),
		decoders.Offer("max_acceleration_x", "g"),
		decoders.Offer("max_acceleration_y", "g"),
		decoders.Offer("max_acceleration_z", "g"),
	))
}

// Data is the decoded payload for FPort 2 (vibration data).
type Data struct {
	NodeType string   `json:"node_type"`
	BatV     float64  `json:"bat_v"`
	Mod      int      `json:"mod"`
	VibCount *uint32  `json:"vib_count,omitempty"`
	WorkMin  *uint32  `json:"work_min,omitempty"`
	TempCSHT *float64 `json:"temp_c_sht,omitempty"`
	HumSHT   *float64 `json:"hum_sht,omitempty"`
	Alarm    string   `json:"alarm"`
	TDC      string   `json:"tdc"`
}

// Measurements returns the numeric readings decoded from this uplink.
func (d *Data) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Float("battery_voltage", "V", d.BatV),
	}
	if d.VibCount != nil {
		ms = append(ms, decoders.Measurement{Name: "vibration_count", Unit: "count", Value: float64(*d.VibCount)})
	}
	if d.WorkMin != nil {
		ms = append(ms, decoders.Measurement{Name: "work_minutes", Unit: "min", Value: float64(*d.WorkMin)})
	}
	ms = decoders.AppendFloat(ms, "temperature", "C", d.TempCSHT)
	ms = decoders.AppendFloat(ms, "humidity", "%", d.HumSHT)
	return ms
}

// AccelData is the decoded payload for FPort 9 (peak acceleration).
type AccelData struct {
	NodeType string  `json:"node_type"`
	BatV     float64 `json:"bat_v"`
	MaxAccXG float64 `json:"max_acc_x_g"`
	MaxAccYG float64 `json:"max_acc_y_g"`
	MaxAccZG float64 `json:"max_acc_z_g"`
}

// Measurements returns the numeric readings decoded from this acceleration uplink.
func (d *AccelData) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float("battery_voltage", "V", d.BatV),
		decoders.Float("max_acceleration_x", "g", d.MaxAccXG),
		decoders.Float("max_acceleration_y", "g", d.MaxAccYG),
		decoders.Float("max_acceleration_z", "g", d.MaxAccZG),
	}
}

// DeviceInfo is the decoded payload for FPort 5.
type DeviceInfo struct {
	NodeType        string  `json:"node_type"`
	SensorModel     string  `json:"sensor_model"`
	FirmwareVersion string  `json:"firmware_version"`
	FrequencyBand   string  `json:"frequency_band"`
	SubBand         any     `json:"sub_band"`
	Bat             float64 `json:"bat"`
}

// Measurements returns the numeric readings decoded from this device-info uplink.
func (d *DeviceInfo) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float("battery_voltage", "V", d.Bat),
	}
}

// DatalogEntry holds one accelerometer record from the FPort 7 data log.
type DatalogEntry struct {
	AccXG float64 `json:"acc_x_g"`
	AccYG float64 `json:"acc_y_g"`
	AccZG float64 `json:"acc_z_g"`
}

// Measurements returns the numeric acceleration readings from this data-log entry.
func (d DatalogEntry) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float("acceleration_x", "g", d.AccXG),
		decoders.Float("acceleration_y", "g", d.AccYG),
		decoders.Float("acceleration_z", "g", d.AccZG),
	}
}

// DatalogData is the decoded FPort 7 payload.
type DatalogData struct {
	NodeType string         `json:"node_type"`
	BatV     float64        `json:"bat_v"`
	Datalog  []DatalogEntry `json:"datalog"`
}

// Measurements returns the numeric readings decoded from this data-log uplink.
func (d *DatalogData) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Float("battery_voltage", "V", d.BatV),
	}
	for _, entry := range d.Datalog {
		ms = append(ms, entry.Measurements()...)
	}
	return ms
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func signed16(hi, lo byte) float64 {
	return float64(int16(uint16(hi)<<8 | uint16(lo)))
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
	case 0x0F:
		return "AS923_4"
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

// Decode decodes an LHT65N-VIB v1 uplink payload.
func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload

	switch u.FPort {
	case 2:
		if len(b) < 11 {
			return nil, fmt.Errorf("lht65nvibv1: port 2 payload too short: %d bytes", len(b))
		}
		batV := float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		mod := int((b[2] >> 2) & 0x07)
		alarm := "FALSE"
		if b[2]&0x01 != 0 {
			alarm = "TRUE"
		}
		tdc := "NO"
		if b[2]&0x02 != 0 {
			tdc = "YES"
		}
		d := &Data{
			NodeType: "LHT65N-VIB",
			BatV:     batV,
			Mod:      mod,
			Alarm:    alarm,
			TDC:      tdc,
		}
		switch mod {
		case 1:
			vc := uint32(b[3])<<24 | uint32(b[4])<<16 | uint32(b[5])<<8 | uint32(b[6])
			wm := uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10])
			d.VibCount = ptr(vc)
			d.WorkMin = ptr(wm)
		case 2:
			vc := uint32(b[3])<<24 | uint32(b[4])<<16 | uint32(b[5])<<8 | uint32(b[6])
			d.VibCount = ptr(vc)
			d.TempCSHT = ptr(round(signed16(b[7], b[8])/100, 2))
			d.HumSHT = ptr(round(float64((uint16(b[9])<<8|uint16(b[10]))&0xFFF)/10, 1))
		case 3:
			d.TempCSHT = ptr(round(signed16(b[3], b[4])/100, 2))
			d.HumSHT = ptr(round(float64((uint16(b[5])<<8|uint16(b[6]))&0xFFF)/10, 1))
			wm := uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10])
			d.WorkMin = ptr(wm)
		}
		return d, nil

	case 5:
		if len(b) < 7 {
			return nil, fmt.Errorf("lht65nvibv1: port 5 payload too short: %d bytes", len(b))
		}
		var sensor string
		if b[0] == 0x3F {
			sensor = "LHT65N-VIB"
		}
		var subBand any
		if b[4] == 0xFF {
			subBand = "NULL"
		} else {
			subBand = int(b[4])
		}
		firmVer := fmt.Sprintf("%d.%d.%d", b[1]&0x0F, (b[2]>>4)&0x0F, b[2]&0x0F)
		return &DeviceInfo{
			NodeType:        "LHT65N-VIB",
			SensorModel:     sensor,
			FirmwareVersion: firmVer,
			FrequencyBand:   freqBand(b[3]),
			SubBand:         subBand,
			Bat:             float64(uint16(b[5])<<8|uint16(b[6])) / 1000,
		}, nil

	case 7:
		if len(b) < 8 {
			return nil, fmt.Errorf("lht65nvibv1: port 7 payload too short: %d bytes", len(b))
		}
		batV := float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		var entries []DatalogEntry
		for k := 2; k+5 < len(b); k += 6 {
			entries = append(entries, DatalogEntry{
				AccXG: round(signed16(b[k], b[k+1])/1000, 3),
				AccYG: round(signed16(b[k+2], b[k+3])/1000, 3),
				AccZG: round(signed16(b[k+4], b[k+5])/1000, 3),
			})
		}
		return &DatalogData{NodeType: "LHT65N-VIB", BatV: batV, Datalog: entries}, nil

	case 9:
		if len(b) < 8 {
			return nil, fmt.Errorf("lht65nvibv1: port 9 payload too short: %d bytes", len(b))
		}
		return &AccelData{
			NodeType: "LHT65N-VIB",
			BatV:     float64(uint16(b[0])<<8|uint16(b[1])) / 1000,
			MaxAccXG: round(signed16(b[2], b[3])/1000, 3),
			MaxAccYG: round(signed16(b[4], b[5])/1000, 3),
			MaxAccZG: round(signed16(b[6], b[7])/1000, 3),
		}, nil

	default:
		return nil, fmt.Errorf("lht65nvibv1: unsupported FPort %d", u.FPort)
	}
}
