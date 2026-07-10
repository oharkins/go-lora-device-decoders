package lht65v1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lht65v1"
)

func TestDecodeBuiltIn(t *testing.T) {
	// batt=3.058V (0x0BF2), temp=22.71°C (0x08DF), hum=56.2% (0x0232),
	// sensorType=1, ext temp=25.00°C (0x09C4)
	payload := []byte{0x0B, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x01, 0x09, 0xC4}

	v, err := lht65v1.Decode(decoders.Uplink{FPort: 2, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lht65v1.Data)

	if d.BatteryVoltage != 3.058 {
		t.Errorf("battery = %v, want 3.058", d.BatteryVoltage)
	}
	if d.Temperature != 22.71 {
		t.Errorf("temperature = %v, want 22.71", d.Temperature)
	}
	if d.Humidity != 56.2 {
		t.Errorf("humidity = %v, want 56.2", d.Humidity)
	}
	if d.SensorType != "Temperature Sensor" {
		t.Errorf("sensor_type = %q", d.SensorType)
	}
	if d.ExternalTemperature == nil || *d.ExternalTemperature != 25.00 {
		t.Errorf("external_temperature = %v, want 25.00", d.ExternalTemperature)
	}
}

func TestNegativeTemperature(t *testing.T) {
	// temp = -5.00°C -> int16 -500 = 0xFE0C
	payload := []byte{0x0B, 0xF2, 0xFE, 0x0C, 0x02, 0x32, 0x00, 0x00, 0x00}

	v, err := lht65v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lht65v1.Data)
	if d.Temperature != -5.00 {
		t.Errorf("temperature = %v, want -5.00", d.Temperature)
	}
}

func TestNoConnection(t *testing.T) {
	payload := []byte{0x0B, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x81, 0x00, 0x00}

	v, err := lht65v1.Decode(decoders.Uplink{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(*lht65v1.Data)
	if d.SensorConnectionStatus == nil || *d.SensorConnectionStatus != "Sensor no connection" {
		t.Errorf("sensor_connection_status = %v", d.SensorConnectionStatus)
	}
}

func TestShortPayload(t *testing.T) {
	if _, err := lht65v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Error("want error on short payload")
	}
}

func TestRegistryLookup(t *testing.T) {
	// Case-insensitive lookup via registry (registered in init).
	payload := []byte{0x0B, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x00, 0x00, 0x00}
	if _, err := decoders.Decode("Dragino", "LHT65", "V1", decoders.Uplink{Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if _, err := decoders.Decode("nope", "nope", "v1", decoders.Uplink{Payload: payload}); err == nil {
		t.Error("want error for unregistered decoder")
	}
}
