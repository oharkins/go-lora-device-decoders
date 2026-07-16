package ltc2v1_test

import (
	"errors"
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/ltc2v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := ltc2v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x01, 0x08, 0x56, 0xFD, 0xC9, 0x01, 0x02, 0x03, 0x04}
	v, err := ltc2v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*ltc2v1.Data)

	if d.Ext != 1 || d.BatV != 3 || d.SysTimestamp != 16909060 {
		t.Fatalf("decoded data = %#v", d)
	}
	if d.TempChannel1 == nil || *d.TempChannel1 != 21.34 || d.TempChannel2 == nil || *d.TempChannel2 != -5.67 {
		t.Fatalf("decoded temperatures = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestDecodePollStatusIgnored(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x41, 0x08, 0x56, 0xFD, 0xC9, 0x01, 0x02, 0x03, 0x04}
	v, err := ltc2v1.Decode(decoders.Uplink{Payload: payload})
	if !errors.Is(err, decoders.ErrIgnored) {
		t.Fatalf("error = %v, want ErrIgnored", err)
	}
	if v != nil {
		t.Fatalf("ignored decode value = %#v, want nil", v)
	}
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "ltc2", "v1")
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
