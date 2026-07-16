// Package lsn50v2s31v1 decodes Dragino LSN50v2-S31 v1 uplinks.
package lsn50v2s31v1

import (
	decoders "github.com/oharkins/go-lora-device-decoders"
	"github.com/oharkins/go-lora-device-decoders/internal/lsn50v2"
)

func init() {
	decoders.Register("dragino", "lsn50v2-s31", "v1", decoders.New(Decode, lsn50v2.Offers()...))
}

type Data = lsn50v2.Data

func Decode(u decoders.Uplink) (any, error) {
	return lsn50v2.Decode("lsn50v2s31v1", u)
}
