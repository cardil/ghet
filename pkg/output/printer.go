package output

import "fmt"

type Printer interface {
	Print(i ...any)
	Println(i ...any)
	Printf(format string, i ...any)

	PrintErr(i ...any)
	PrintErrln(i ...any)
	PrintErrf(format string, i ...any)

	InputOutput
}

type stdPrinter struct {
	InputOutput
}

func (p stdPrinter) Print(i ...any) {
	_, _ = fmt.Fprint(p.OutOrStdout(), i...)
}

func (p stdPrinter) Println(i ...any) {
	p.Print(fmt.Sprintln(i...))
}

func (p stdPrinter) Printf(format string, i ...any) {
	p.Print(fmt.Sprintf(format, i...))
}

func (p stdPrinter) PrintErr(i ...any) {
	_, _ = fmt.Fprint(p.ErrOrStderr(), i...)
}

func (p stdPrinter) PrintErrln(i ...any) {
	p.PrintErr(fmt.Sprintln(i...))
}

func (p stdPrinter) PrintErrf(format string, i ...any) {
	p.PrintErr(fmt.Sprintf(format, i...))
}

var _ Printer = stdPrinter{}
