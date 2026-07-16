package lsn50v1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lsn50v1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := lsn50v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDecodeHappyPath(t *testing.T) {
	payload := []byte{0x0B, 0xB8, 0x00, 0xD7, 0x04, 0xD2, 0x83, 0x00, 0xEA, 0x02, 0x37}
	v, err := lsn50v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lsn50v1.Data)

	if d.WorkMode != "IIC" || d.BatV != 3 || d.DigitalIStatus != "H" || d.EXTITrigger != "TRUE" || d.DoorStatus != "CLOSE" {
		t.Fatalf("decoded status = %#v", d)
	}
	if d.TempC1 == nil || *d.TempC1 != 21.5 || d.ADCCH0V == nil || *d.ADCCH0V != 1.234 {
		t.Fatalf("decoded built-in readings = %#v", d)
	}
	if d.TempCSHT == nil || *d.TempCSHT != 23.4 || d.HumSHT == nil || *d.HumSHT != 56.7 {
		t.Fatalf("decoded SHT readings = %#v", d)
	}
	if got := d.MessageKind(); got != decoders.KindTelemetry {
		t.Fatalf("MessageKind() = %q, want %q", got, decoders.KindTelemetry)
	}
	assertAllValid(t, d.Measurements())
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "lsn50", "v1")
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
