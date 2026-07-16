# go-lora-device-decoders

Go library of LoRaWAN payload decoders, keyed by **manufacturer / product / version**.

## Install

```
go get github.com/oharkins/go-lora-device-decoders
```

## Usage

```go
import (
    "errors"

    decoders "github.com/oharkins/go-lora-device-decoders"
    _ "github.com/oharkins/go-lora-device-decoders/all" // register all decoders
)

func handle(fport uint8, payload []byte) error {
    v, err := decoders.Decode("dragino", "lht65", "v1", decoders.Uplink{
        FPort:   fport,
        Payload: payload,
    })
    if errors.Is(err, decoders.ErrIgnored) {
        return nil // config/status frame — not telemetry
    }
    if err != nil {
        return err
    }
    // Uniform pipeline view:
    if ms, ok := decoders.MeasurementsOf(v); ok {
        for _, m := range ms {
            _ = m // Name, Value, Unit
        }
    }
    // Or type-assert to *lht65v1.Data for device-specific fields
    _ = v
    return nil
}
```

Import a single decoder instead of `all` to keep binaries lean:

```go
import _ "github.com/oharkins/go-lora-device-decoders/dragino/lht65v1"
```

Dynamic lookup (e.g. device type from a device registry):

```go
d, ok := decoders.Get(dev.Manufacturer, dev.Product, dev.Version)
offers, _ := decoders.Offers(dev.Manufacturer, dev.Product, dev.Version)
// offers lists measurement names/units this decoder can produce
_ = d
_ = offers
```

### Pipeline interfaces

Every registered decoder implements:

- `Decode(Uplink) (any, error)` — typed payload (or `ErrIgnored`)
- `Offers() []Offering` — measurements the device can produce (discoverable before decode)

Decoded telemetry should implement `Measured` (`Measurements() []Measurement`) so a pipeline can ingest readings without per-device type switches. JSON field names on `Data` stay device-native; measurement names aim for a shared vocabulary (e.g. `battery_voltage`).

### Configurable decoders (e.g. Vega TP-11)

Some decoders accept a range config to convert raw sensor output into engineering units:

```go
import (
    decoders "github.com/oharkins/go-lora-device-decoders"
    tp11    "github.com/oharkins/go-lora-device-decoders/vega/tp-11"
)

func init() {
    // Register a 0-5 m water level variant
    decoders.Register("vega", "tp-11-0-5m", "v1",
        tp11.NewDecoder(tp11.RangeConfig{MinVal: 0, MaxVal: 5, Unit: "m"}))

    // Register a 0-100 PSI pressure variant
    decoders.Register("vega", "tp-11-0-100psi", "v1",
        tp11.NewDecoder(tp11.RangeConfig{MinVal: 0, MaxVal: 100, Unit: "psi"}))
}
```

## Layout

```
decoders.go                  registry, Decoder interface, Uplink type
all/                         blank-imports every decoder
<manufacturer>/<product>/    one package per decoder, self-registers in init()
```

Registered key strings are lowercase: `dragino/lht65/v1`.

---

## Decoders

| Manufacturer | Product | Version | Package | Notes |
|---|---|---|---|---|
| Dragino | LHT65 | v1 | `dragino/lht65v1` | Temp/humidity + ext sensor |
| Dragino | LHT65N | v1 | `dragino/lht65nv1` | Updated LHT65, 14 ext sensor modes |
| Dragino | LHT65N-PIR | v1 | `dragino/lht65npirv1` | LHT65N + PIR motion |
| Dragino | LHT65N-VIB | v1 | `dragino/lht65nvibv1` | Vibration + accelerometer |
| Dragino | LAQ4 | v1 | `dragino/laq4v1` | Air quality (CO2, TVOC) |
| Dragino | LBT1 | v1 | `dragino/lbt1v1` | BLE beacon scanner |
| Dragino | LDDS04 | v1 | `dragino/ldds04v1` | 4-channel distance |
| Dragino | LDDS20 | v1 | `dragino/ldds20v1` | Distance (mm) |
| Dragino | LDDS45 | v1 | `dragino/ldds45v1` | Distance (mm) |
| Dragino | LDDS75 | v1 | `dragino/ldds75v1` | Distance (mm) |
| Dragino | LDS01 | v1 | `dragino/lds01v1` | Door / water leak |
| Dragino | LDS02 | v1 | `dragino/lds02v1` | Door / water leak |
| Dragino | LGT92 | v1 | `dragino/lgt92v1` | GPS tracker |
| Dragino | LLDS12 | v1 | `dragino/llds12v1` | LiDAR distance + DS18B20 |
| Dragino | LLMS01 | v1 | `dragino/llms01v1` | Leaf moisture + temp |
| Dragino | LSE01 | v1 | `dragino/lse01v1` | Soil moisture / temp / conductivity |
| Dragino | LSN50 | v1 | `dragino/lsn50v1` | Multi-mode sensor node |
| Dragino | LSN50v2-D20 | v1 | `dragino/lsn50v2d20v1` | Multi-mode sensor node v2 |
| Dragino | LSN50v2-D20-D22-D23 | v1 | `dragino/lsn50v2d20d22d23v1` | Triple DS18B20 variant |
| Dragino | LSN50v2-S31 | v1 | `dragino/lsn50v2s31v1` | Multi-mode sensor node v2 |
| Dragino | LSNPK01 | v1 | `dragino/lsnpk01v1` | Soil NPK + temp |
| Dragino | LSPH01 | v1 | `dragino/lsph01v1` | Soil pH + temp |
| Dragino | LT22222-L | v1 | `dragino/lt22222lv1` | I/O controller |
| Dragino | LT33222-L | v1 | `dragino/lt33222lv1` | I/O controller |
| Dragino | LTC2 | v1 | `dragino/ltc2v1` | Dual-channel temp / resistance |
| Dragino | LWL01 | v1 | `dragino/lwl01v1` | Door / water leak |
| Dragino | LWL02 | v1 | `dragino/lwl02v1` | Door / water leak |
| Vega | TP-11 | v1 | `vega/tp-11` | 4-20 mA transmitter (raw mA) |

---

## Contributing test payloads

The fastest way to improve decoder reliability is adding real-world captured payloads to the test files. Each decoder that has tests uses a shared `cases` slice at the top of its `decoder_test.go`. Adding a new sample is four steps:

### 1. Capture a payload

Get the base64-encoded payload from your network server (ChirpStack, TTN, Helium, etc.). It will look something like:

```
AS4A8renZByQAdAHAADqAw==
```

### 2. Decode it manually to find expected values

```bash
echo "AS4A8renZByQAdAHAADqAw==" | base64 -d | xxd
```

Use the decoder's field comments or the device datasheet to map bytes to values.
If you're not sure what the expected values should be, leave a comment in the PR and someone can help.

### 3. Add it to the `cases` slice

Open the decoder's `decoder_test.go` (e.g. `vega/tp-11/decoder_test.go`) and append to `cases`:

```go
var cases = []testCase{
    // existing entries ...
    {
        name:        "sample 3 — my device SN 0042",
        b64:         "AT8A9WXpZR9eAdAHAACEAw==",
        fport:       2,
        reason:      "Sending packet by the time",
        batteryPct:  63,
        temperature: 31,
        ma:          9.00,
        maLow:       3.50,
        maHigh:      20.00,
        value:       1.5625,
        valueLow:    -1,     // 3.50 mA < 4 mA → sensor fault
        valueHigh:   5.0,
    },
}
```

### 4. Run the tests

```bash
go test ./vega/tp-11/... -v
```

All subtests are named after the `name` field so failures are easy to trace back to the sample.

### Tips

- **More samples = better coverage.** Edge cases like near-zero readings, max-range readings, fault conditions (mA < 4), and negative temperatures are especially valuable.
- **Include context in `name`** — device serial number, site, or date helps future debugging.
- **Don't worry about getting every field right.** Submit what you have and open a PR; the CI will catch any mismatches.

---

## Adding a new decoder

1. Create `<manufacturer>/<product>/decoder.go`.
2. Define a typed `Data` struct with JSON tags; use pointer fields + `omitempty` for optional fields.
3. Implement `Decode(decoders.Uplink) (any, error)` with a minimum payload length check. Return `decoders.ErrIgnored` for non-telemetry frames.
4. Implement `Measurements() []decoders.Measurement` on telemetry payloads.
5. Register in `init()` with declared offerings:

   ```go
   func init() {
       decoders.Register("acme", "widget", "v1", decoders.New(Decode,
           decoders.Offer("battery_voltage", "V"),
           decoders.Offer("temperature", "C"),
       ))
   }
   ```

6. Add the blank import to `all/all.go`.
7. Add a `decoder_test.go` following the table-driven pattern above.
