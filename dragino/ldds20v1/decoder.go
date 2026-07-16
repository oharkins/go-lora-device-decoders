// Package ldds20v1 decodes Dragino LDDS20 v1 uplinks.
package ldds20v1

import (
	"fmt"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "ldds20", "v1", decoders.New(Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.DistanceMM, decoders.Millimeter),
		decoders.Offer(decoders.InterruptStatus, ""),
	))
}

type Data struct {
	BatV            float64 `json:"bat_v"`
	DistanceMM      *int    `json:"distance_mm,omitempty"`
	DistanceStatus  string  `json:"distance_status,omitempty"`
	InterruptStatus int     `json:"interrupt_status"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	measurements := []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
	}
	switch d.DistanceStatus {
	case "Invalid Reading":
		raw := 0
		if d.DistanceMM != nil {
			raw = *d.DistanceMM
		}
		measurements = append(measurements, decoders.FloatQuality(decoders.DistanceMM, decoders.Millimeter, float64(raw), false, decoders.QualityInvalid))
	case "No Sensor":
		measurements = append(measurements, decoders.FloatQuality(decoders.DistanceMM, decoders.Millimeter, 0, false, decoders.QualityNoSensor))
	default:
		measurements = decoders.AppendInt(measurements, decoders.DistanceMM, decoders.Millimeter, d.DistanceMM)
	}
	measurements = append(measurements, decoders.Int(decoders.InterruptStatus, "", d.InterruptStatus))
	return measurements
}

func ptr[T any](v T) *T { return &v }

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 2 {
		return nil, fmt.Errorf("ldds20v1: payload too short: %d bytes", len(b))
	}
	d := &Data{
		BatV:            float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		InterruptStatus: int(b[len(b)-1]),
	}
	if len(b) == 5 {
		v := int(uint16(b[2])<<8 | uint16(b[3]))
		if v < 20 {
			d.DistanceStatus = "Invalid Reading"
			d.DistanceMM = ptr(v)
		} else {
			d.DistanceMM = ptr(v)
		}
	} else {
		d.DistanceStatus = "No Sensor"
	}
	return d, nil
}
