// Package lht65v1 decodes Dragino LHT65 v1 uplinks.
// Registers as dragino/lht65/v1.
package lht65v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lht65", "v1", decoders.New(Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.Temperature, decoders.Celsius),
		decoders.Offer(decoders.Humidity, decoders.Percent),
		decoders.Offer(decoders.ExternalTemp, decoders.Celsius),
		decoders.Offer(decoders.Illumination, decoders.Lux),
		decoders.Offer(decoders.ADCVoltage, decoders.Volt),
		decoders.Offer(decoders.InterruptCount, decoders.Count),
	))
}

var sensorTypes = map[byte]string{
	0: "No external sensor",
	1: "Temperature Sensor",
	4: "Interrupt/Door Sensor send",
	5: "Illumination Sensor",
	6: "ADC Sensor",
	7: "Interrupt Sensor count",
}

var batLevels = [4]string{"Ultra Low", "Low", "OK", "Good"}

// Data is the decoded LHT65 uplink.
type Data struct {
	BatLevel               string   `json:"bat_level"`
	SensorType             string   `json:"sensor_type"`
	BatteryVoltage         float64  `json:"battery_voltage"`
	Temperature            float64  `json:"temperature"`
	Humidity               float64  `json:"humidity"`
	ExternalTemperature    *float64 `json:"external_temperature,omitempty"`
	InterruptPinLevel      *string  `json:"interrupt_pin_level,omitempty"`
	InterruptStatus        *string  `json:"interrupt_status,omitempty"`
	Illumination           *int     `json:"illumination,omitempty"`
	ADCVoltage             *float64 `json:"adc_voltage,omitempty"`
	InterruptCount         *int     `json:"interrupt_count,omitempty"`
	SensorConnectionStatus *string  `json:"sensor_connection_status,omitempty"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

// Measurements returns the numeric readings decoded from this uplink.
func (d *Data) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatteryVoltage),
		decoders.Float(decoders.Temperature, decoders.Celsius, d.Temperature),
		decoders.Float(decoders.Humidity, decoders.Percent, d.Humidity),
	}
	ms = decoders.AppendFloat(ms, decoders.ExternalTemp, decoders.Celsius, d.ExternalTemperature)
	ms = decoders.AppendInt(ms, decoders.Illumination, decoders.Lux, d.Illumination)
	ms = decoders.AppendFloat(ms, decoders.ADCVoltage, decoders.Volt, d.ADCVoltage)
	ms = decoders.AppendInt(ms, decoders.InterruptCount, decoders.Count, d.InterruptCount)
	if d.SensorConnectionStatus != nil {
		markNoConnection(ms, map[string]bool{
			decoders.ExternalTemp:   true,
			decoders.Illumination:   true,
			decoders.ADCVoltage:     true,
			decoders.InterruptCount: true,
		})
	}
	return ms
}

func markNoConnection(ms []decoders.Measurement, names map[string]bool) {
	for i := range ms {
		if names[ms[i].Name] {
			ms[i].Valid = false
			ms[i].Quality = decoders.QualityNoConnection
		}
	}
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

// Decode decodes an LHT65 v1 uplink payload.
func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 9 {
		return nil, fmt.Errorf("lht65v1: payload too short: %d bytes (want >= 9)", len(b))
	}

	sensorType := b[6] & 0x7F

	name, ok := sensorTypes[sensorType]
	if !ok {
		name = "Unknown sensor type"
	}

	batteryRaw := uint16(b[0])<<8 | uint16(b[1])
	d := &Data{
		BatLevel:       batLevels[batteryRaw>>14],
		SensorType:     name,
		BatteryVoltage: float64(batteryRaw&0x3FFF) / 1000,
		Temperature:    round(float64(int16(uint16(b[2])<<8|uint16(b[3])))/100, 2),
		Humidity:       round(float64(uint16(b[4])<<8|uint16(b[5]))/10, 1),
	}

	raw := uint16(b[7])<<8 | uint16(b[8])

	switch sensorType {
	case 1:
		d.ExternalTemperature = ptr(round(float64(int16(raw))/100, 2))
	case 4:
		lvl := "Low"
		if b[7] != 0 {
			lvl = "High"
		}
		st := "False"
		if b[8] != 0 {
			st = "True"
		}
		d.InterruptPinLevel = ptr(lvl)
		d.InterruptStatus = ptr(st)
	case 5:
		d.Illumination = ptr(int(raw))
	case 6:
		d.ADCVoltage = ptr(float64(raw) / 1000)
	case 7:
		d.InterruptCount = ptr(int(raw))
	}

	if b[6]&0x80 != 0 {
		d.SensorConnectionStatus = ptr("Sensor no connection")
	}

	return d, nil
}
