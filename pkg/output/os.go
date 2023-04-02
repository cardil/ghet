package output

import (
	"io"
	"os"
)

type OsInOut struct{}

var OsPrinter = stdPrinter{OsInOut{}} //nolint:gochecknoglobals

func (o OsInOut) InOrStdin() io.Reader {
	return os.Stdin
}

func (o OsInOut) OutOrStdout() io.Writer {
	return os.Stdout
}

func (o OsInOut) ErrOrStderr() io.Writer {
	return os.Stderr
}

var _ InputOutput = OsInOut{}
