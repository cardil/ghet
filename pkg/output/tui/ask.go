package tui

import (
	"context"
	"fmt"

	"github.com/cardil/ghet/pkg/output"
	"github.com/erikgeiser/promptkit/selection"
)

type AskFunc func(ctx context.Context, options []string, format string, a ...any) string

func NewBubbleAsk(ctx context.Context, options []string, format string, a ...any) string {
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
