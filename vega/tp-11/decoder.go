// Package tp11 decodes Vega TP-11 4-20 mA transmitter uplinks.
//
// The base decoder (registered as vega/tp-11/v1) returns raw mA readings.
// To add engineering-unit conversion, build a configured decoder with NewDecoder
// and register it under your own product key:
//
//	decoders.Register("vega", "tp-11-0-5m", "v1",
//	    tp11.NewDecoder(tp11.RangeConfig{MinVal: 0, MaxVal: 5, Unit: "m"}))
package tp11

import (
	"fmt"
	"math"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func init() {
	decoders.Register("vega", "tp-11", "v1", decoders.New(Decode, baseOffers()...))
}

// RangeConfig maps the 4-20 mA signal to an engineering unit using a linear
// scale. MinVal corresponds to 4 mA, MaxVal to 20 mA.
type RangeConfig struct {
	MinVal float64
	MaxVal float64
	Unit   string
}

// Convert maps a mA value to an engineering unit value.
// Returns -1 if mA is below 4 (sensor fault / not connected).
func (r RangeConfig) Convert(mA float64) float64 {
	if mA < 4 {
		return -1
	}
	v := ((mA-4)/16)*(r.MaxVal-r.MinVal) + r.MinVal
	return math.Max(r.MinVal, math.Min(r.MaxVal, v))
}

// NewDecoder returns a Decoder that applies cfg's linear conversion on top of
// the standard TP-11 byte parsing. The resulting Data will have Value,
// ValueLow, ValueHigh, and Unit populated in addition to the raw mA fields.
func NewDecoder(cfg RangeConfig) decoders.Decoder {
	return decoders.New(func(u decoders.Uplink) (any, error) {
		raw, err := Decode(u)
		if err != nil || raw == nil {
			return raw, err
		}
		d := raw.(*Data)
		v := cfg.Convert(d.MA)
		vl := cfg.Convert(d.MALow)
		vh := cfg.Convert(d.MAHigh)
		d.Value = &v
		d.ValueLow = &vl
		d.ValueHigh = &vh
		d.Unit = cfg.Unit
		return d, nil
	}, configuredOffers(cfg.Unit)...)
}

func baseOffers() []decoders.Offering {
	return []decoders.Offering{
		decoders.Offer(decoders.BatteryPercent, decoders.Percent),
		decoders.Offer(decoders.CurrentMA, decoders.MilliAmp),
		decoders.Offer(decoders.CurrentMALow, decoders.MilliAmp),
		decoders.Offer(decoders.CurrentMAHigh, decoders.MilliAmp),
		decoders.Offer(decoders.Temperature, decoders.Celsius),
	}
}

func configuredOffers(unit string) []decoders.Offering {
	offers := baseOffers()
	offers = append(offers,
		decoders.Offer(decoders.Value, unit),
		decoders.Offer("value_low", unit),
		decoders.Offer("value_high", unit),
	)
	return offers
}

var reasons = [...]string{
	"Sending packet by the time",
	"By the security input 1 triggering",
	"By the security input 2 triggering",
	"External power state was change",
	"Measurement is out of the specified limits",
	"Transmitting by the request",
}

// Data is the decoded TP-11 uplink.
// Value, ValueLow, ValueHigh, and Unit are only set when using NewDecoder.
type Data struct {
	Reason            string   `json:"reason"`
	BatteryPercentage int      `json:"battery_percentage"`
	MA                float64  `json:"ma"`
	MALow             float64  `json:"ma_low"`
	MAHigh            float64  `json:"ma_high"`
	Temperature       int      `json:"temperature"`
	Value             *float64 `json:"value,omitempty"`
	ValueLow          *float64 `json:"value_low,omitempty"`
	ValueHigh         *float64 `json:"value_high,omitempty"`
	Unit              string   `json:"unit,omitempty"`
}

func (d *Data) MessageKind() decoders.Kind { return decoders.KindTelemetry }

func (d *Data) Measurements() []decoders.Measurement {
	ms := []decoders.Measurement{
		decoders.Int(decoders.BatteryPercent, decoders.Percent, d.BatteryPercentage),
		currentMeasurement(decoders.CurrentMA, d.MA),
		currentMeasurement(decoders.CurrentMALow, d.MALow),
		currentMeasurement(decoders.CurrentMAHigh, d.MAHigh),
		decoders.Int(decoders.Temperature, decoders.Celsius, d.Temperature),
	}
	ms = appendValueMeasurement(ms, decoders.Value, d.Unit, d.Value)
	ms = appendValueMeasurement(ms, "value_low", d.Unit, d.ValueLow)
	ms = appendValueMeasurement(ms, "value_high", d.Unit, d.ValueHigh)
	return ms
}

func currentMeasurement(name string, v float64) decoders.Measurement {
	if v < 4 {
		return decoders.FloatQuality(name, decoders.MilliAmp, v, false, decoders.QualityFault)
	}
	return decoders.Float(name, decoders.MilliAmp, v)
}

func appendValueMeasurement(ms []decoders.Measurement, name, unit string, v *float64) []decoders.Measurement {
	if v == nil {
		return ms
	}
	return append(ms, valueMeasurement(name, unit, *v))
}

func valueMeasurement(name, unit string, v float64) decoders.Measurement {
	if v == -1 {
		return decoders.FloatQuality(name, unit, v, false, decoders.QualityFault)
	}
	return decoders.Float(name, unit, v)
}

// Decode decodes a raw TP-11 uplink. mA values are present; Value fields are
// not populated — use NewDecoder for engineering-unit conversion.
func Decode(u decoders.Uplink) (any, error) {
	b := u.Payload

	if u.FPort == 4 {
		return nil, decoders.ErrIgnored
	}
	if len(b) < 16 {
		return nil, fmt.Errorf("tp11: payload too short: %d bytes (want >= 16)", len(b))
	}

	reasonIdx := b[12] & 0x7F
	reason := "Unknown reason"
	if int(reasonIdx) < len(reasons) {
		reason = reasons[reasonIdx]
	}

	return &Data{
		Reason:            reason,
		BatteryPercentage: int(b[1]),
		MA:                float64(uint16(b[15])<<8|uint16(b[14])) / 100,
		MALow:             float64(uint16(b[9])<<8|uint16(b[8])) / 100,
		MAHigh:            float64(uint16(b[11])<<8|uint16(b[10])) / 100,
		Temperature:       int(b[7]),
	}, nil
}
