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

func TestDecodeDS18B20Disconnected(t *testing.T) {
	// Base64: y40B4APOAX//f/8=
	// Ext=1 (DS18B20), probe returns 0x7FFF sentinel (not connected).
	payload := []byte{0xCB, 0x8D, 0x01, 0xE0, 0x03, 0xCE, 0x01, 0x7F, 0xFF, 0x7F, 0xFF}
	v, err := lht65nv1.Decode(decoders.Uplink{FPort: 2, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d, ok := v.(*lht65nv1.Data)
	if !ok {
		t.Fatalf("got %T, want *Data", v)
	}
	if d.BatStatus != "Good" {
		t.Errorf("bat_status = %q, want \"Good\"", d.BatStatus)
	}
	if d.BatV == nil || *d.BatV != 2.957 {
		t.Errorf("bat_v = %v, want 2.957", d.BatV)
	}
	if d.TempCSHT == nil || *d.TempCSHT != 4.8 {
		t.Errorf("temp_c_sht = %v, want 4.8", d.TempCSHT)
	}
	if d.HumSHT == nil || *d.HumSHT != 97.4 {
		t.Errorf("hum_sht = %v, want 97.4", d.HumSHT)
	}
	if d.ExtSensor != "Temperature Sensor" {
		t.Errorf("ext_sensor = %q, want \"Temperature Sensor\"", d.ExtSensor)
	}
	if d.TempCDS == nil || *d.TempCDS != 327.67 {
		t.Errorf("temp_c_ds = %v, want 327.67 (0x7FFF sentinel)", d.TempCDS)
	}
}
