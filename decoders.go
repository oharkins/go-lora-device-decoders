// Package decoders provides a registry of LoRaWAN payload decoders keyed by
// manufacturer, product, and version.
package decoders

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Uplink is the input to a decoder.
type Uplink struct {
	FPort   uint8
	Payload []byte
}

// Decoder decodes a raw uplink into a typed struct. The returned value must
// be JSON-marshalable.
type Decoder interface {
	Decode(u Uplink) (any, error)
}

// DecoderFunc adapts a function to the Decoder interface.
type DecoderFunc func(u Uplink) (any, error)

func (f DecoderFunc) Decode(u Uplink) (any, error) { return f(u) }

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
