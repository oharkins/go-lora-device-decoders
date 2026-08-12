package em300v1_test

import (
	"encoding/hex"
	"errors"
	"math"
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/milesight/em300v1"
)

func decode(t *testing.T, m em300v1.Model, hexPayload string) any {
	t.Helper()
	b, err := hex.DecodeString(hexPayload)
	if err != nil {
		t.Fatal(err)
	}
	v, err := em300v1.Decode(m, decoders.Uplink{FPort: 85, Payload: b})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func data(t *testing.T, m em300v1.Model, hexPayload string) *em300v1.Data {
	t.Helper()
	v := decode(t, m, hexPayload)
	d, ok := v.(*em300v1.Data)
	if !ok {
		t.Fatalf("got %T, want *Data", v)
	}
	return d
}

// User guide 5.2.1: EM300-MCS periodic packet.
func TestMCSPeriodic(t *testing.T) {
	d := data(t, em300v1.ModelMCS, "03671001046871060000")
	if d.Temperature == nil || *d.Temperature != 27.2 {
		t.Errorf("temperature = %v, want 27.2", d.Temperature)
	}
	if d.Humidity == nil || *d.Humidity != 56.5 {
		t.Errorf("humidity = %v, want 56.5", d.Humidity)
	}
	if d.MagnetStatus == nil || *d.MagnetStatus != 0 {
		t.Errorf("magnet_status = %v, want 0", d.MagnetStatus)
	}
	if decoders.KindOf(d) != decoders.KindTelemetry {
		t.Errorf("KindOf = %q, want telemetry", decoders.KindOf(d))
	}

	want := []decoders.Measurement{
		decoders.Float(decoders.Temperature, decoders.Celsius, 27.2),
		decoders.Float(decoders.Humidity, decoders.Percent, 56.5),
		decoders.Int(decoders.DoorOpenStatus, "", 0),
	}
	got := d.Measurements()
	if len(got) != len(want) {
		t.Fatalf("Measurements() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Measurements()[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

// User guide 5.2.1: EM300-MLD periodic packet.
func TestMLDPeriodic(t *testing.T) {
	d := data(t, em300v1.ModelMLD, "050000")
	if d.WaterLeak == nil || *d.WaterLeak != 0 {
		t.Errorf("water_leak = %v, want 0", d.WaterLeak)
	}
}

// User guide 5.2.1: battery level packet.
func TestBattery(t *testing.T) {
	d := data(t, em300v1.ModelTH, "017564")
	if d.Battery == nil || *d.Battery != 100 {
		t.Errorf("battery = %v, want 100", d.Battery)
	}
}

// User guide 5.2.1: water leakage change packet.
func TestLeakChange(t *testing.T) {
	d := data(t, em300v1.ModelZLD, "03671001046871050001")
	if d.WaterLeak == nil || *d.WaterLeak != 1 {
		t.Errorf("water_leak = %v, want 1", d.WaterLeak)
	}
}

func TestNegativeTemperature(t *testing.T) {
	// -101 = 0xFF9B little-endian => -10.1 °C
	d := data(t, em300v1.ModelTH, "03679bff")
	if d.Temperature == nil || *d.Temperature != -10.1 {
		t.Errorf("temperature = %v, want -10.1", d.Temperature)
	}
}

// User guide 5.1: basic information example.
func TestDeviceInfo(t *testing.T) {
	v := decode(t, em300v1.ModelTH, "ff0bffff0101ff166136c40091605408ff090300ff0a0101ff0f00")
	info, ok := v.(*em300v1.DeviceInfo)
	if !ok {
		t.Fatalf("got %T, want *DeviceInfo", v)
	}
	if decoders.KindOf(v) != decoders.KindDeviceInfo {
		t.Fatalf("KindOf = %q, want device_info", decoders.KindOf(v))
	}
	if !info.PowerOn {
		t.Error("power_on = false, want true")
	}
	if info.ProtocolVersion != "V1" {
		t.Errorf("protocol_version = %q, want V1", info.ProtocolVersion)
	}
	if info.SerialNumber != "6136c40091605408" {
		t.Errorf("sn = %q, want 6136c40091605408", info.SerialNumber)
	}
	if info.HardwareVersion != "V3.0" {
		t.Errorf("hardware_version = %q, want V3.0", info.HardwareVersion)
	}
	if info.SoftwareVersion != "V1.1" {
		t.Errorf("software_version = %q, want V1.1", info.SoftwareVersion)
	}
	if info.DeviceClass != "Class A" {
		t.Errorf("device_class = %q, want Class A", info.DeviceClass)
	}
}

// User guide 5.2.2: EM300-DI periodic packet (counter mode).
func TestDICounter(t *testing.T) {
	d := data(t, em300v1.ModelDI, "03671e0104689405e10a000a0000005b43")
	if d.Temperature == nil || *d.Temperature != 28.6 {
		t.Errorf("temperature = %v, want 28.6", d.Temperature)
	}
	if d.Humidity == nil || *d.Humidity != 74 {
		t.Errorf("humidity = %v, want 74", d.Humidity)
	}
	if d.WaterConv == nil || *d.WaterConv != 1 {
		t.Errorf("water_conv = %v, want 1", d.WaterConv)
	}
	if d.PulseConv == nil || *d.PulseConv != 1 {
		t.Errorf("pulse_conv = %v, want 1", d.PulseConv)
	}
	if d.WaterConsumption == nil || *d.WaterConsumption != 219 {
		t.Errorf("water_consumption = %v, want 219", d.WaterConsumption)
	}
}

// User guide 5.2.2: EM300-DI periodic packet (digital mode). Channel 05/00 is
// digital input on the DI, not water leak.
func TestDIDigital(t *testing.T) {
	d := data(t, em300v1.ModelDI, "03671e01046894050001")
	if d.DigitalInput == nil || *d.DigitalInput != 1 {
		t.Errorf("digital_input = %v, want 1", d.DigitalInput)
	}
	if d.WaterLeak != nil {
		t.Errorf("water_leak = %v, want nil on EM300-DI", d.WaterLeak)
	}
}

// User guide 5.2.2: pulse alarm packet.
func TestDIPulseAlarm(t *testing.T) {
	d := data(t, em300v1.ModelDI, "85e10a000a0000005b4301")
	if d.WaterConsumption == nil || *d.WaterConsumption != 219 {
		t.Errorf("water_consumption = %v, want 219", d.WaterConsumption)
	}
	if d.Alarm == nil || *d.Alarm != 1 {
		t.Errorf("alarm = %v, want 1", d.Alarm)
	}
	if d.AlarmType == nil || *d.AlarmType != "water_outage_timeout_alarm" {
		t.Errorf("alarm_type = %v, want water_outage_timeout_alarm", d.AlarmType)
	}
}

// Legacy pulse counter channel (firmware <= V1.2).
func TestDILegacyPulseCounter(t *testing.T) {
	d := data(t, em300v1.ModelDI, "05c8db000000")
	if d.PulseCount == nil || *d.PulseCount != 219 {
		t.Errorf("pulse_count = %v, want 219", d.PulseCount)
	}
}

// User guide 5.4: EM300-DI historical data example.
func TestDIHistory(t *testing.T) {
	v := decode(t, em300v1.ModelDI, "21ce0d755b630801570002000a0064003333af41")
	dl, ok := v.(*em300v1.DatalogData)
	if !ok {
		t.Fatalf("got %T, want *DatalogData", v)
	}
	if decoders.KindOf(v) != decoders.KindDatalog {
		t.Fatalf("KindOf = %q, want datalog", decoders.KindOf(v))
	}
	if len(dl.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(dl.Records))
	}
	rec := dl.Records[0]
	if rec.Timestamp != 1666938125 { // 2022/10/28 14:22:05 UTC
		t.Errorf("timestamp = %d, want 1666938125", rec.Timestamp)
	}
	if rec.Temperature == nil || *rec.Temperature != 26.4 {
		t.Errorf("temperature = %v, want 26.4", rec.Temperature)
	}
	if rec.Humidity == nil || *rec.Humidity != 43.5 {
		t.Errorf("humidity = %v, want 43.5", rec.Humidity)
	}
	if rec.AlarmType != nil {
		t.Errorf("alarm_type = %v, want nil (no alarm)", rec.AlarmType)
	}
	if rec.InterfaceType == nil || *rec.InterfaceType != "counter" {
		t.Errorf("interface_type = %v, want counter", rec.InterfaceType)
	}
	if rec.WaterConv == nil || *rec.WaterConv != 1 {
		t.Errorf("water_conv = %v, want 1", rec.WaterConv)
	}
	if rec.PulseConv == nil || *rec.PulseConv != 10 {
		t.Errorf("pulse_conv = %v, want 10", rec.PulseConv)
	}
	want := float64(math.Float32frombits(0x41af3333)) // 21.9
	if rec.WaterConsumption == nil || *rec.WaterConsumption != want {
		t.Errorf("water_consumption = %v, want %v", rec.WaterConsumption, want)
	}
}

// Official Milesight decoder example: telemetry plus a 20ce historical record
// in one uplink (https://github.com/Milesight-IoT/SensorDecoders/tree/main/em-series/em300-di).
func TestDIOfficialExample(t *testing.T) {
	d := data(t, em300v1.ModelDI, "01755c0367340104686520ce9e74466310015d020000010000")
	if d.Battery == nil || *d.Battery != 92 {
		t.Errorf("battery = %v, want 92", d.Battery)
	}
	if d.Temperature == nil || *d.Temperature != 30.8 {
		t.Errorf("temperature = %v, want 30.8", d.Temperature)
	}
	if d.Humidity == nil || *d.Humidity != 50.5 {
		t.Errorf("humidity = %v, want 50.5", d.Humidity)
	}
	if len(d.History) != 1 {
		t.Fatalf("history = %d, want 1", len(d.History))
	}
	rec := d.History[0]
	if rec.Timestamp != 1665561758 {
		t.Errorf("timestamp = %d, want 1665561758", rec.Timestamp)
	}
	if rec.Temperature == nil || *rec.Temperature != 27.2 {
		t.Errorf("history temperature = %v, want 27.2", rec.Temperature)
	}
	if rec.Humidity == nil || *rec.Humidity != 46.5 {
		t.Errorf("history humidity = %v, want 46.5", rec.Humidity)
	}
	if rec.InterfaceType == nil || *rec.InterfaceType != "counter" {
		t.Errorf("interface_type = %v, want counter", rec.InterfaceType)
	}
	if rec.PulseCount == nil || *rec.PulseCount != 256 {
		t.Errorf("pulse_count = %v, want 256", rec.PulseCount)
	}
	if rec.DigitalInput != nil {
		t.Errorf("digital_input = %v, want nil in counter mode", rec.DigitalInput)
	}
}

// Official 20ce layout: gpio_type 1 = gpio at offset 8, pulse unused.
func TestDIHistoryGPIO(t *testing.T) {
	v := decode(t, em300v1.ModelDI, "20ce9e74466310015d010100000000")
	dl, ok := v.(*em300v1.DatalogData)
	if !ok {
		t.Fatalf("got %T, want *DatalogData", v)
	}
	if len(dl.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(dl.Records))
	}
	rec := dl.Records[0]
	if rec.InterfaceType == nil || *rec.InterfaceType != "digital" {
		t.Errorf("interface_type = %v, want digital", rec.InterfaceType)
	}
	if rec.DigitalInput == nil || *rec.DigitalInput != 1 {
		t.Errorf("digital_input = %v, want 1", rec.DigitalInput)
	}
	if rec.PulseCount != nil {
		t.Errorf("pulse_count = %v, want nil in gpio mode", rec.PulseCount)
	}
}

// Official 21ce gpio mode: type 1 populates gpio, not water conversion fields.
func TestDIHistoryV2GPIO(t *testing.T) {
	v := decode(t, em300v1.ModelDI, "21ce0d755b630801570001010000000000000000")
	dl, ok := v.(*em300v1.DatalogData)
	if !ok {
		t.Fatalf("got %T, want *DatalogData", v)
	}
	rec := dl.Records[0]
	if rec.InterfaceType == nil || *rec.InterfaceType != "digital" {
		t.Errorf("interface_type = %v, want digital", rec.InterfaceType)
	}
	if rec.DigitalInput == nil || *rec.DigitalInput != 1 {
		t.Errorf("digital_input = %v, want 1", rec.DigitalInput)
	}
	if rec.WaterConsumption != nil {
		t.Errorf("water_consumption = %v, want nil in gpio mode", rec.WaterConsumption)
	}
}

// User guide 5.2.2 / official decoder: GPIO alarm packet.
func TestDIDigitalAlarm(t *testing.T) {
	d := data(t, em300v1.ModelDI, "85000101")
	if d.DigitalInput == nil || *d.DigitalInput != 1 {
		t.Errorf("digital_input = %v, want 1", d.DigitalInput)
	}
	if d.Alarm == nil || *d.Alarm != 1 {
		t.Errorf("alarm = %v, want 1", d.Alarm)
	}
	if d.AlarmType == nil || *d.AlarmType != "digital_input_alarm" {
		t.Errorf("alarm_type = %v, want digital_input_alarm", d.AlarmType)
	}
}

func TestDeviceInfoTSLAndReset(t *testing.T) {
	v := decode(t, em300v1.ModelDI, "ffff0102fffe01ff0f03")
	info, ok := v.(*em300v1.DeviceInfo)
	if !ok {
		t.Fatalf("got %T, want *DeviceInfo", v)
	}
	if info.TslVersion != "V1.2" {
		t.Errorf("tsl_version = %q, want V1.2", info.TslVersion)
	}
	if !info.Reset {
		t.Error("reset = false, want true")
	}
	if info.DeviceClass != "Class CtoB" {
		t.Errorf("device_class = %q, want Class CtoB", info.DeviceClass)
	}
}

// Historical data for the TH: two 20ce records in one uplink.
func TestTHHistory(t *testing.T) {
	v := decode(t, em300v1.ModelTH, "20ce0d755b6308015720ce0d755b63080157")
	dl, ok := v.(*em300v1.DatalogData)
	if !ok {
		t.Fatalf("got %T, want *DatalogData", v)
	}
	if len(dl.Records) != 2 {
		t.Fatalf("records = %d, want 2", len(dl.Records))
	}
	rec := dl.Records[0]
	if rec.Temperature == nil || *rec.Temperature != 26.4 {
		t.Errorf("temperature = %v, want 26.4", rec.Temperature)
	}
	if rec.Humidity == nil || *rec.Humidity != 43.5 {
		t.Errorf("humidity = %v, want 43.5", rec.Humidity)
	}
}

// Historical data for the SLD includes the leak status byte.
func TestSLDHistory(t *testing.T) {
	v := decode(t, em300v1.ModelSLD, "20ce0d755b6308015701")
	dl := v.(*em300v1.DatalogData)
	if len(dl.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(dl.Records))
	}
	if dl.Records[0].WaterLeak == nil || *dl.Records[0].WaterLeak != 1 {
		t.Errorf("water_leak = %v, want 1", dl.Records[0].WaterLeak)
	}
}

// User guide 5.2.3: EM300-CL periodic packet.
func TestCLPeriodic(t *testing.T) {
	d := data(t, em300v1.ModelCL, "01756403ed01")
	if d.Battery == nil || *d.Battery != 100 {
		t.Errorf("battery = %v, want 100", d.Battery)
	}
	if d.LiquidStatus == nil || *d.LiquidStatus != 1 {
		t.Errorf("liquid_status = %v, want 1 (full)", d.LiquidStatus)
	}
}

// User guide 5.2.3: EM300-CL liquid level alarm (status + alarm flag).
func TestCLAlarm(t *testing.T) {
	d := data(t, em300v1.ModelCL, "83ed0201")
	if d.LiquidStatus == nil || *d.LiquidStatus != 2 {
		t.Errorf("liquid_status = %v, want 2 (empty)", d.LiquidStatus)
	}
	if d.Alarm == nil || *d.Alarm != 1 {
		t.Errorf("alarm = %v, want 1", d.Alarm)
	}
}

// Liquid status 0xff means sensor error: the measurement must be invalid.
func TestCLSensorError(t *testing.T) {
	d := data(t, em300v1.ModelCL, "03edff")
	ms := d.Measurements()
	if len(ms) != 1 {
		t.Fatalf("Measurements() = %#v, want 1 entry", ms)
	}
	if ms[0].Name != decoders.LiquidStatus || ms[0].Valid || ms[0].Quality != decoders.QualityFault {
		t.Errorf("liquid_status measurement = %#v, want invalid with quality fault", ms[0])
	}
}

// Data enquiry replies (fc6b/fc6c) are not telemetry.
func TestEnquiryReplyIgnored(t *testing.T) {
	b, _ := hex.DecodeString("fc6c00")
	_, err := em300v1.Decode(em300v1.ModelTH, decoders.Uplink{FPort: 85, Payload: b})
	if !errors.Is(err, decoders.ErrIgnored) {
		t.Fatalf("want ErrIgnored, got %v", err)
	}
}

func TestShortPayload(t *testing.T) {
	if _, err := em300v1.Decode(em300v1.ModelTH, decoders.Uplink{Payload: []byte{0x01, 0x75}}); err == nil {
		t.Error("want error on short payload")
	}
}

func TestTruncatedChannel(t *testing.T) {
	// Temperature channel with only one data byte.
	b, _ := hex.DecodeString("036710")
	if _, err := em300v1.Decode(em300v1.ModelTH, decoders.Uplink{Payload: b}); err == nil {
		t.Error("want error on truncated channel")
	}
}

func TestUnknownChannel(t *testing.T) {
	b, _ := hex.DecodeString("0299ffff")
	if _, err := em300v1.Decode(em300v1.ModelTH, decoders.Uplink{Payload: b}); err == nil {
		t.Error("want error on unknown channel")
	}
}

func TestRegistryLookup(t *testing.T) {
	models := []string{"em300-th", "em300-mcs", "em300-sld", "em300-zld", "em300-mld", "em300-di", "em300-cl"}
	for _, m := range models {
		d, ok := decoders.Get("milesight", m, "v1")
		if !ok {
			t.Errorf("milesight/%s/v1 not registered", m)
			continue
		}
		if len(d.Offers()) == 0 {
			t.Errorf("milesight/%s/v1 declares no offerings", m)
		}
	}
	b, _ := hex.DecodeString("017564")
	if _, err := decoders.Decode("Milesight", "EM300-TH", "V1", decoders.Uplink{FPort: 85, Payload: b}); err != nil {
		t.Fatal(err)
	}
}
