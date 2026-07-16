package tp11_test

import (
	"encoding/base64"
	"errors"
	"math"
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	tp11 "github.com/oharkins/go-lora-device-decoders/vega/tp-11"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

type testCase struct {
	name        string
	b64         string
	fport       uint8
	reason      string
	batteryPct  int
	temperature int
	ma          float64
	maLow       float64
	maHigh      float64
	maLowValid  bool
	// expected values after 0-5 m conversion
	value     float64
	valueLow  float64
	valueHigh float64
}

var cases = []testCase{
	{
		name:        "sample 1",
		b64:         "AS4A8renZByQAdAHAADqAw==",
		fport:       2,
		reason:      "Sending packet by the time",
		batteryPct:  46,
		temperature: 28,
		ma:          10.02,
		maLow:       4.00,
		maHigh:      20.00,
		maLowValid:  true,
		value:       1.88125, // ((10.02-4)/16)*5
		valueLow:    0.0,     // 4 mA → 0 m
		valueHigh:   5.0,     // 20 mA → 5 m
	},
	{
		name:        "sample 2",
		b64:         "AT8A9WXpZR9eAdAHAACEAw==",
		fport:       2,
		reason:      "Sending packet by the time",
		batteryPct:  63,
		temperature: 31,
		ma:          9.00,
		maLow:       3.50, // below 4 mA → fault
		maHigh:      20.00,
		maLowValid:  false,
		value:       1.5625, // ((9.00-4)/16)*5
		valueLow:    -1,     // 3.50 mA → fault
		valueHigh:   5.0,
	},
}

func mustDecode(b64 string) []byte {
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(err)
	}
	return b
}

func assertOffers(t *testing.T, got []decoders.Offering, want []decoders.Offering) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("offers len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("offer %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestOffers(t *testing.T) {
	dec, ok := decoders.Get("vega", "tp-11", "v1")
	if !ok {
		t.Fatal("registered decoder not found")
	}
	baseOffers := []decoders.Offering{
		decoders.Offer(decoders.BatteryPercent, decoders.Percent),
		decoders.Offer(decoders.CurrentMA, decoders.MilliAmp),
		decoders.Offer(decoders.CurrentMALow, decoders.MilliAmp),
		decoders.Offer(decoders.CurrentMAHigh, decoders.MilliAmp),
		decoders.Offer(decoders.Temperature, decoders.Celsius),
	}
	assertOffers(t, dec.Offers(), baseOffers)

	configured := tp11.NewDecoder(tp11.RangeConfig{MinVal: 0, MaxVal: 5, Unit: "m"})
	assertOffers(t, configured.Offers(), append(baseOffers,
		decoders.Offer(decoders.Value, "m"),
		decoders.Offer("value_low", "m"),
		decoders.Offer("value_high", "m"),
	))
}

func TestDecode_Base(t *testing.T) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := tp11.Decode(decoders.Uplink{FPort: tc.fport, Payload: mustDecode(tc.b64)})
			if err != nil {
				t.Fatal(err)
			}
			d := v.(*tp11.Data)

			if d.Reason != tc.reason {
				t.Errorf("reason = %q, want %q", d.Reason, tc.reason)
			}
			if d.BatteryPercentage != tc.batteryPct {
				t.Errorf("battery_percentage = %d, want %d", d.BatteryPercentage, tc.batteryPct)
			}
			if d.Temperature != tc.temperature {
				t.Errorf("temperature = %d, want %d", d.Temperature, tc.temperature)
			}
			if d.MA != tc.ma {
				t.Errorf("ma = %v, want %v", d.MA, tc.ma)
			}
			if d.MALow != tc.maLow {
				t.Errorf("ma_low = %v, want %v", d.MALow, tc.maLow)
			}
			if d.MAHigh != tc.maHigh {
				t.Errorf("ma_high = %v, want %v", d.MAHigh, tc.maHigh)
			}
			if d.Value != nil || d.ValueLow != nil || d.ValueHigh != nil {
				t.Error("base decoder should not populate Value fields")
			}
			if got := d.MessageKind(); got != decoders.KindTelemetry {
				t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
			}
			ms, ok := decoders.MeasurementsOf(d)
			if !ok {
				t.Fatal("base data should expose measurements")
			}
			if len(ms) != 5 {
				t.Fatalf("measurements len = %d, want 5: %#v", len(ms), ms)
			}
			assertMeasurementValid(t, ms, decoders.CurrentMA, true)
			assertMeasurementValid(t, ms, decoders.CurrentMALow, tc.maLowValid)
			assertMeasurementValid(t, ms, decoders.CurrentMAHigh, true)
		})
	}
}

func TestDecode_WithRange_0_5m(t *testing.T) {
	dec := tp11.NewDecoder(tp11.RangeConfig{MinVal: 0, MaxVal: 5, Unit: "m"})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := dec.Decode(decoders.Uplink{FPort: tc.fport, Payload: mustDecode(tc.b64)})
			if err != nil {
				t.Fatal(err)
			}
			d := v.(*tp11.Data)

			if d.Unit != "m" {
				t.Errorf("unit = %q, want \"m\"", d.Unit)
			}
			if d.Value == nil {
				t.Fatal("value is nil")
			}
			if !approxEqual(*d.Value, tc.value) {
				t.Errorf("value = %v, want %v", *d.Value, tc.value)
			}
			if d.ValueLow == nil {
				t.Fatal("value_low is nil")
			}
			if !approxEqual(*d.ValueLow, tc.valueLow) {
				t.Errorf("value_low = %v, want %v", *d.ValueLow, tc.valueLow)
			}
			if d.ValueHigh == nil {
				t.Fatal("value_high is nil")
			}
			if !approxEqual(*d.ValueHigh, tc.valueHigh) {
				t.Errorf("value_high = %v, want %v", *d.ValueHigh, tc.valueHigh)
			}
			ms, ok := decoders.MeasurementsOf(d)
			if !ok {
				t.Fatal("configured data should expose measurements")
			}
			if len(ms) != 8 {
				t.Fatalf("measurements len = %d, want 8: %#v", len(ms), ms)
			}
			if got := ms[5]; got.Name != decoders.Value || got.Unit != "m" || !approxEqual(got.Value, tc.value) {
				t.Fatalf("value measurement = %#v, want value=%v unit=m", got, tc.value)
			}
			if got := ms[6]; got.Name != "value_low" || got.Unit != "m" || !approxEqual(got.Value, tc.valueLow) {
				t.Fatalf("value_low measurement = %#v, want value=%v unit=m", got, tc.valueLow)
			}
			if got := ms[7]; got.Name != "value_high" || got.Unit != "m" || !approxEqual(got.Value, tc.valueHigh) {
				t.Fatalf("value_high measurement = %#v, want value=%v unit=m", got, tc.valueHigh)
			}
			assertMeasurementValid(t, ms, decoders.Value, true)
			assertMeasurementValid(t, ms, "value_low", tc.maLowValid)
			assertMeasurementValid(t, ms, "value_high", true)
		})
	}
}

func assertMeasurementValid(t *testing.T, ms []decoders.Measurement, name string, valid bool) {
	t.Helper()
	for _, m := range ms {
		if m.Name != name {
			continue
		}
		if m.Valid != valid {
			t.Fatalf("%s Valid = %v, want %v in %#v", name, m.Valid, valid, ms)
		}
		if !valid && m.Quality != decoders.QualityFault {
			t.Fatalf("%s Quality = %q, want %q", name, m.Quality, decoders.QualityFault)
		}
		return
	}
	t.Fatalf("measurement %q not found in %#v", name, ms)
}

func TestDecode_Port4(t *testing.T) {
	v, err := tp11.Decode(decoders.Uplink{FPort: 4, Payload: mustDecode(cases[0].b64)})
	if !errors.Is(err, decoders.ErrIgnored) {
		t.Fatalf("error = %v, want ErrIgnored", err)
	}
	if v != nil {
		t.Errorf("port 4 should return nil, got %v", v)
	}
}

func TestDecode_ShortPayload(t *testing.T) {
	_, err := tp11.Decode(decoders.Uplink{FPort: 2, Payload: []byte{0x01, 0x2E}})
	if err == nil {
		t.Error("want error for short payload")
	}
}

func TestRegistryLookup(t *testing.T) {
	v, err := decoders.Decode("vega", "tp-11", "v1", decoders.Uplink{FPort: 2, Payload: mustDecode(cases[0].b64)})
	if err != nil {
		t.Fatal(err)
	}
	if v == nil {
		t.Error("registry decode returned nil")
	}
}
