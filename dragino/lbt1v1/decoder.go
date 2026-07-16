// Package lbt1v1 decodes Dragino LBT1 v1 uplinks (BLE beacon scanner).
// The payload bytes are ASCII characters; strings are built by treating each
// byte as a character code (equivalent to JS String.fromCharCode(byte)).
package lbt1v1

import (
	"encoding/hex"
	"fmt"
	"strconv"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lbt1", "v1", decoders.New(
		Decode,
		decoders.Offer("battery_voltage", "V"),
		decoders.Offer("major", ""),
		decoders.Offer("minor", ""),
		decoders.Offer("rssi", "dBm"),
		decoders.Offer("power", "dBm"),
		decoders.Offer("step_count", "count"),
		decoders.Offer("alarm", ""),
	))
}

type Data struct {
	UUID               string  `json:"uuid"`
	ADDR               string  `json:"addr"`
	Major              int     `json:"major"`
	Minor              int     `json:"minor"`
	RSSI               any     `json:"rssi"`
	Power              any     `json:"power"`
	DeviceInformation1 string  `json:"device_information1"`
	DeviceInformation2 string  `json:"device_information2"`
	DeviceInformation3 string  `json:"device_information3"`
	StepCount          int     `json:"step_count"`
	Alarm              int     `json:"alarm"`
	BatV               float64 `json:"bat_v"`
}

func (d *Data) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Float("battery_voltage", "V", d.BatV),
		decoders.Int("major", "", d.Major),
		decoders.Int("minor", "", d.Minor),
		decoders.Int("step_count", "count", d.StepCount),
		decoders.Int("alarm", "", d.Alarm),
	}
	switch v := d.RSSI.(type) {
	case int:
		ms = append(ms, decoders.Int("rssi", "dBm", v))
	case float64:
		ms = append(ms, decoders.Float("rssi", "dBm", v))
	}
	switch v := d.Power.(type) {
	case int:
		ms = append(ms, decoders.Int("power", "dBm", v))
	case float64:
		ms = append(ms, decoders.Float("power", "dBm", v))
	}
	return ms
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 6 {
		return nil, fmt.Errorf("lbt1v1: payload too short: %d bytes", len(b))
	}

	batV := float64(uint16(b[0])<<8|uint16(b[1])) / 1000
	mode := int(b[5])
	alarm := int((b[2] >> 4) & 0x0F)
	stepCount := int(uint32(b[2]&0x0F)<<16 | uint32(b[3])<<8 | uint32(b[4]))

	d := &Data{
		BatV:      batV,
		Alarm:     alarm,
		StepCount: stepCount,
	}

	switch mode {
	case 1:
		if len(b) < 11 {
			return nil, fmt.Errorf("lbt1v1 mode 1: need 11 bytes, got %d", len(b))
		}
		d.UUID = string(b[6:11])
	case 2:
		if len(b) < 50 {
			return nil, fmt.Errorf("lbt1v1 mode 2: need 50 bytes, got %d", len(b))
		}
		d.UUID = string(b[6:38])
		d.ADDR = string(b[38:50])
	case 3:
		if len(b) < 32 {
			return nil, fmt.Errorf("lbt1v1 mode 3: need 32 bytes, got %d", len(b))
		}
		d.UUID = string(b[6:18])
		major, err := strconv.ParseInt(string(b[18:22]), 16, 64)
		if err == nil {
			d.Major = int(major)
		}
		minor, err := strconv.ParseInt(string(b[22:26]), 16, 64)
		if err == nil {
			d.Minor = int(minor)
		}
		power, err := strconv.ParseInt(string(b[26:28]), 16, 64)
		if err == nil {
			d.Power = int(power)
		}
		rssi, err := strconv.ParseInt(string(b[28:32]), 10, 64)
		if err == nil {
			d.RSSI = int(rssi)
		}
	case 4:
		if len(b) < 48 {
			return nil, fmt.Errorf("lbt1v1 mode 4: need 48 bytes, got %d", len(b))
		}
		d.DeviceInformation1 = string(b[6:20])
		d.DeviceInformation2 = string(b[20:34])
		d.DeviceInformation3 = string(b[34:48])
	case 5:
		if len(b) < 30 {
			return nil, fmt.Errorf("lbt1v1 mode 5: need 30 bytes, got %d", len(b))
		}
		d.Major = int(uint16(b[22])<<8 | uint16(b[23]))
		d.Minor = int(uint16(b[24])<<8 | uint16(b[25]))
		d.Power = int(uint16(b[26])<<8 | uint16(b[27]))
		// JS bug: loop overwrites con, so rssi = last byte (b[29]) as hex
		d.RSSI = hex.EncodeToString([]byte{b[29]})
		d.UUID = hex.EncodeToString(b[6:22])
	}
	return d, nil
}
