package main

import (
	"github.com/cardil/ghet/internal/ght"
	"github.com/wavesoftware/go-commandline"
)

func main() {
	commandline.New(new(ght.App)).ExecuteOrDie(ght.Options...)
}

// Main is used for testing purposes.
func Main() { //nolint:deadcode
	main()
}
