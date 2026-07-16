package decoders_test

import (
	"errors"
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
)

func TestNewRequiresOffersCopy(t *testing.T) {
	offers := []decoders.Offering{decoders.Offer("temperature", "C")}
	d := decoders.New(func(decoders.Uplink) (any, error) {
		return map[string]float64{"temperature": 21.5}, nil
	}, offers...)
	offers[0].Name = "mutated"
	got := d.Offers()
	if len(got) != 1 || got[0].Name != "temperature" {
		t.Fatalf("Offers() = %#v, want temperature (copied)", got)
	}
}

func TestRegisterAndOffers(t *testing.T) {
	const m, p, v = "testco", "widget", "v1"
	decoders.Register(m, p, v, decoders.New(
		func(u decoders.Uplink) (any, error) {
			if len(u.Payload) == 0 {
				return nil, decoders.ErrIgnored
			}
			return &sample{Temp: 20, Hum: 40}, nil
		},
		decoders.Offer("temperature", "C"),
		decoders.Offer("humidity", "%"),
	))

	offers, ok := decoders.Offers(m, p, v)
	if !ok {
		t.Fatal("Offers lookup failed")
	}
	if len(offers) != 2 || offers[0].Name != "temperature" || offers[1].Unit != "%" {
		t.Fatalf("unexpected offers: %#v", offers)
	}

	raw, err := decoders.Decode(m, p, v, decoders.Uplink{Payload: []byte{1}})
	if err != nil {
		t.Fatal(err)
	}
	ms, ok := decoders.MeasurementsOf(raw)
	if !ok || len(ms) != 2 {
		t.Fatalf("MeasurementsOf = %#v ok=%v", ms, ok)
	}

	_, err = decoders.Decode(m, p, v, decoders.Uplink{})
	if !errors.Is(err, decoders.ErrIgnored) {
		t.Fatalf("want ErrIgnored, got %v", err)
	}
}

func TestParseKey(t *testing.T) {
	k, err := decoders.ParseKey(" Dragino / LHT65 / V1 ")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != "dragino/lht65/v1" {
		t.Fatalf("got %s", k)
	}
	if _, err := decoders.ParseKey("bad"); err == nil {
		t.Fatal("want error")
	}
}

func TestAppendHelpers(t *testing.T) {
	temp := 22.5
	var missing *float64
	count := 3
	ms := []decoders.Measurement{decoders.Float("battery_voltage", "V", 3.1)}
	ms = decoders.AppendFloat(ms, "temperature", "C", &temp)
	ms = decoders.AppendFloat(ms, "missing", "C", missing)
	ms = decoders.AppendInt(ms, "count", "count", &count)
	if len(ms) != 3 {
		t.Fatalf("len=%d want 3: %#v", len(ms), ms)
	}
}

func TestKindOfDefaultsToTelemetry(t *testing.T) {
	if got := decoders.KindOf(struct{}{}); got != decoders.KindTelemetry {
		t.Fatalf("KindOf(non-message) = %q, want %q", got, decoders.KindTelemetry)
	}
}

func TestFloatHelpersValidity(t *testing.T) {
	m := decoders.Float("temperature", "C", 21.5)
	if !m.Valid {
		t.Fatalf("Float() Valid = false, want true: %#v", m)
	}

	fault := decoders.FloatQuality("temperature", "C", -1, false, decoders.QualityFault)
	if fault.Valid {
		t.Fatalf("FloatQuality() Valid = true, want false: %#v", fault)
	}
	if fault.Quality != decoders.QualityFault {
		t.Fatalf("FloatQuality() Quality = %q, want %q", fault.Quality, decoders.QualityFault)
	}

	valid := decoders.FloatQuality("temperature", "C", 21.5, true, decoders.QualityFault)
	if !valid.Valid || valid.Quality != "" {
		t.Fatalf("valid FloatQuality() = %#v, want Valid true with empty Quality", valid)
	}
}

type sample struct {
	Temp float64
	Hum  float64
}

func (s *sample) Measurements() []decoders.Measurement {
	return []decoders.Measurement{
		decoders.Float("temperature", "C", s.Temp),
		decoders.Float("humidity", "%", s.Hum),
	}
}
