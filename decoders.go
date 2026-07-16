// Package decoders provides a registry of LoRaWAN payload decoders keyed by
// manufacturer, product, and version.
//
// Pipeline consumers typically:
//  1. Look up a decoder (or its Offers) from device metadata
//  2. Call Decode on an uplink
//  3. Treat ErrIgnored as a soft skip
//  4. Route by KindOf(v), then read Measurements() for a uniform view
package decoders

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ErrIgnored means the uplink was valid but should not be treated as telemetry
// (config ack, poll status, empty/special frames, etc.).
var ErrIgnored = errors.New("decoders: uplink ignored")

// Uplink is the input to a decoder.
type Uplink struct {
	FPort   uint8
	Payload []byte
}

// Kind classifies a decoded uplink for pipeline routing.
type Kind string

const (
	KindTelemetry  Kind = "telemetry"
	KindDeviceInfo Kind = "device_info"
	KindDatalog    Kind = "datalog"
	KindAccel      Kind = "acceleration"
)

// Message is implemented by decoded payloads that declare their pipeline kind.
// If a value does not implement Message, KindOf treats it as KindTelemetry.
type Message interface {
	MessageKind() Kind
}

// Offering declares a measurement a decoder can produce. Name should be a
// stable, snake_case identifier from the canonical vocabulary (see names.go).
type Offering struct {
	Name string `json:"name"`
	Unit string `json:"unit,omitempty"`
}

// Offer is a convenience constructor for Offering.
func Offer(name, unit string) Offering {
	return Offering{Name: name, Unit: unit}
}

// Measurement is one normalized reading from a decoded uplink.
// Valid is true for usable readings; when false, Quality explains why.
type Measurement struct {
	Name    string  `json:"name"`
	Value   float64 `json:"value"`
	Unit    string  `json:"unit,omitempty"`
	Valid   bool    `json:"valid"`
	Quality string  `json:"quality,omitempty"`
}

// Common quality reasons for Valid == false.
const (
	QualityInvalid      = "invalid"
	QualityNoSensor     = "no_sensor"
	QualityFault        = "fault"
	QualityNoConnection = "no_connection"
	QualityOutOfRange   = "out_of_range"
)

// Measured is implemented by decoded payloads that expose pipeline-ready readings.
type Measured interface {
	Measurements() []Measurement
}

// Decoder decodes a raw uplink into a typed value and declares the measurements
// it can offer. The returned value must be JSON-marshalable; prefer implementing
// Measured and Message for pipeline consumers.
type Decoder interface {
	Decode(u Uplink) (any, error)
	Offers() []Offering
}

type staticDecoder struct {
	fn     func(Uplink) (any, error)
	offers []Offering
}

func (d staticDecoder) Decode(u Uplink) (any, error) { return d.fn(u) }
func (d staticDecoder) Offers() []Offering {
	out := make([]Offering, len(d.offers))
	copy(out, d.offers)
	return out
}

// New builds a Decoder from a decode function and the offerings it can produce.
// Every registered decoder should use New so pipelines can discover required
// (and optional) measurement names up front.
func New(fn func(Uplink) (any, error), offers ...Offering) Decoder {
	if fn == nil {
		panic("decoders: New called with nil decode function")
	}
	cp := make([]Offering, len(offers))
	copy(cp, offers)
	return staticDecoder{fn: fn, offers: cp}
}

// DecoderFunc adapts a bare function to Decoder with no declared offerings.
// Prefer New when registering decoders used in a sensor pipeline.
type DecoderFunc func(u Uplink) (any, error)

func (f DecoderFunc) Decode(u Uplink) (any, error) { return f(u) }
func (f DecoderFunc) Offers() []Offering            { return nil }

// Key identifies a decoder: manufacturer/product/version.
type Key struct {
	Manufacturer string
	Product      string
	Version      string
}

func (k Key) String() string {
	return fmt.Sprintf("%s/%s/%s", k.Manufacturer, k.Product, k.Version)
}

// ParseKey parses "manufacturer/product/version".
func ParseKey(s string) (Key, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return Key{}, fmt.Errorf("invalid key %q: want manufacturer/product/version", s)
	}
	return Key{normalize(parts[0]), normalize(parts[1]), normalize(parts[2])}, nil
}

func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

var (
	mu       sync.RWMutex
	registry = map[Key]Decoder{}
)

// Register adds a decoder. Panics on duplicate registration — intended to be
// called from package init().
func Register(manufacturer, product, version string, d Decoder) {
	if d == nil {
		panic("decoders: Register called with nil Decoder")
	}
	k := Key{normalize(manufacturer), normalize(product), normalize(version)}
	mu.Lock()
	defer mu.Unlock()
	if _, dup := registry[k]; dup {
		panic(fmt.Sprintf("decoders: duplicate registration for %s", k))
	}
	registry[k] = d
}

// Get returns the decoder for manufacturer/product/version.
func Get(manufacturer, product, version string) (Decoder, bool) {
	mu.RLock()
	defer mu.RUnlock()
	d, ok := registry[Key{normalize(manufacturer), normalize(product), normalize(version)}]
	return d, ok
}

// GetByKey is Get with a Key.
func GetByKey(k Key) (Decoder, bool) {
	return Get(k.Manufacturer, k.Product, k.Version)
}

// Offers returns the measurement offerings for a registered decoder.
func Offers(manufacturer, product, version string) ([]Offering, bool) {
	d, ok := Get(manufacturer, product, version)
	if !ok {
		return nil, false
	}
	return d.Offers(), true
}

// Decode is a convenience wrapper: look up and decode in one call.
func Decode(manufacturer, product, version string, u Uplink) (any, error) {
	d, ok := Get(manufacturer, product, version)
	if !ok {
		return nil, fmt.Errorf("decoders: no decoder registered for %s/%s/%s",
			normalize(manufacturer), normalize(product), normalize(version))
	}
	return d.Decode(u)
}

// List returns all registered keys, sorted.
func List() []Key {
	mu.RLock()
	defer mu.RUnlock()
	keys := make([]Key, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	return keys
}

// MeasurementsOf returns Measurements when v implements Measured.
func MeasurementsOf(v any) ([]Measurement, bool) {
	if m, ok := v.(Measured); ok {
		return m.Measurements(), true
	}
	return nil, false
}

// KindOf returns the message kind. Values that do not implement Message are
// treated as KindTelemetry.
func KindOf(v any) Kind {
	if m, ok := v.(Message); ok {
		return m.MessageKind()
	}
	return KindTelemetry
}

// AppendFloat appends a valid measurement when v is non-nil.
func AppendFloat(dst []Measurement, name, unit string, v *float64) []Measurement {
	if v == nil {
		return dst
	}
	return append(dst, Float(name, unit, *v))
}

// AppendInt appends a valid measurement when v is non-nil.
func AppendInt(dst []Measurement, name, unit string, v *int) []Measurement {
	if v == nil {
		return dst
	}
	return append(dst, Int(name, unit, *v))
}

// AppendInt64 appends a valid measurement when v is non-nil.
func AppendInt64(dst []Measurement, name, unit string, v *int64) []Measurement {
	if v == nil {
		return dst
	}
	return append(dst, Measurement{Name: name, Value: float64(*v), Unit: unit, Valid: true})
}

// Float is a helper for a valid required numeric field.
func Float(name, unit string, v float64) Measurement {
	return Measurement{Name: name, Value: v, Unit: unit, Valid: true}
}

// Int is a helper for a valid required integer field.
func Int(name, unit string, v int) Measurement {
	return Measurement{Name: name, Value: float64(v), Unit: unit, Valid: true}
}

// FloatQuality is a reading with an explicit validity/quality.
func FloatQuality(name, unit string, v float64, valid bool, quality string) Measurement {
	m := Measurement{Name: name, Value: v, Unit: unit, Valid: valid}
	if !valid {
		m.Quality = quality
	}
	return m
}
