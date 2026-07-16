package lt22222lv1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lt22222lv1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := lt22222lv1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{0x04, 0xB0, 0x09, 0xC4, 0x30, 0x39, 0x1A, 0x85, 0x18, 0x00, 0x41}
	v, err := lt22222lv1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lt22222lv1.Data)

	if d.HardwareMode != "LT22222" || d.WorkMode != "2ACI+2AVI" {
		t.Fatalf("decoded modes = %#v", d)
	}
	if d.AVI1V == nil || *d.AVI1V != 1.2 || d.AVI2V == nil || *d.AVI2V != 2.5 {
		t.Fatalf("decoded voltage readings = %#v", d)
	}
	if d.ACI1MA == nil || *d.ACI1MA != 12.345 || d.ACI2MA == nil || *d.ACI2MA != 6.789 {
		t.Fatalf("decoded current readings = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "lt22222-l", "v1")
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
