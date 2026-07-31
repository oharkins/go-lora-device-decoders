// Package em300v1 decodes Milesight EM300 series uplinks.
//
// The whole series shares one channel/type TLV payload format (data fields are
// little-endian); the model only changes which channels appear, the meaning of
// channel 05 type 00 (water leak vs. digital input on EM300-DI), and the layout
// of historical-data records. One decoder is registered per model:
//
//	milesight/em300-th/v1   temperature & humidity
//	milesight/em300-mcs/v1  magnet switch
//	milesight/em300-sld/v1  spot leak detection
//	milesight/em300-zld/v1  zone leak detection
//	milesight/em300-mld/v1  membrane leak detection
//	milesight/em300-di/v1   pulse counter / digital input
//	milesight/em300-cl/v1   capacitive liquid level
//
// Basic-information frames (channel ff) decode to *DeviceInfo (KindDeviceInfo),
// historical-data frames (20ce / 21ce) to *DatalogData (KindDatalog), and data
// enquiry replies (fc6b / fc6c) return ErrIgnored.
package em300v1

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

// Model selects the EM300 variant, which affects channel semantics and
// historical-record layout.
type Model string

const (
	ModelTH  Model = "em300-th"
	ModelMCS Model = "em300-mcs"
	ModelSLD Model = "em300-sld"
	ModelZLD Model = "em300-zld"
	ModelMLD Model = "em300-mld"
	ModelDI  Model = "em300-di"
	ModelCL  Model = "em300-cl"
)

func init() {
	for _, m := range []Model{ModelTH, ModelMCS, ModelSLD, ModelZLD, ModelMLD, ModelDI, ModelCL} {
		decoders.Register("milesight", string(m), "v1", NewDecoder(m))
	}
}

// NewDecoder returns the decoder for one EM300 model.
func NewDecoder(m Model) decoders.Decoder {
	return decoders.New(func(u decoders.Uplink) (any, error) {
		return Decode(m, u)
	}, offersFor(m)...)
}

func offersFor(m Model) []decoders.Offering {
	tempHum := []decoders.Offering{
		decoders.Offer(decoders.Temperature, decoders.Celsius),
		decoders.Offer(decoders.Humidity, decoders.Percent),
	}
	offers := []decoders.Offering{decoders.Offer(decoders.BatteryPercent, decoders.Percent)}
	switch m {
	case ModelTH:
		offers = append(offers, tempHum...)
	case ModelMCS:
		offers = append(offers, tempHum...)
		offers = append(offers, decoders.Offer(decoders.DoorOpenStatus, ""))
	case ModelSLD, ModelZLD:
		offers = append(offers, tempHum...)
		offers = append(offers, decoders.Offer(decoders.WaterLeakStatus, ""))
	case ModelMLD:
		offers = append(offers, decoders.Offer(decoders.WaterLeakStatus, ""))
	case ModelDI:
		offers = append(offers, tempHum...)
		offers = append(offers,
			decoders.Offer(decoders.DigitalInput, ""),
			decoders.Offer(decoders.PulseCount, decoders.Count),
			decoders.Offer(decoders.WaterConsumption, ""),
			decoders.Offer(decoders.Alarm, ""),
		)
	case ModelCL:
		offers = append(offers,
			decoders.Offer(decoders.LiquidStatus, ""),
			decoders.Offer(decoders.CalibrationStatus, ""),
			decoders.Offer(decoders.Alarm, ""),
		)
	}
	return offers
}

// Data is a decoded EM300 telemetry uplink. Only channels present in the
// payload are set.
type Data struct {
	Battery          *int     `json:"battery,omitempty"`            // %
	Temperature      *float64 `json:"temperature,omitempty"`        // °C
	Humidity         *float64 `json:"humidity,omitempty"`           // %RH
	WaterLeak        *int     `json:"water_leak,omitempty"`         // 0 no leak, 1 leaked
	MagnetStatus     *int     `json:"magnet_status,omitempty"`      // 0 close, 1 open
	DigitalInput     *int     `json:"digital_input,omitempty"`      // 0 low, 1 high (EM300-DI)
	PulseCount       *int64   `json:"pulse_count,omitempty"`        // EM300-DI counter (firmware <= V1.2)
	WaterConv        *float64 `json:"water_conv,omitempty"`         // EM300-DI pulse conversion
	PulseConv        *float64 `json:"pulse_conv,omitempty"`         // EM300-DI pulse conversion
	WaterConsumption *float64 `json:"water_consumption,omitempty"`  // EM300-DI, unit per pulse conversion setting
	LiquidStatus     *int     `json:"liquid_status,omitempty"`      // 0 uncalibrated, 1 full, 2 empty, 255 sensor error (EM300-CL)
	Calibration      *int     `json:"calibration_status,omitempty"` // 0 failure, 1 success (EM300-CL)
	Alarm            *int     `json:"alarm,omitempty"`              // 1 alarm, 0 alarm dismiss
	AlarmType        *string  `json:"alarm_type,omitempty"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

// Measurements returns the numeric readings decoded from this uplink.
func (d *Data) Measurements() []decoders.Measurement {
	var ms []decoders.Measurement
	ms = decoders.AppendInt(ms, decoders.BatteryPercent, decoders.Percent, d.Battery)
	ms = decoders.AppendFloat(ms, decoders.Temperature, decoders.Celsius, d.Temperature)
	ms = decoders.AppendFloat(ms, decoders.Humidity, decoders.Percent, d.Humidity)
	ms = decoders.AppendInt(ms, decoders.WaterLeakStatus, "", d.WaterLeak)
	ms = decoders.AppendInt(ms, decoders.DoorOpenStatus, "", d.MagnetStatus)
	ms = decoders.AppendInt(ms, decoders.DigitalInput, "", d.DigitalInput)
	ms = decoders.AppendInt64(ms, decoders.PulseCount, decoders.Count, d.PulseCount)
	ms = decoders.AppendFloat(ms, decoders.WaterConsumption, "", d.WaterConsumption)
	if d.LiquidStatus != nil {
		if *d.LiquidStatus == 0xff {
			ms = append(ms, decoders.FloatQuality(decoders.LiquidStatus, "", float64(*d.LiquidStatus), false, decoders.QualityFault))
		} else {
			ms = append(ms, decoders.Int(decoders.LiquidStatus, "", *d.LiquidStatus))
		}
	}
	ms = decoders.AppendInt(ms, decoders.CalibrationStatus, "", d.Calibration)
	ms = decoders.AppendInt(ms, decoders.Alarm, "", d.Alarm)
	return ms
}

// DeviceInfo is the basic-information frame sent whenever the device joins
// the network (channel ff).
type DeviceInfo struct {
	PowerOn         bool   `json:"power_on,omitempty"`
	ProtocolVersion string `json:"protocol_version,omitempty"`
	HardwareVersion string `json:"hardware_version,omitempty"`
	SoftwareVersion string `json:"software_version,omitempty"`
	DeviceClass     string `json:"device_class,omitempty"`
	SerialNumber    string `json:"sn,omitempty"`
}

func (d *DeviceInfo) MessageKind() decoders.Kind           { return decoders.KindDeviceInfo }
func (d *DeviceInfo) Measurements() []decoders.Measurement { return nil }

// DatalogEntry is one historical-data record (channel 20ce, or 21ce for
// EM300-DI).
type DatalogEntry struct {
	Timestamp        int64    `json:"timestamp"` // unix seconds
	Temperature      *float64 `json:"temperature,omitempty"`
	Humidity         *float64 `json:"humidity,omitempty"`
	MagnetStatus     *int     `json:"magnet_status,omitempty"`
	WaterLeak        *int     `json:"water_leak,omitempty"`
	InterfaceType    *string  `json:"interface_type,omitempty"` // "digital" or "counter" (EM300-DI)
	AlarmType        *string  `json:"alarm_type,omitempty"`
	DigitalInput     *int     `json:"digital_input,omitempty"`
	PulseCount       *int64   `json:"pulse_count,omitempty"`
	WaterConv        *float64 `json:"water_conv,omitempty"`
	PulseConv        *float64 `json:"pulse_conv,omitempty"`
	WaterConsumption *float64 `json:"water_consumption,omitempty"`
}

// DatalogData is a historical-data uplink (retransmission or enquiry reply).
type DatalogData struct {
	Records []DatalogEntry `json:"records"`
}

func (d *DatalogData) MessageKind() decoders.Kind           { return decoders.KindDatalog }
func (d *DatalogData) Measurements() []decoders.Measurement { return nil }

var diAlarmTypes = map[byte]string{
	1: "water_outage_timeout_alarm",
	2: "water_outage_timeout_alarm_dismiss",
	3: "water_flow_timeout_alarm",
	4: "water_flow_timeout_alarm_dismiss",
	5: "digital_input_alarm",
	6: "digital_input_alarm_dismiss",
}

// Decode decodes an EM300 uplink for the given model. It returns *Data for
// telemetry, *DeviceInfo for basic-information frames, *DatalogData for
// historical data, or ErrIgnored for data enquiry replies.
func Decode(model Model, u decoders.Uplink) (any, error) {
	b := u.Payload
	if len(b) < 3 {
		return nil, fmt.Errorf("em300v1: payload too short: %d bytes (want >= 3)", len(b))
	}

	var (
		d       Data
		hasData bool
		info    DeviceInfo
		hasInfo bool
		records []DatalogEntry
		acked   bool
	)

	for i := 0; i < len(b); {
		if len(b)-i < 2 {
			return nil, fmt.Errorf("em300v1: trailing byte at offset %d", i)
		}
		ch, ty := b[i], b[i+1]
		i += 2

		n, err := dataLen(model, ch, ty)
		if err != nil {
			return nil, err
		}
		if len(b)-i < n {
			return nil, fmt.Errorf("em300v1: truncated channel %#02x type %#02x: need %d bytes, have %d", ch, ty, n, len(b)-i)
		}
		v := b[i : i+n]
		i += n

		switch {
		case ch == 0x01 && ty == 0x75: // battery level
			d.Battery = ptr(int(v[0]))
			hasData = true
		case ch == 0x03 && ty == 0x67: // temperature (also threshold alarm packets)
			d.Temperature = ptr(temp10(v))
			hasData = true
		case ch == 0x04 && ty == 0x68: // humidity
			d.Humidity = ptr(float64(v[0]) / 2)
			hasData = true
		case ch == 0x05 && ty == 0x00: // water leak, or digital input on EM300-DI
			if model == ModelDI {
				d.DigitalInput = ptr(int(v[0]))
			} else {
				d.WaterLeak = ptr(int(v[0]))
			}
			hasData = true
		case ch == 0x06 && ty == 0x00: // magnet status
			d.MagnetStatus = ptr(int(v[0]))
			hasData = true
		case ch == 0x05 && ty == 0xc8: // pulse counter (firmware <= V1.2)
			d.PulseCount = ptr(int64(binary.LittleEndian.Uint32(v)))
			hasData = true
		case ch == 0x05 && ty == 0xe1: // pulse counter with conversion
			setPulseConversion(&d, v)
			hasData = true
		case ch == 0x85 && ty == 0x00: // digital input alarm
			d.DigitalInput = ptr(int(v[0]))
			d.Alarm = ptr(int(v[1]))
			if v[1] == 0 {
				d.AlarmType = ptr(diAlarmTypes[6])
			} else {
				d.AlarmType = ptr(diAlarmTypes[5])
			}
			hasData = true
		case ch == 0x85 && ty == 0xe1: // pulse alarm
			setPulseConversion(&d, v[:8])
			name, active := pulseAlarm(v[8])
			d.AlarmType = ptr(name)
			d.Alarm = ptr(active)
			hasData = true
		case ch == 0x03 && ty == 0xed: // liquid level status
			d.LiquidStatus = ptr(int(v[0]))
			hasData = true
		case ch == 0x04 && ty == 0xee: // calibration status
			d.Calibration = ptr(int(v[0]))
			hasData = true
		case ch == 0x83 && ty == 0xed: // liquid level alarm
			d.LiquidStatus = ptr(int(v[0]))
			d.Alarm = ptr(int(v[1]))
			hasData = true
		case ch == 0xff: // basic information
			decodeInfo(&info, ty, v)
			hasInfo = true
		case ch == 0x20 && ty == 0xce: // historical data
			records = append(records, decodeHistory(model, v))
		case ch == 0x21 && ty == 0xce: // EM300-DI historical data
			records = append(records, decodeDIHistory(v))
		case ch == 0xfc: // data enquiry reply
			acked = true
		}
	}

	switch {
	case len(records) > 0:
		return &DatalogData{Records: records}, nil
	case hasData:
		return &d, nil
	case hasInfo:
		return &info, nil
	case acked:
		return nil, decoders.ErrIgnored
	}
	return nil, fmt.Errorf("em300v1: no decodable channels in payload")
}

// dataLen returns the data length in bytes for a channel/type pair.
func dataLen(model Model, ch, ty byte) (int, error) {
	switch {
	case ch == 0x01 && ty == 0x75,
		ch == 0x04 && ty == 0x68,
		ch == 0x05 && ty == 0x00,
		ch == 0x06 && ty == 0x00,
		ch == 0x03 && ty == 0xed,
		ch == 0x04 && ty == 0xee:
		return 1, nil
	case ch == 0x03 && ty == 0x67,
		ch == 0x85 && ty == 0x00,
		ch == 0x83 && ty == 0xed:
		return 2, nil
	case ch == 0x05 && ty == 0xc8:
		return 4, nil
	case ch == 0x05 && ty == 0xe1:
		return 8, nil
	case ch == 0x85 && ty == 0xe1:
		return 9, nil
	case ch == 0xfc && (ty == 0x6b || ty == 0x6c):
		return 1, nil
	case ch == 0x20 && ty == 0xce:
		return historyLen(model)
	case ch == 0x21 && ty == 0xce:
		return 18, nil
	case ch == 0xff:
		switch ty {
		case 0x0b, 0x01, 0x0f:
			return 1, nil
		case 0x09, 0x0a:
			return 2, nil
		case 0x16:
			return 8, nil
		}
	}
	return 0, fmt.Errorf("em300v1: unknown channel %#02x type %#02x", ch, ty)
}

// historyLen is the 20ce record length: 4-byte timestamp + model-specific contents.
func historyLen(model Model) (int, error) {
	switch model {
	case ModelTH:
		return 4 + 3, nil // temperature + humidity
	case ModelMCS, ModelSLD, ModelZLD:
		return 4 + 4, nil // temperature + humidity + door/leak status
	case ModelMLD:
		return 4 + 1, nil // leak status
	case ModelDI:
		return 4 + 9, nil // temperature + humidity + interface type + counter + digital (firmware <= V1.2)
	}
	return 0, fmt.Errorf("em300v1: historical data not supported for %s", model)
}

func decodeInfo(info *DeviceInfo, ty byte, v []byte) {
	switch ty {
	case 0x0b:
		info.PowerOn = true
	case 0x01:
		info.ProtocolVersion = fmt.Sprintf("V%d", v[0])
	case 0x09:
		info.HardwareVersion = fmt.Sprintf("V%d.%d", v[0], v[1]>>4)
	case 0x0a:
		info.SoftwareVersion = fmt.Sprintf("V%x.%x", v[0], v[1])
	case 0x0f:
		switch v[0] {
		case 0:
			info.DeviceClass = "Class A"
		case 1:
			info.DeviceClass = "Class B"
		case 2:
			info.DeviceClass = "Class C"
		default:
			info.DeviceClass = fmt.Sprintf("Unknown (%d)", v[0])
		}
	case 0x16:
		info.SerialNumber = hex.EncodeToString(v)
	}
}

func decodeHistory(model Model, v []byte) DatalogEntry {
	rec := DatalogEntry{Timestamp: int64(binary.LittleEndian.Uint32(v))}
	rest := v[4:]
	switch model {
	case ModelTH:
		rec.Temperature = ptr(temp10(rest))
		rec.Humidity = ptr(float64(rest[2]) / 2)
	case ModelMCS:
		rec.Temperature = ptr(temp10(rest))
		rec.Humidity = ptr(float64(rest[2]) / 2)
		rec.MagnetStatus = ptr(int(rest[3]))
	case ModelSLD, ModelZLD:
		rec.Temperature = ptr(temp10(rest))
		rec.Humidity = ptr(float64(rest[2]) / 2)
		rec.WaterLeak = ptr(int(rest[3]))
	case ModelMLD:
		rec.WaterLeak = ptr(int(rest[0]))
	case ModelDI:
		rec.Temperature = ptr(temp10(rest))
		rec.Humidity = ptr(float64(rest[2]) / 2)
		if rest[3] == 0 {
			rec.InterfaceType = ptr("digital")
			rec.DigitalInput = ptr(int(rest[8]))
		} else {
			rec.InterfaceType = ptr("counter")
			rec.PulseCount = ptr(int64(binary.LittleEndian.Uint32(rest[4:8])))
		}
	}
	return rec
}

func decodeDIHistory(v []byte) DatalogEntry {
	rec := DatalogEntry{Timestamp: int64(binary.LittleEndian.Uint32(v))}
	rec.Temperature = ptr(temp10(v[4:6]))
	rec.Humidity = ptr(float64(v[6]) / 2)
	if name, ok := diAlarmTypes[v[7]]; ok {
		rec.AlarmType = ptr(name)
	}
	if v[8] == 0 {
		rec.InterfaceType = ptr("digital")
		rec.DigitalInput = ptr(int(v[9]))
	} else {
		rec.InterfaceType = ptr("counter")
	}
	rec.WaterConv = ptr(float64(binary.LittleEndian.Uint16(v[10:12])) / 10)
	rec.PulseConv = ptr(float64(binary.LittleEndian.Uint16(v[12:14])) / 10)
	rec.WaterConsumption = ptr(float64(math.Float32frombits(binary.LittleEndian.Uint32(v[14:18]))))
	return rec
}

// setPulseConversion decodes the 8-byte pulse conversion block:
// water_conv (2B) + pulse_conv (2B) + water consumption (float32).
func setPulseConversion(d *Data, v []byte) {
	d.WaterConv = ptr(float64(binary.LittleEndian.Uint16(v[0:2])) / 10)
	d.PulseConv = ptr(float64(binary.LittleEndian.Uint16(v[2:4])) / 10)
	d.WaterConsumption = ptr(float64(math.Float32frombits(binary.LittleEndian.Uint32(v[4:8]))))
}

func pulseAlarm(status byte) (name string, active int) {
	name, ok := diAlarmTypes[status]
	if !ok {
		name = fmt.Sprintf("unknown_alarm_%d", status)
	}
	if status == 1 || status == 3 || status == 5 {
		active = 1
	}
	return name, active
}

func temp10(v []byte) float64 {
	return float64(int16(binary.LittleEndian.Uint16(v))) / 10
}

func ptr[T any](v T) *T { return &v }
