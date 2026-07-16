package ldds20v1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/ldds20v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := ldds20v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x01, 0xF4, 0x01}
	v, err := ldds20v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*ldds20v1.Data)

	if d.BatV != 3 || d.DistanceMM == nil || *d.DistanceMM != 500 || d.InterruptStatus != 1 {
		t.Fatalf("decoded data = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestDecodeInvalidDistance(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x00, 0x13, 0x00}
	v, err := ldds20v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*ldds20v1.Data)
	if d.DistanceMM == nil || *d.DistanceMM != 19 || d.DistanceStatus != "Invalid Reading" {
		t.Fatalf("decoded invalid distance = %#v", d)
	}
	for _, m := range d.Measurements() {
		if m.Name != decoders.DistanceMM {
			continue
		}
		if m.Valid || m.Quality != decoders.QualityInvalid || m.Value != 19 {
			t.Fatalf("distance measurement = %#v, want invalid value 19", m)
		}
		return
	}
	t.Fatal("distance_mm measurement not found")
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "ldds20", "v1")
	if !ok {
		t.Fatal("registered decoder not found")
	}
	if len(d.Offers()) == 0 {
		t.Fatal("offers are empty")
	}
}

func assertAllValid(t *testing.T, ms []decoders.Measurement) {
	t.Helper()
	if len(ms) == 0 {
		t.Fatal("measurements are empty")
	}
	for _, m := range ms {
		if !m.Valid {
			t.Fatalf("measurement not valid: %#v", m)
		}
	}
}
