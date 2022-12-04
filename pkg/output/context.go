package output

import "context"

var printerKey = struct{}{}

var defaultPrinter = OsPrinter

func FromContext(ctx context.Context) Printer {
	p, ok := ctx.Value(printerKey).(Printer)
	if !ok {
		return defaultPrinter
	}
	return p
}

func WithContext(ctx context.Context, p Printer) context.Context {
	return context.WithValue(ctx, printerKey, p)
}
