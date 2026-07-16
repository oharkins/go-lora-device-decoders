package llds12v1_test

import (
	"errors"
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/llds12v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := llds12v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	// bat=3.000V, ds=25.0C, dist=100.0cm, signal=100, interrupt=0, lidarTemp=25, msg=1
	payload := []byte{0x0B, 0xB8, 0x00, 0xFA, 0x03, 0xE8, 0x00, 0x64, 0x00, 0x19, 0x01}
	v, err := llds12v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*llds12v1.Data)
	if d.BatV != 3 || d.TempCDS18B20 != 25 || d.LidarDistanceCM != 100 || d.LidarSignalStrength != 100 {
		t.Fatalf("decoded = %#v", d)
	}
	if d.MessageKind() != decoders.KindTelemetry {
		t.Fatalf("MessageKind = %q", d.MessageKind())
	}
	ms := d.Measurements()
	if len(ms) == 0 {
		t.Fatal("empty measurements")
	}
	for _, m := range ms {
		if !m.Valid {
			t.Fatalf("invalid measurement: %#v", m)
		}
	}
}

func TestDecodeIgnored(t *testing.T) {
	payload := make([]byte, 11)
	payload[0], payload[10] = 0x03, 0x02
	_, err := llds12v1.Decode(decoders.Uplink{Payload: payload})
	if !errors.Is(err, decoders.ErrIgnored) {
		t.Fatalf("want ErrIgnored, got %v", err)
	}
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "llds12", "v1")
	if !ok || len(d.Offers()) == 0 {
		t.Fatal("registry lookup failed")
	}
}
