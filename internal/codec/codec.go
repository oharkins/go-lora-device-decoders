// Package codec provides small binary helpers shared by device decoders.
package codec

import "math"

// Ptr returns a pointer to v.
func Ptr[T any](v T) *T { return &v }

// Round rounds v to the given number of decimal places.
func Round(v float64, places int) float64 {
	f := math.Pow10(places)
	return math.Round(v*f) / f
}

// U16BE reads a big-endian uint16 from b[i:].
func U16BE(b []byte, i int) uint16 {
	return uint16(b[i])<<8 | uint16(b[i+1])
}

// I16BE reads a big-endian int16 from b[i:] as float64.
func I16BE(b []byte, i int) float64 {
	return float64(int16(U16BE(b, i)))
}

// U32BE reads a big-endian uint32 from b[i:].
func U32BE(b []byte, i int) uint32 {
	return uint32(b[i])<<24 | uint32(b[i+1])<<16 | uint32(b[i+2])<<8 | uint32(b[i+3])
}

// BatV14 decodes Dragino-style battery voltage: lower 14 bits of a BE uint16, in volts.
func BatV14(b []byte, i int) float64 {
	return float64(U16BE(b, i)&0x3FFF) / 1000
}

// BatV16 decodes a full BE uint16 battery word as volts (/1000).
func BatV16(b []byte, i int) float64 {
	return float64(U16BE(b, i)) / 1000
}
