package tui

import "context"

type Widgets struct {
	NewSpinner  NewSpinnerFunc
	NewProgress NewProgressFunc
	Printf      PrintfFunc
}

type widgetsKey struct{}

func EnsureWidgets(ctx context.Context) context.Context {
	return WithWidgets(ctx, defaultWidgets())
}

func WithWidgets(ctx context.Context, w *Widgets) context.Context {
	return context.WithValue(ctx, widgetsKey{}, w)
}

func WidgetsFrom(ctx context.Context) *Widgets {
	if w, ok := ctx.Value(widgetsKey{}).(*Widgets); ok {
		return w
	}
	return defaultWidgets()
}

func defaultWidgets() *Widgets {
	return &Widgets{
		NewSpinner:  NewBubbleSpinner,
		NewProgress: NewBubbleProgress,
		Printf:      FmtPrintfFunc,
	}
}
