package tui

import (
	"context"
	"fmt"

	"github.com/cardil/ghet/pkg/output"
	"github.com/erikgeiser/promptkit/selection"
)

type Chooser[T any] func(ctx context.Context, options []T, format string, a ...any) T

func BubbleChooser[T any](ctx context.Context, options []T, format string, a ...any) T {
	prt := output.PrinterFrom(ctx)
	l := output.LoggerFrom(ctx)
	sel := selection.New(fmt.Sprintf(format, a...), options)
	sel.PageSize = 3
	sel.Input = prt.InOrStdin()
	sel.Output = prt.OutOrStdout()
	chosen, err := sel.RunPrompt()
	if err != nil {
		l.Fatal(err)
	}
	return chosen
}
