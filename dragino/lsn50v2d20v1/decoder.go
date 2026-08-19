// Package lsn50v2d20v1 decodes Dragino LSN50v2-D20 v1 uplinks.
package lsn50v2d20v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lsn50v2-d20", "v1", decoders.DecoderFunc(Decode))
}

type Data struct {
	WorkMode               string   `json:"work_mode,omitempty"`
	BatteryVoltage         float64  `json:"battery_voltage"`
	Temperature1           *float64 `json:"temperature_1,omitempty"`
	ADC0V                  *float64 `json:"adc_0_v,omitempty"`
	ADC1V                  *float64 `json:"adc_1_v,omitempty"`
	ADC4V                  *float64 `json:"adc_4_v,omitempty"`
	DigitalInput           string   `json:"digital_input,omitempty"`
	InterruptTrigger       string   `json:"interrupt_trigger,omitempty"`
	DoorStatus             string   `json:"door_status,omitempty"`
	TemperatureSHT         *float64 `json:"temperature_sht,omitempty"`
	HumidityExternal       *float64 `json:"humidity_external,omitempty"`
	IlluminationLux        *int     `json:"illumination_lux,omitempty"`
	DistanceCM             *float64 `json:"distance_cm,omitempty"`
	DistanceSignalStrength *float64 `json:"distance_signal_strength,omitempty"`
	Temperature2           *float64 `json:"temperature_2,omitempty"`
	Temperature3           *float64 `json:"temperature_3,omitempty"`
	Weight                 *int     `json:"weight,omitempty"`
	Count                  *int     `json:"count,omitempty"`
	TempC1Min              *int8    `json:"temp_c1_min,omitempty"`
	TempC1Max              *int8    `json:"temp_c1_max,omitempty"`
	SHTEmpMin              *int8    `json:"sht_temp_min,omitempty"`
	SHTEmpMax              *int8    `json:"sht_temp_max,omitempty"`
	SHTHumMin              *uint8   `json:"sht_hum_min,omitempty"`
	SHTHumMax              *uint8   `json:"sht_hum_max,omitempty"`
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func signed16(hi, lo byte) float64 {
	return float64(int16(uint16(hi)<<8 | uint16(lo)))
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("lsn50v2d20v1: payload too short: %d bytes", len(b))
	}
	mode := (b[6] & 0x7C) >> 2
	d := &Data{}

	if mode != 2 && mode != 31 {
		d.BatteryVoltage = float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		d.Temperature1 = ptr(round(signed16(b[2], b[3])/10, 2))
		d.ADC0V = ptr(float64(uint16(b[4])<<8|uint16(b[5])) / 1000)
		if b[6]&0x02 != 0 {
			d.DigitalInput = "H"
		} else {
			d.DigitalInput = "L"
		}
		if mode != 6 {
			if b[6]&0x01 != 0 {
				d.InterruptTrigger = "TRUE"
			} else {
				d.InterruptTrigger = "FALSE"
			}
			if b[6]&0x80 != 0 {
				d.DoorStatus = "CLOSE"
			} else {
				d.DoorStatus = "OPEN"
			}
		}
	}

	switch mode {
	case 0:
		d.WorkMode = "IIC"
		iic9 := uint16(b[9])<<8 | uint16(b[10])
		if iic9 == 0 {
			v := int(int16(uint16(b[7])<<8 | uint16(b[8])))
			d.IlluminationLux = ptr(v)
		} else {
			d.TemperatureSHT = ptr(round(signed16(b[7], b[8])/10, 2))
			d.HumidityExternal = ptr(round(float64(uint16(b[9])<<8|uint16(b[10]))/10, 1))
		}
	case 1:
		d.WorkMode = "Distance"
		d.DistanceCM = ptr(round(float64(uint16(b[7])<<8|uint16(b[8]))/10, 1))
		sig := uint16(b[9])<<8 | uint16(b[10])
		if sig != 0xFFFF {
			s := round(float64(sig), 0)
			d.DistanceSignalStrength = ptr(s)
		}
	case 2:
		d.WorkMode = "3ADC"
		d.BatteryVoltage = float64(b[11]) / 10
		d.ADC0V = ptr(float64(uint16(b[0])<<8|uint16(b[1])) / 1000)
		d.ADC1V = ptr(float64(uint16(b[2])<<8|uint16(b[3])) / 1000)
		d.ADC4V = ptr(float64(uint16(b[4])<<8|uint16(b[5])) / 1000)
		if b[6]&0x02 != 0 {
			d.DigitalInput = "H"
		} else {
			d.DigitalInput = "L"
		}
		if b[6]&0x01 != 0 {
			d.InterruptTrigger = "TRUE"
		} else {
			d.InterruptTrigger = "FALSE"
		}
		if b[6]&0x80 != 0 {
			d.DoorStatus = "CLOSE"
		} else {
			d.DoorStatus = "OPEN"
		}
		iic9 := uint16(b[9])<<8 | uint16(b[10])
		if iic9 == 0 {
			v := int(int16(uint16(b[7])<<8 | uint16(b[8])))
			d.IlluminationLux = ptr(v)
		} else {
			d.TemperatureSHT = ptr(round(signed16(b[7], b[8])/10, 2))
			d.HumidityExternal = ptr(round(float64(uint16(b[9])<<8|uint16(b[10]))/10, 1))
		}
	case 3:
		d.WorkMode = "3DS18B20"
		d.Temperature2 = ptr(round(signed16(b[7], b[8])/10, 2))
		d.Temperature3 = ptr(round(signed16(b[9], b[10])/10, 1))
	case 4:
		d.WorkMode = "Weight"
		w := int(int16(uint16(b[7])<<8 | uint16(b[8])))
		d.Weight = ptr(w)
	case 5:
		d.WorkMode = "Count"
		c := int(uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10]))
		d.Count = ptr(c)
	case 31:
		d.WorkMode = "ALARM"
		d.BatteryVoltage = float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		d.Temperature1 = ptr(round(signed16(b[2], b[3])/10, 2))
		d.TempC1Min = ptr(int8(b[4]))
		d.TempC1Max = ptr(int8(b[5]))
		d.SHTEmpMin = ptr(int8(b[7]))
		d.SHTEmpMax = ptr(int8(b[8]))
		d.SHTHumMin = ptr(b[9])
		d.SHTHumMax = ptr(b[10])
	}
	return d, nil
}
