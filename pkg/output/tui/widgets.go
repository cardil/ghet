package tui

import (
	"context"

	"emperror.dev/errors"
	"github.com/cardil/ghet/pkg/output"
)

// ErrNotInteractive is returned when the user is not in an interactive session.
var ErrNotInteractive = errors.New("not interactive session")

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

func Interactive[T any](ctx context.Context) (*InteractiveWidgets[T], error) {
	prt := output.PrinterFrom(ctx)
	if !output.IsTerminal(prt.InOrStdin()) {
		return nil, errors.WithStack(ErrNotInteractive)
	}

	return &InteractiveWidgets[T]{
		Choose: BubbleChooser[T],
	}, nil
}

type InteractiveWidgets[T any] struct {
	Choose Chooser[T]
}
