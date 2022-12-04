package output

import (
	"bytes"
	"io"
)

func NewTestPrinter() TestPrinter {
	return TestPrinter{
		stdPrinter{
			TestOutputs{},
		},
	}
}

type TestPrinter struct {
	stdPrinter
}

func (p TestPrinter) Outputs() TestOutputs {
	return p.StandardOutputs.(TestOutputs) //nolint:forcetypeassert
}

type TestOutputs struct {
	Out, Err bytes.Buffer
}

func (t TestOutputs) OutOrStderr() io.Writer {
	return &t.Out
}

func (t TestOutputs) ErrOrStderr() io.Writer {
	return &t.Err
}

var _ StandardOutputs = TestOutputs{}
