package laq4v1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/laq4v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := laq4v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x05, 0x00, 0x7B, 0x01, 0x90, 0x00, 0xD7, 0x02, 0x2B}
	v, err := laq4v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*laq4v1.Data)

	if d.BatV != 3 || d.WorkMode != "CO2" || d.Alarm == nil || !*d.Alarm {
		t.Fatalf("decoded data = %#v", d)
	}
	if d.TVOCPPB == nil || *d.TVOCPPB != 123 || d.CO2PPM == nil || *d.CO2PPM != 400 {
		t.Fatalf("decoded gas readings = %#v", d)
	}
	if d.TempCSHT == nil || *d.TempCSHT != 21.5 || d.HumSHT == nil || *d.HumSHT != 55.5 {
		t.Fatalf("decoded SHT readings = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "laq4", "v1")
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
