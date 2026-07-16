// Package lt22222lv1 decodes Dragino LT22222-L v1 uplinks (I/O controller).
package lt22222lv1

import (
	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/internal/ltio"
)

func init() {
	decoders.Register("dragino", "lt22222-l", "v1", decoders.New(Decode, ltio.Offers()...))
}

type Data = ltio.Data

func Decode(u decoders.Uplink) (any, error) {
	return ltio.Decode("lt22222lv1", u)
}
