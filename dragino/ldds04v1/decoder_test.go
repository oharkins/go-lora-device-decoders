package ldds04v1_test

import (
	"errors"
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/ldds04v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := ldds04v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x03, 0xE9, 0x07, 0xD2, 0x0B, 0xBB, 0x0F, 0xA4, 0x01}
	v, err := ldds04v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*ldds04v1.Data)

	if d.BatV != 3 || d.EXTITrigger || d.Distance1CM != 100.1 || d.Distance2CM != 200.2 {
		t.Fatalf("decoded first readings = %#v", d)
	}
	if d.Distance3CM != 300.3 || d.Distance4CM != 400.4 || d.MesType != 1 {
		t.Fatalf("decoded later readings = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestDecodeIgnored(t *testing.T) {
	payload := []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
	v, err := ldds04v1.Decode(decoders.Uplink{Payload: payload})
	if !errors.Is(err, decoders.ErrIgnored) {
		t.Fatalf("error = %v, want ErrIgnored", err)
	}
	if v != nil {
		t.Fatalf("ignored decode value = %#v, want nil", v)
	}
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "ldds04", "v1")
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
