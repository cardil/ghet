package output

import (
	"io"
	"os"
)

type OsOutputs struct{}

var OsPrinter = stdPrinter{OsOutputs{}} //nolint:gochecknoglobals

func (o OsOutputs) OutOrStderr() io.Writer {
	return os.Stdout
}

func (o OsOutputs) ErrOrStderr() io.Writer {
	return os.Stderr
}

var _ StandardOutputs = OsOutputs{}
