package lht65v1_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/dragino/lht65v1"
)

func decode(t *testing.T, payload []byte) *lht65v1.Data {
	t.Helper()
	v, err := lht65v1.Decode(decoders.Uplink{FPort: 2, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	return v.(*lht65v1.Data)
}

func TestDecodeBuiltIn(t *testing.T) {
	// batt=3.058V, bits15-14=00 → "Ultra Low", temp=22.71°C, hum=56.2%,
	// sensorType=1 (Temperature Sensor), ext temp=25.00°C
	payload := []byte{0x0B, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x01, 0x09, 0xC4}
	d := decode(t, payload)

	if d.BatLevel != "Ultra Low" {
		t.Errorf("bat_level = %q, want \"Ultra Low\"", d.BatLevel)
	}
	if d.BatteryVoltage != 3.058 {
		t.Errorf("battery_voltage = %v, want 3.058", d.BatteryVoltage)
	}
	if d.Temperature != 22.71 {
		t.Errorf("temperature = %v, want 22.71", d.Temperature)
	}
	if d.Humidity != 56.2 {
		t.Errorf("humidity = %v, want 56.2", d.Humidity)
	}
	if d.SensorType != "Temperature Sensor" {
		t.Errorf("sensor_type = %q, want \"Temperature Sensor\"", d.SensorType)
	}
	if d.ExternalTemperature == nil || *d.ExternalTemperature != 25.00 {
		t.Errorf("external_temperature = %v, want 25.00", d.ExternalTemperature)
	}
}

func TestBatLevel(t *testing.T) {
	// The top 2 bits of the 16-bit battery word encode the battery level.
	// Voltage bits (0-13) are held constant at 0x0BF2 (3.058 V) across all cases.
	//   bits15-14 = 00 → b[0]=0x0B → "Ultra Low"
	//   bits15-14 = 01 → b[0]=0x4B → "Low"
	//   bits15-14 = 10 → b[0]=0x8B → "OK"
	//   bits15-14 = 11 → b[0]=0xCB → "Good"
	cases := []struct {
		b0   byte
		want string
	}{
		{0x0B, "Ultra Low"},
		{0x4B, "Low"},
		{0x8B, "OK"},
		{0xCB, "Good"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			payload := []byte{tc.b0, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x00, 0x00, 0x00}
			d := decode(t, payload)
			if d.BatLevel != tc.want {
				t.Errorf("bat_level = %q, want %q", d.BatLevel, tc.want)
			}
			if d.BatteryVoltage != 3.058 {
				t.Errorf("battery_voltage = %v, want 3.058 (bat level bits must not affect voltage)", d.BatteryVoltage)
			}
		})
	}
}

func TestSensorType4_InterruptDoor(t *testing.T) {
	// Sensor type 4: Interrupt/Door Sensor send
	// b[7]=0x01 (High), b[8]=0x01 (True)
	payload := []byte{0x0B, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x04, 0x01, 0x01}
	d := decode(t, payload)

	if d.SensorType != "Interrupt/Door Sensor send" {
		t.Errorf("sensor_type = %q, want \"Interrupt/Door Sensor send\"", d.SensorType)
	}
	if d.InterruptPinLevel == nil || *d.InterruptPinLevel != "High" {
		t.Errorf("interrupt_pin_level = %v, want \"High\"", d.InterruptPinLevel)
	}
	if d.InterruptStatus == nil || *d.InterruptStatus != "True" {
		t.Errorf("interrupt_status = %v, want \"True\"", d.InterruptStatus)
	}

	// Low / False variant
	payload[7], payload[8] = 0x00, 0x00
	d = decode(t, payload)
	if *d.InterruptPinLevel != "Low" {
		t.Errorf("interrupt_pin_level = %v, want \"Low\"", *d.InterruptPinLevel)
	}
	if *d.InterruptStatus != "False" {
		t.Errorf("interrupt_status = %v, want \"False\"", *d.InterruptStatus)
	}
}

func TestNegativeTemperature(t *testing.T) {
	// temp = -5.00°C → int16 -500 = 0xFE0C
	payload := []byte{0x0B, 0xF2, 0xFE, 0x0C, 0x02, 0x32, 0x00, 0x00, 0x00}
	d := decode(t, payload)
	if d.Temperature != -5.00 {
		t.Errorf("temperature = %v, want -5.00", d.Temperature)
	}
}

func TestNoConnection(t *testing.T) {
	// bit 7 of b[6] set → sensor no connection
	payload := []byte{0x0B, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x81, 0x00, 0x00}
	d := decode(t, payload)
	if d.SensorConnectionStatus == nil || *d.SensorConnectionStatus != "Sensor no connection" {
		t.Errorf("sensor_connection_status = %v, want \"Sensor no connection\"", d.SensorConnectionStatus)
	}
}

func TestShortPayload(t *testing.T) {
	if _, err := lht65v1.Decode(decoders.Uplink{Payload: []byte{0x01}}); err == nil {
		t.Error("want error on short payload")
	}
}

func TestRegistryLookup(t *testing.T) {
	payload := []byte{0x0B, 0xF2, 0x08, 0xDF, 0x02, 0x32, 0x00, 0x00, 0x00}
	if _, err := decoders.Decode("Dragino", "LHT65", "V1", decoders.Uplink{Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if _, err := decoders.Decode("nope", "nope", "v1", decoders.Uplink{Payload: payload}); err == nil {
		t.Error("want error for unregistered decoder")
	}
}
