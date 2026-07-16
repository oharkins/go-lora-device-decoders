package lgt92v1_test

import (
	"math"
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lgt92v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := lgt92v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{
		0x03, 0x11, 0xF0, 0xC8,
		0xFF, 0xFE, 0x0C, 0xC8,
		0x4C, 0xE4, 0x64,
		0x04, 0xD2, 0xFD, 0xC9,
		0x96, 0x30, 0x39,
	}
	v, err := lgt92v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lgt92v1.Data)

	if !approx(d.Latitude, 51.5074) || !approx(d.Longitude, -0.1278) || !approx(d.BatV, 3.3) {
		t.Fatalf("decoded location/battery = %#v", d)
	}
	if d.AlarmStatus != "TRUE" || d.MotionMode != "Move" || d.LEDUpDown != "ON" || d.Firmware != 164 {
		t.Fatalf("decoded status = %#v", d)
	}
	if !approx(d.Roll, 12.34) || !approx(d.Pitch, -5.67) || !approx(d.HDOP, 1.5) || !approx(d.Altitude, 123.45) {
		t.Fatalf("decoded attitude = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "lgt92", "v1")
	if !ok {
		t.Fatal("registered decoder not found")
	}
	if len(d.Offers()) == 0 {
		t.Fatal("offers are empty")
	}
}

func approx(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
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
