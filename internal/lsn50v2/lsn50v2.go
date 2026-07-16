// Package lsn50v2 decodes the shared Dragino LSN50v2-D20/S31 uplink layout.
package lsn50v2

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

// Data is the shared decoded LSN50v2-D20/S31 uplink.
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

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	measurements := []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
	}
	measurements = decoders.AppendFloat(measurements, decoders.TemperatureProbe, decoders.Celsius, d.TempC1)
	measurements = decoders.AppendFloat(measurements, decoders.ADCCH0Voltage, decoders.Volt, d.ADCCH0V)
	measurements = decoders.AppendFloat(measurements, decoders.ADCCH1Voltage, decoders.Volt, d.ADCCH1V)
	measurements = decoders.AppendFloat(measurements, decoders.ADCCH4Voltage, decoders.Volt, d.ADCCH4V)
	measurements = decoders.AppendFloat(measurements, decoders.Temperature, decoders.Celsius, d.TempCSHT)
	measurements = decoders.AppendFloat(measurements, decoders.Humidity, decoders.Percent, d.HumSHT)
	measurements = decoders.AppendInt(measurements, decoders.Illumination, decoders.Lux, d.Illum)
	measurements = decoders.AppendFloat(measurements, decoders.DistanceCM, decoders.Centimeter, d.DistanceCM)
	measurements = decoders.AppendFloat(measurements, decoders.DistanceSignalStrength, "", d.DistanceSignal)
	measurements = decoders.AppendFloat(measurements, decoders.Temperature2, decoders.Celsius, d.TempC2)
	measurements = decoders.AppendFloat(measurements, decoders.Temperature3, decoders.Celsius, d.TempC3)
	measurements = decoders.AppendInt(measurements, decoders.Weight, decoders.Gram, d.Weight)
	measurements = decoders.AppendInt(measurements, decoders.EventCount, decoders.Count, d.Count)
	if d.TempC1Min != nil {
		measurements = append(measurements, decoders.Int(decoders.TemperatureProbeMin, decoders.Celsius, int(*d.TempC1Min)))
	}
	if d.TempC1Max != nil {
		measurements = append(measurements, decoders.Int(decoders.TemperatureProbeMax, decoders.Celsius, int(*d.TempC1Max)))
	}
	if d.SHTEmpMin != nil {
		measurements = append(measurements, decoders.Int(decoders.SHTTemperatureMin, decoders.Celsius, int(*d.SHTEmpMin)))
	}
	if d.SHTEmpMax != nil {
		measurements = append(measurements, decoders.Int(decoders.SHTTemperatureMax, decoders.Celsius, int(*d.SHTEmpMax)))
	}
	if d.SHTHumMin != nil {
		measurements = append(measurements, decoders.Int(decoders.SHTHumidityMin, decoders.Percent, int(*d.SHTHumMin)))
	}
	if d.SHTHumMax != nil {
		measurements = append(measurements, decoders.Int(decoders.SHTHumidityMax, decoders.Percent, int(*d.SHTHumMax)))
	}
	return measurements
}

// Offers are the measurements LSN50v2-D20/S31 devices can produce.
func Offers() []decoders.Offering {
	return []decoders.Offering{
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.TemperatureProbe, decoders.Celsius),
		decoders.Offer(decoders.ADCCH0Voltage, decoders.Volt),
		decoders.Offer(decoders.ADCCH1Voltage, decoders.Volt),
		decoders.Offer(decoders.ADCCH4Voltage, decoders.Volt),
		decoders.Offer(decoders.Temperature, decoders.Celsius),
		decoders.Offer(decoders.Humidity, decoders.Percent),
		decoders.Offer(decoders.Illumination, decoders.Lux),
		decoders.Offer(decoders.DistanceCM, decoders.Centimeter),
		decoders.Offer(decoders.DistanceSignalStrength, ""),
		decoders.Offer(decoders.Temperature2, decoders.Celsius),
		decoders.Offer(decoders.Temperature3, decoders.Celsius),
		decoders.Offer(decoders.Weight, decoders.Gram),
		decoders.Offer(decoders.EventCount, decoders.Count),
		decoders.Offer(decoders.TemperatureProbeMin, decoders.Celsius),
		decoders.Offer(decoders.TemperatureProbeMax, decoders.Celsius),
		decoders.Offer(decoders.SHTTemperatureMin, decoders.Celsius),
		decoders.Offer(decoders.SHTTemperatureMax, decoders.Celsius),
		decoders.Offer(decoders.SHTHumidityMin, decoders.Percent),
		decoders.Offer(decoders.SHTHumidityMax, decoders.Percent),
	}
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func signed16(hi, lo byte) float64 {
	return float64(int16(uint16(hi)<<8 | uint16(lo)))
}

// Decode parses a shared LSN50v2-D20/S31 payload. name is used in error messages.
func Decode(name string, u decoders.Uplink) (*Data, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("%s: payload too short: %d bytes", name, len(b))
	}
	mode := (b[6] & 0x7C) >> 2
	if mode == 2 && len(b) < 12 {
		return nil, fmt.Errorf("%s: payload too short: %d bytes (want >= 12 for mode 2)", name, len(b))
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
		iic9 := uint16(b[9])<<8 | uint16(b[10])
		if iic9 == 0 {
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
			s := round(float64(sig), 0)
			d.DistanceSignal = ptr(s)
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
		iic9 := uint16(b[9])<<8 | uint16(b[10])
		if iic9 == 0 {
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
		w := int(int16(uint16(b[7])<<8 | uint16(b[8])))
		d.Weight = ptr(w)
	case 5:
		d.WorkMode = "Count"
		c := int(uint32(b[7])<<24 | uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10]))
		d.Count = ptr(c)
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
