package lsph01v1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lsph01v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := lsph01v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x00, 0xFA, 0x02, 0xA6, 0x00, 0xD7, 0x01, 0x00, 0x02}
	v, err := lsph01v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lsph01v1.Data)

	if d.BatV != 3 || d.TempCDS18B20 != 25 || d.PH1Soil != 6.78 || d.TempSoil != 21.5 {
		t.Fatalf("decoded readings = %#v", d)
	}
	if d.InterruptFlag != 1 || d.MessageType != 2 {
		t.Fatalf("decoded flags = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "lsph01", "v1")
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
