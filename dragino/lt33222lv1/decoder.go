// Package lt33222lv1 decodes Dragino LT33222-L v1 uplinks (I/O controller).
// The byte-level format is identical to the LT22222-L; hardware mode is
// reported inside the payload (byte 10 bits 7-6).
package lt33222lv1

import (
	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/internal/ltio"
)

func init() {
	decoders.Register("dragino", "lt33222-l", "v1", decoders.New(Decode, ltio.Offers()...))
}

type Data = ltio.Data

func Decode(u decoders.Uplink) (any, error) {
	return ltio.Decode("lt33222lv1", u)
}
