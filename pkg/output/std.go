package output

import "io"

type StandardOutputs interface {
	OutOrStderr() io.Writer
	ErrOrStderr() io.Writer
}
