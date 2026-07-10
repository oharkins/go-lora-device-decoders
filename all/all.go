// Package all registers every decoder in the library via blank imports.
// Import it for side effects:
//
//	import _ "github.com/odis/lorawan-decoders/all"
package all

import (
	_ "github.com/odis/lorawan-decoders/dragino/lht65v1"
)
