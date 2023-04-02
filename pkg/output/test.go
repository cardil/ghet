package output

import (
	"bytes"
	"io"
	"strings"
)

func NewTestPrinter() TestPrinter {
	buf := bytes.NewBufferString("")
	return NewTestPrinterWithInput(buf)
}

func NewTestPrinterWithInput(input io.Reader) TestPrinter {
	return TestPrinter{
		stdPrinter{
			testInOut{
				in: input,
			},
		},
	}
}

func NewTestPrinterWithAnswers(answers []string) TestPrinter {
	return NewTestPrinterWithInput(bytes.NewBufferString(strings.Join(answers, "\n")))
}

type TestPrinter struct {
	stdPrinter
}

func (p TestPrinter) Outputs() TestOutputs {
	return p.InputOutput.(testInOut).TestOutputs //nolint:forcetypeassert
}

type TestOutputs struct {
	Out, Err bytes.Buffer
}

func (t TestOutputs) OutOrStdout() io.Writer {
	return &t.Out
}

func (t TestOutputs) ErrOrStderr() io.Writer {
	return &t.Err
}

type testInOut struct {
	in io.Reader
	TestOutputs
}

func (t testInOut) InOrStdin() io.Reader {
	return t.in
}

var _ InputOutput = testInOut{}
