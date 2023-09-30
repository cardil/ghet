package download

import (
	"context"

	"github.com/cardil/ghet/pkg/output"
	"github.com/cardil/ghet/pkg/output/tui"
	slog "github.com/go-eden/slf4go"
)

func Action(ctx context.Context, args Args) error {
	ctx = output.EnsureLogger(ctx, slog.Fields{
		"owner": args.Owner,
		"repo":  args.Repo,
	})
	ctx = tui.EnsureWidgets(ctx)
	plan, err := CreatePlan(ctx, args)
	if err != nil {
		return err
	}
	return plan.Download(ctx, args)
}
