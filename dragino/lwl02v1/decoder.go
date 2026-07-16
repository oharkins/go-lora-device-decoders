// Package lwl02v1 decodes Dragino LWL02 v1 uplinks.
package lwl02v1

import (
	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/internal/doorleak"
)

func init() {
	decoders.Register("dragino", "lwl02", "v1", decoders.New(Decode, doorleak.Offers()...))
}

type Data = doorleak.Data

func Decode(u decoders.Uplink) (any, error) {
	return doorleak.Decode("lwl02v1", u)
}
