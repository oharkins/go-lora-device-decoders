// Package all registers every decoder in the library via blank imports.
// Import it for side effects:
//
//	import _ "github.com/oharkins/go-lora-device-decoders/all"
package all

import (
	_ "github.com/oharkins/go-lora-device-decoders/dragino/lht65v1"
)
