package lht65nv1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lht65nv1"
)

func TestDecodeShortPayload(t *testing.T) {
	if _, err := lht65nv1.Decode(decoders.Uplink{FPort: 2, Payload: []byte{0x01}}); err == nil {
		t.Fatal("want error for short payload")
	}
}

func TestDeviceInfoKind(t *testing.T) {
	// FPort 5 device info: model + fw + freq + subband + bat
	payload := []byte{0x01, 0x01, 0x05, 0x01, 0x00, 0x0B, 0xB8}
	v, err := lht65nv1.Decode(decoders.Uplink{FPort: 5, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	info, ok := v.(*lht65nv1.DeviceInfo)
	if !ok {
		t.Fatalf("got %T, want *DeviceInfo", v)
	}
	if info.MessageKind() != decoders.KindDeviceInfo {
		t.Fatalf("MessageKind = %q, want device_info", info.MessageKind())
	}
	if decoders.KindOf(v) != decoders.KindDeviceInfo {
		t.Fatalf("KindOf = %q", decoders.KindOf(v))
	}
	ms := info.Measurements()
	if len(ms) == 0 || ms[0].Name != decoders.BatteryVoltage || !ms[0].Valid {
		t.Fatalf("measurements = %#v", ms)
	}
}

func TestDatalogAckKind(t *testing.T) {
	payload := make([]byte, 11)
	payload[6] = 0x40 // pollStatus bit
	v, err := lht65nv1.Decode(decoders.Uplink{FPort: 2, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	ack, ok := v.(*lht65nv1.DatalogAck)
	if !ok {
		t.Fatalf("got %T, want *DatalogAck", v)
	}
	if ack.MessageKind() != decoders.KindDatalog {
		t.Fatalf("MessageKind = %q", ack.MessageKind())
	}
	if len(ack.Measurements()) != 0 {
		t.Fatalf("want empty measurements, got %#v", ack.Measurements())
	}
}

func TestRegistryGet(t *testing.T) {
	d, ok := decoders.Get("dragino", "lht65n", "v1")
	if !ok || len(d.Offers()) == 0 {
		t.Fatal("registry lookup failed")
	}
}
