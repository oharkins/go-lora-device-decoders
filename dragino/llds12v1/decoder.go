// Package llds12v1 decodes Dragino LLDS12 v1 uplinks.
package llds12v1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "llds12", "v1", decoders.New(Decode,
		decoders.Offer(decoders.BatteryVoltage, decoders.Volt),
		decoders.Offer(decoders.DSTemperature, decoders.Celsius),
		decoders.Offer(decoders.DistanceCM, decoders.Centimeter),
		decoders.Offer(decoders.DistanceSignalStrength, ""),
		decoders.Offer(decoders.LidarTemperature, decoders.Celsius),
		decoders.Offer(decoders.InterruptFlag, ""),
		decoders.Offer(decoders.MessageType, ""),
	))
}

type Data struct {
	BatV                float64 `json:"bat_v"`
	TempCDS18B20        float64 `json:"temp_c_ds18b20"`
	LidarDistanceCM     float64 `json:"lidar_distance_cm"`
	LidarSignalStrength int     `json:"lidar_signal_strength"`
	LidarTemp           int     `json:"lidar_temp"`
	InterruptFlag       int     `json:"interrupt_flag"`
	MessageType         int     `json:"message_type"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float(decoders.BatteryVoltage, decoders.Volt, d.BatV),
		decoders.Float(decoders.DSTemperature, decoders.Celsius, d.TempCDS18B20),
		decoders.Float(decoders.DistanceCM, decoders.Centimeter, d.LidarDistanceCM),
		decoders.Int(decoders.DistanceSignalStrength, "", d.LidarSignalStrength),
		decoders.Int(decoders.LidarTemperature, decoders.Celsius, d.LidarTemp),
		decoders.Int(decoders.InterruptFlag, "", d.InterruptFlag),
		decoders.Int(decoders.MessageType, "", d.MessageType),
	}
}

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 11 {
		return nil, fmt.Errorf("llds12v1: payload too short: %d bytes (want >= 11)", len(b))
	}
	if b[0] == 0x03 && b[10] == 0x02 {
		return nil, decoders.ErrIgnored
	}
	return &Data{
		BatV:                float64((uint16(b[0])<<8|uint16(b[1]))&0x3FFF) / 1000,
		TempCDS18B20:        round(float64(int16(uint16(b[2])<<8|uint16(b[3])))/10, 2),
		LidarDistanceCM:     float64(uint16(b[4])<<8|uint16(b[5])) / 10,
		LidarSignalStrength: int(uint16(b[6])<<8 | uint16(b[7])),
		LidarTemp:           int(int8(b[9])),
		InterruptFlag:       int(b[8]),
		MessageType:         int(b[10]),
	}, nil
}
