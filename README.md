# lorawan-decoders

Go library of LoRaWAN payload decoders, keyed by **manufacturer / product / version**.

## Install

```
go get github.com/oharkins/go-lora-device-decoders
```

## Usage

```go
import (
    decoders "github.com/oharkins/go-lora-device-decoders"
    _ "github.com/oharkins/go-lora-device-decoders/all" // register all decoders
)

func handle(payload []byte) error {
    v, err := decoders.Decode("dragino", "lht65", "v1", decoders.Uplink{
        FPort:   2,
        Payload: payload,
    })
    if err != nil {
        return err
    }
    // v is *lht65v1.Data — JSON-marshalable, or type-assert for typed access
    _ = v
    return nil
}
```

Import a single decoder instead of `all` to keep binaries lean:

```go
import _ "github.com/oharkins/go-lora-device-decoders/dragino/lht65v1"
```

Dynamic lookup (e.g. device type from a DynamoDB device registry record):

```go
d, ok := decoders.Get(dev.Manufacturer, dev.Product, dev.FirmwareVersion)
```

## Layout

```
decoders.go              registry, Decoder interface, Uplink type
all/                     blank-imports every decoder
<manufacturer>/<product><version>/
                         one package per decoder, self-registers in init()
```

Registered key strings are lowercase: `dragino/lht65/v1`.

## Adding a decoder

1. Create `<manufacturer>/<product><version>/decoder.go`.
2. Define a typed `Data` struct with JSON tags; use pointer fields + `omitempty` for conditional fields.
3. Implement `Decode(decoders.Uplink) (any, error)` with a payload length check.
4. Register in `init()`:

   ```go
   func init() {
       decoders.Register("acme", "widget", "v2", decoders.DecoderFunc(Decode))
   }
   ```

5. Add the blank import to `all/all.go`.
6. Add a `decoder_test.go` with known-good payloads (including negative temps / sign extension if applicable).

## Decoders

| Manufacturer | Product | Version | Package |
|---|---|---|---|
| Dragino | LHT65 | v1 | `dragino/lht65v1` |
