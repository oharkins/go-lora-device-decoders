// Package lsn50v2s31v1 decodes Dragino LSN50v2-S31 v1 uplinks.
// Byte layout is identical to LSN50v2-D20.
package lsn50v2s31v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lsn50v2-s31", "v1", decoders.New(Decode,
		decoders.Offer("bat_v", "V"),
		decoders.Offer("temp_c1", "C"),
		decoders.Offer("adc_ch0v", "V"),
		decoders.Offer("adc_ch1v", "V"),
		decoders.Offer("adc_ch4v", "V"),
		decoders.Offer("temp_c_sht", "C"),
		decoders.Offer("hum_sht", "%"),
		decoders.Offer("illum", "lux"),
		decoders.Offer("distance_cm", "cm"),
		decoders.Offer("distance_signal_strength", ""),
		decoders.Offer("temp_c2", "C"),
		decoders.Offer("temp_c3", "C"),
		decoders.Offer("weight", "g"),
		decoders.Offer("count", ""),
		decoders.Offer("temp_c1_min", "C"),
		decoders.Offer("temp_c1_max", "C"),
		decoders.Offer("sht_temp_min", "C"),
		decoders.Offer("sht_temp_max", "C"),
		decoders.Offer("sht_hum_min", "%"),
		decoders.Offer("sht_hum_max", "%"),
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
	Illum          *int     `json:"illum,omitempty"`
	DistanceCM     *float64 `json:"distance_cm,omitempty"`
	DistanceSignal *float64 `json:"distance_signal_strength,omitempty"`
	TempC2         *float64 `json:"temp_c2,omitempty"`
	TempC3         *float64 `json:"temp_c3,omitempty"`
	Weight         *int     `json:"weight,omitempty"`
	Count          *int     `json:"count,omitempty"`
	TempC1Min      *int8    `json:"temp_c1_min,omitempty"`
	TempC1Max      *int8    `json:"temp_c1_max,omitempty"`
	SHTEmpMin      *int8    `json:"sht_temp_min,omitempty"`
	SHTEmpMax      *int8    `json:"sht_temp_max,omitempty"`
	SHTHumMin      *uint8   `json:"sht_hum_min,omitempty"`
	SHTHumMax      *uint8   `json:"sht_hum_max,omitempty"`
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
	measurements = decoders.AppendInt(measurements, "illum", "lux", d.Illum)
	measurements = decoders.AppendFloat(measurements, "distance_cm", "cm", d.DistanceCM)
	measurements = decoders.AppendFloat(measurements, "distance_signal_strength", "", d.DistanceSignal)
	measurements = decoders.AppendFloat(measurements, "temp_c2", "C", d.TempC2)
	measurements = decoders.AppendFloat(measurements, "temp_c3", "C", d.TempC3)
	measurements = decoders.AppendInt(measurements, "weight", "g", d.Weight)
	measurements = decoders.AppendInt(measurements, "count", "", d.Count)
	if d.TempC1Min != nil {
		measurements = append(measurements, decoders.Int("temp_c1_min", "C", int(*d.TempC1Min)))
	}
	if d.TempC1Max != nil {
		measurements = append(measurements, decoders.Int("temp_c1_max", "C", int(*d.TempC1Max)))
	}
	if d.SHTEmpMin != nil {
		measurements = append(measurements, decoders.Int("sht_temp_min", "C", int(*d.SHTEmpMin)))
	}
	if d.SHTEmpMax != nil {
		measurements = append(measurements, decoders.Int("sht_temp_max", "C", int(*d.SHTEmpMax)))
	}
	if d.SHTHumMin != nil {
		measurements = append(measurements, decoders.Int("sht_hum_min", "%", int(*d.SHTHumMin)))
	}
	if d.SHTHumMax != nil {
		measurements = append(measurements, decoders.Int("sht_hum_max", "%", int(*d.SHTHumMax)))
	}
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
		return nil, fmt.Errorf("lsn50v2s31v1: payload too short: %d bytes", len(b))
	}
	mode := (b[6] & 0x7C) >> 2
	if mode == 2 && len(b) < 12 {
		return nil, fmt.Errorf("lsn50v2s31v1: payload too short: %d bytes (want >= 12 for mode 2)", len(b))
	}
	d := &Data{}

	if mode != 2 && mode != 31 {
		d.BatV = float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		d.TempC1 = ptr(round(signed16(b[2], b[3])/10, 2))
		d.ADCCH0V = ptr(float64(uint16(b[4])<<8|uint16(b[5])) / 1000)
		if b[6]&0x02 != 0 {
			d.DigitalIStatus = "H"
		} else {
			d.DigitalIStatus = "L"
		}
		if mode != 6 {
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
		}
	}

	switch mode {
	case 0:
		d.WorkMode = "IIC"
		if uint16(b[9])<<8|uint16(b[10]) == 0 {
			v := int(int16(uint16(b[7])<<8 | uint16(b[8])))
			d.Illum = ptr(v)
		} else {
			d.TempCSHT = ptr(round(signed16(b[7], b[8])/10, 2))
			d.HumSHT = ptr(round(float64(uint16(b[9])<<8|uint16(b[10]))/10, 1))
		}
	case 1:
		d.WorkMode = "Distance"
		d.DistanceCM = ptr(round(float64(uint16(b[7])<<8|uint16(b[8]))/10, 1))
		sig := uint16(b[9])<<8 | uint16(b[10])
		if sig != 0xFFFF {
			d.DistanceSignal = ptr(round(float64(sig), 0))
		}
	case 2:
		d.WorkMode = "3ADC"
		d.BatV = float64(b[11]) / 10
		d.ADCCH0V = ptr(float64(uint16(b[0])<<8|uint16(b[1])) / 1000)
		d.ADCCH1V = ptr(float64(uint16(b[2])<<8|uint16(b[3])) / 1000)
		d.ADCCH4V = ptr(float64(uint16(b[4])<<8|uint16(b[5])) / 1000)
		if b[6]&0x02 != 0 {
			d.DigitalIStatus = "H"
		} else {
			d.DigitalIStatus = "L"
		}
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
			v := int(int16(uint16(b[7])<<8 | uint16(b[8])))
			d.Illum = ptr(v)
		} else {
			d.TempCSHT = ptr(round(signed16(b[7], b[8])/10, 2))
			d.HumSHT = ptr(round(float64(uint16(b[9])<<8|uint16(b[10]))/10, 1))
		}
	case 3:
		d.WorkMode = "3DS18B20"
		d.TempC2 = ptr(round(signed16(b[7], b[8])/10, 2))
		d.TempC3 = ptr(round(signed16(b[9], b[10])/10, 1))
	case 4:
		d.WorkMode = "Weight"
		d.Weight = ptr(int(int16(uint16(b[7])<<8 | uint16(b[8]))))
	case 5:
		d.WorkMode = "Count"
		d.Count = ptr(int(uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10])))
	case 31:
		d.WorkMode = "ALARM"
		d.BatV = float64(uint16(b[0])<<8|uint16(b[1])) / 1000
		d.TempC1 = ptr(round(signed16(b[2], b[3])/10, 2))
		d.TempC1Min = ptr(int8(b[4]))
		d.TempC1Max = ptr(int8(b[5]))
		d.SHTEmpMin = ptr(int8(b[7]))
		d.SHTEmpMax = ptr(int8(b[8]))
		d.SHTHumMin = ptr(b[9])
		d.SHTHumMax = ptr(b[10])
	}
	return d, nil
}
