package output

import "io"

type InputOutput interface {
	InOrStdin() io.Reader
	OutOrStdout() io.Writer
	ErrOrStderr() io.Writer
}
