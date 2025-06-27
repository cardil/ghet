package download

import (
	"github.com/cardil/ghet/pkg/ghet/install"
)

type Download struct {
	install.Installation
	Destination string
}

// Args is a deprecated alias
//
// Deprecated: use Download instead.
type Args = Download
