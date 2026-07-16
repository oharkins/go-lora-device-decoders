// Package lsn50v1 decodes Dragino LSN50 v1 uplinks (multi-mode sensor node).
package lsn50v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lsn50", "v1", decoders.New(Decode,
		decoders.Offer("bat_v", "V"),
		decoders.Offer("temp_c1", "C"),
		decoders.Offer("adc_ch0v", "V"),
		decoders.Offer("adc_ch1v", "V"),
		decoders.Offer("adc_ch4v", "V"),
		decoders.Offer("temp_c_sht", "C"),
		decoders.Offer("hum_sht", "%"),
		decoders.Offer("distance", "cm"),
		decoders.Offer("temp_c2", "C"),
		decoders.Offer("temp_c3", "C"),
		decoders.Offer("weight", "g"),
	))
}

type Data struct {
	WorkMode       string   `json:"work_mode,omitempty"`
	BatV           float64  `json:"bat_v"`
	TempC1         *float64 `json:"temp_c1,omitempty"`
	ADCCH0V        *float64 `json:"adc_ch0v,omitempty"`
	ADCCH1V        *float64 `json:"adc_ch1v,omitempty"`
	ADCCH4V        *float64 `json:"adc_ch4v,omitempty"`
	DigitalIStatus string   `json:"digital_i_status,omitempty"`
	EXTITrigger    string   `json:"exti_trigger,omitempty"`
	DoorStatus     string   `json:"door_status,omitempty"`
	TempCSHT       *float64 `json:"temp_c_sht,omitempty"`
	HumSHT         *float64 `json:"hum_sht,omitempty"`
	Distance       *float64 `json:"distance,omitempty"`
	TempC2         *float64 `json:"temp_c2,omitempty"`
	TempC3         *float64 `json:"temp_c3,omitempty"`
	Weight         *int     `json:"weight,omitempty"`
}

func (d *Data) Measurements() []decoders.Measurement {
	measurements := []decoders.Measurement{
		decoders.Float("bat_v", "V", d.BatV),
	}
	measurements = decoders.AppendFloat(measurements, "temp_c1", "C", d.TempC1)
	measurements = decoders.AppendFloat(measurements, "adc_ch0v", "V", d.ADCCH0V)
	measurements = decoders.AppendFloat(measurements, "adc_ch1v", "V", d.ADCCH1V)
	measurements = decoders.AppendFloat(measurements, "adc_ch4v", "V", d.ADCCH4V)
	measurements = decoders.AppendFloat(measurements, "temp_c_sht", "C", d.TempCSHT)
	measurements = decoders.AppendFloat(measurements, "hum_sht", "%", d.HumSHT)
	measurements = decoders.AppendFloat(measurements, "distance", "cm", d.Distance)
	measurements = decoders.AppendFloat(measurements, "temp_c2", "C", d.TempC2)
	measurements = decoders.AppendFloat(measurements, "temp_c3", "C", d.TempC3)
	measurements = decoders.AppendInt(measurements, "weight", "g", d.Weight)
	return measurements
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
		return nil, fmt.Errorf("lsn50v1: payload too short: %d bytes (want >= 11)", len(b))
	}
	mode := (b[6] & 0x7C) >> 2
	if mode == 2 && len(b) < 12 {
		return nil, fmt.Errorf("lsn50v1: payload too short: %d bytes (want >= 12 for mode 2)", len(b))
	}
	d := &Data{}

	switch mode {
	case 2:
		d.WorkMode = "3ADC"
		d.BatV = float64(b[11]) / 10
		d.ADCCH0V = ptr(float64(uint16(b[0])<<8|uint16(b[1])) / 1000)
		d.ADCCH1V = ptr(float64(uint16(b[2])<<8|uint16(b[3])) / 1000)
		d.ADCCH4V = ptr(float64(uint16(b[4])<<8|uint16(b[5])) / 1000)
	default:
		d.BatV = float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		d.TempC1 = ptr(round(signed16(b[2], b[3])/10, 2))
		d.ADCCH0V = ptr(float64(uint16(b[4])<<8|uint16(b[5])) / 1000)
	}

	if b[6]&0x02 != 0 {
		d.DigitalIStatus = "H"
	} else {
		d.DigitalIStatus = "L"
	}

	switch mode {
	case 0:
		d.WorkMode = "IIC"
		if b[6]&0x01 != 0 {
			d.EXTITrigger = "TRUE"
		} else {
			d.EXTITrigger = "FALSE"
		}
		if b[6]&0x80 != 0 {
			d.DoorStatus = "CLOSE"
		} else {
			d.DoorStatus = "OPEN"
		}
		d.TempCSHT = ptr(round(signed16(b[7], b[8])/10, 2))
		d.HumSHT = ptr(round(float64(uint16(b[9])<<8|uint16(b[10]))/10, 1))
	case 1:
		d.WorkMode = "Distance"
		if b[6]&0x01 != 0 {
			d.EXTITrigger = "TRUE"
		} else {
			d.EXTITrigger = "FALSE"
		}
		if b[6]&0x80 != 0 {
			d.DoorStatus = "CLOSE"
		} else {
			d.DoorStatus = "OPEN"
		}
		d.Distance = ptr(round(float64(uint16(b[7])<<8|uint16(b[8]))/10, 1))
	case 2:
		if b[6]&0x01 != 0 {
			d.EXTITrigger = "TRUE"
		} else {
			d.EXTITrigger = "FALSE"
		}
		if b[6]&0x80 != 0 {
			d.DoorStatus = "CLOSE"
		} else {
			d.DoorStatus = "OPEN"
		}
		if uint16(b[9])<<8|uint16(b[10]) == 0 {
			illum := int(int16(uint16(b[7])<<8 | uint16(b[8])))
			_ = illum
		} else {
			d.TempCSHT = ptr(round(signed16(b[7], b[8])/10, 2))
			d.HumSHT = ptr(round(float64(uint16(b[9])<<8|uint16(b[10]))/10, 1))
		}
	case 3:
		d.WorkMode = "3DS18B20"
		if b[6]&0x01 != 0 {
			d.EXTITrigger = "TRUE"
		} else {
			d.EXTITrigger = "FALSE"
		}
		if b[6]&0x80 != 0 {
			d.DoorStatus = "CLOSE"
		} else {
			d.DoorStatus = "OPEN"
		}
		d.TempC2 = ptr(round(signed16(b[7], b[8])/10, 2))
		d.TempC3 = ptr(round(signed16(b[9], b[10])/10, 1))
	case 4:
		d.WorkMode = "Weight"
		if b[6]&0x01 != 0 {
			d.EXTITrigger = "TRUE"
		} else {
			d.EXTITrigger = "FALSE"
		}
		if b[6]&0x80 != 0 {
			d.DoorStatus = "CLOSE"
		} else {
			d.DoorStatus = "OPEN"
		}
		w := int(int16(uint16(b[7])<<8 | uint16(b[8])))
		d.Weight = ptr(w)
	}
	return d, nil
}
