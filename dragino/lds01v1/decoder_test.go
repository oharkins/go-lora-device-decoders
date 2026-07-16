package lds01v1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lds01v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := lds01v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeDoorModeHappyPath(t *testing.T) {
	payload := []byte{0x8B, 0xB8, 0x01, 0x00, 0x00, 0x05, 0x00, 0x00, 0x1E, 0x01}
	v, err := lds01v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lds01v1.Data)

	if d.BatV != 3 || d.Mod != 1 {
		t.Fatalf("decoded data = %#v", d)
	}
	if d.DoorOpenStatus == nil || *d.DoorOpenStatus != 1 || d.DoorOpenTimes == nil || *d.DoorOpenTimes != 5 {
		t.Fatalf("decoded door fields = %#v", d)
	}
	if d.LastDoorOpenDuration == nil || *d.LastDoorOpenDuration != 30 || d.Alarm == nil || *d.Alarm != 1 {
		t.Fatalf("decoded duration/alarm = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "lds01", "v1")
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
