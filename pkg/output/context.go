package output

import "context"

type printerKey struct{}

func PrinterFrom(ctx context.Context) Printer {
	p, ok := ctx.Value(printerKey{}).(Printer)
	if !ok {
		return defaultPrinter()
	}
	return p
}

func WithContext(ctx context.Context, p Printer) context.Context {
	return context.WithValue(ctx, printerKey{}, p)
}

func defaultPrinter() Printer {
	return OsPrinter
}
