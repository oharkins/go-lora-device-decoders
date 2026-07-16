// Package lds01v1 decodes Dragino LDS01 v1 uplinks.
package lds01v1

import (
	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/internal/doorleak"
)

func init() {
	decoders.Register("dragino", "lds01", "v1", decoders.New(Decode, doorleak.Offers()...))
}

type Data = doorleak.Data

func Decode(u decoders.Uplink) (any, error) {
	return doorleak.Decode("lds01v1", u)
}
