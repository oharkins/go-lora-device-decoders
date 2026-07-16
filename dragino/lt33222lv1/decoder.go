// Package lt33222lv1 decodes Dragino LT33222-L v1 uplinks (I/O controller).
// The byte-level format is identical to the LT22222-L; hardware mode is
// reported inside the payload (byte 10 bits 7-6).
package lt33222lv1

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("dragino", "lt33222-l", "v1", decoders.New(
		Decode,
		decoders.Offer("count1_times", "count"),
		decoders.Offer("count2_times", "count"),
		decoders.Offer("acount_times", "count"),
		decoders.Offer("avi1_v", "V"),
		decoders.Offer("avi2_v", "V"),
		decoders.Offer("aci1_ma", "mA"),
		decoders.Offer("aci2_ma", "mA"),
	))
}

type Data struct {
	HardwareMode string   `json:"hardware_mode,omitempty"`
	WorkMode     string   `json:"work_mode,omitempty"`
	DO1Status    string   `json:"do1_status,omitempty"`
	DO2Status    string   `json:"do2_status,omitempty"`
	DO3Status    string   `json:"do3_status,omitempty"`
	RO1Status    string   `json:"ro1_status,omitempty"`
	RO2Status    string   `json:"ro2_status,omitempty"`
	DI1Status    string   `json:"di1_status,omitempty"`
	DI2Status    string   `json:"di2_status,omitempty"`
	DI3Status    string   `json:"di3_status,omitempty"`
	FirstStatus  string   `json:"first_status,omitempty"`
	Count1Times  *int     `json:"count1_times,omitempty"`
	Count2Times  *int     `json:"count2_times,omitempty"`
	AcountTimes  *int     `json:"acount_times,omitempty"`
	AVI1V        *float64 `json:"avi1_v,omitempty"`
	AVI2V        *float64 `json:"avi2_v,omitempty"`
	ACI1MA       *float64 `json:"aci1_ma,omitempty"`
	ACI2MA       *float64 `json:"aci2_ma,omitempty"`
	ModeStatus   string   `json:"mode_status,omitempty"`
	AV1LFlag     string   `json:"av1l_flag,omitempty"`
	AV1HFlag     string   `json:"av1h_flag,omitempty"`
	AV2LFlag     string   `json:"av2l_flag,omitempty"`
	AV2HFlag     string   `json:"av2h_flag,omitempty"`
	AC1LFlag     string   `json:"ac1l_flag,omitempty"`
	AC1HFlag     string   `json:"ac1h_flag,omitempty"`
	AC2LFlag     string   `json:"ac2l_flag,omitempty"`
	AC2HFlag     string   `json:"ac2h_flag,omitempty"`
	AV1LStatus   string   `json:"av1l_status,omitempty"`
	AV1HStatus   string   `json:"av1h_status,omitempty"`
	AV2LStatus   string   `json:"av2l_status,omitempty"`
	AV2HStatus   string   `json:"av2h_status,omitempty"`
	AC1LStatus   string   `json:"ac1l_status,omitempty"`
	AC1HStatus   string   `json:"ac1h_status,omitempty"`
	AC2LStatus   string   `json:"ac2l_status,omitempty"`
	AC2HStatus   string   `json:"ac2h_status,omitempty"`
	DI1Flag      string   `json:"di1_flag,omitempty"`
	DI2Flag      string   `json:"di2_flag,omitempty"`
}

func (d *Data) Measurements() []decoders.Measurement {
	var ms []decoders.Measurement
	ms = decoders.AppendInt(ms, "count1_times", "count", d.Count1Times)
	ms = decoders.AppendInt(ms, "count2_times", "count", d.Count2Times)
	ms = decoders.AppendInt(ms, "acount_times", "count", d.AcountTimes)
	ms = decoders.AppendFloat(ms, "avi1_v", "V", d.AVI1V)
	ms = decoders.AppendFloat(ms, "avi2_v", "V", d.AVI2V)
	ms = decoders.AppendFloat(ms, "aci1_ma", "mA", d.ACI1MA)
	ms = decoders.AppendFloat(ms, "aci2_ma", "mA", d.ACI2MA)
	return ms
}

func ptr[T any](v T) *T { return &v }

func round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

func flag(b byte, bit byte) string {
	if b&bit != 0 {
		return "True"
	}
	return "False"
}

func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) != 11 {
		return nil, fmt.Errorf("lt33222lv1: want 11 bytes, got %d", len(b))
	}
	hardware := (b[10] & 0xC0) >> 6
	mode0 := b[10] & 0xFF
	mode := b[10] & 0x3F
	d := &Data{}

	switch hardware {
	case 0:
		d.HardwareMode = "LT33222"
		if b[8]&0x04 != 0 {
			d.DO3Status = "L"
		} else {
			d.DO3Status = "H"
		}
		if mode0 == 1 {
			if b[8]&0x20 != 0 {
				d.DI3Status = "H"
			} else {
				d.DI3Status = "L"
			}
		}
	case 1:
		d.HardwareMode = "LT22222"
	}

	if mode != 6 {
		if b[8]&0x01 != 0 {
			d.DO1Status = "L"
		} else {
			d.DO1Status = "H"
		}
		if b[8]&0x02 != 0 {
			d.DO2Status = "L"
		} else {
			d.DO2Status = "H"
		}
		if b[8]&0x80 != 0 {
			d.RO1Status = "ON"
		} else {
			d.RO1Status = "OFF"
		}
		if b[8]&0x40 != 0 {
			d.RO2Status = "ON"
		} else {
			d.RO2Status = "OFF"
		}
		if mode != 1 {
			if mode != 5 {
				c := int(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
				d.Count1Times = ptr(c)
			}
			if b[8]&0x20 != 0 {
				d.FirstStatus = "Yes"
			} else {
				d.FirstStatus = "No"
			}
		}
	}

	switch mode {
	case 1:
		d.WorkMode = "2ACI+2AVI"
		d.AVI1V = ptr(round(float64(int16(uint16(b[0])<<8|uint16(b[1])))/1000, 3))
		d.AVI2V = ptr(round(float64(int16(uint16(b[2])<<8|uint16(b[3])))/1000, 3))
		d.ACI1MA = ptr(round(float64(int16(uint16(b[4])<<8|uint16(b[5])))/1000, 3))
		d.ACI2MA = ptr(round(float64(int16(uint16(b[6])<<8|uint16(b[7])))/1000, 3))
		if b[8]&0x08 != 0 {
			d.DI1Status = "H"
		} else {
			d.DI1Status = "L"
		}
		if b[8]&0x10 != 0 {
			d.DI2Status = "H"
		} else {
			d.DI2Status = "L"
		}
	case 2:
		d.WorkMode = "Count mode 1"
		c := int(uint32(b[4])<<24 | uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7]))
		d.Count2Times = ptr(c)
	case 3:
		d.WorkMode = "2ACI+1Count"
		d.ACI1MA = ptr(round(float64(int16(uint16(b[4])<<8|uint16(b[5])))/1000, 3))
		d.ACI2MA = ptr(round(float64(int16(uint16(b[6])<<8|uint16(b[7])))/1000, 3))
	case 4:
		d.WorkMode = "Count mode 2"
		c := int(uint32(b[4])<<24 | uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7]))
		d.AcountTimes = ptr(c)
	case 5:
		d.WorkMode = "1ACI+2AVI+1Count"
		d.AVI1V = ptr(round(float64(int16(uint16(b[0])<<8|uint16(b[1])))/1000, 3))
		d.AVI2V = ptr(round(float64(int16(uint16(b[2])<<8|uint16(b[3])))/1000, 3))
		d.ACI1MA = ptr(round(float64(int16(uint16(b[4])<<8|uint16(b[5])))/1000, 3))
		c := int(uint16(b[6])<<8 | uint16(b[7]))
		d.Count1Times = ptr(c)
	case 6:
		d.WorkMode = "Exit mode"
		if b[9] != 0 {
			d.ModeStatus = "True"
		} else {
			d.ModeStatus = "False"
		}
		d.AV1LFlag = flag(b[0], 0x80)
		d.AV1HFlag = flag(b[0], 0x40)
		d.AV2LFlag = flag(b[0], 0x20)
		d.AV2HFlag = flag(b[0], 0x10)
		d.AC1LFlag = flag(b[0], 0x08)
		d.AC1HFlag = flag(b[0], 0x04)
		d.AC2LFlag = flag(b[0], 0x02)
		d.AC2HFlag = flag(b[0], 0x01)
		d.AV1LStatus = flag(b[1], 0x80)
		d.AV1HStatus = flag(b[1], 0x40)
		d.AV2LStatus = flag(b[1], 0x20)
		d.AV2HStatus = flag(b[1], 0x10)
		d.AC1LStatus = flag(b[1], 0x08)
		d.AC1HStatus = flag(b[1], 0x04)
		d.AC2LStatus = flag(b[1], 0x02)
		d.AC2HStatus = flag(b[1], 0x01)
		d.DI2Status = flag(b[2], 0x08)
		d.DI2Flag = flag(b[2], 0x04)
		d.DI1Status = flag(b[2], 0x02)
		d.DI1Flag = flag(b[2], 0x01)
	}
	return d, nil
}
