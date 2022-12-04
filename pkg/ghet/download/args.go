package download

import "github.com/cardil/ghet/pkg/ghet/install"

type Args struct {
	install.Args
	Destination string
}
