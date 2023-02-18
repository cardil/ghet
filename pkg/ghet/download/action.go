package download

import (
	"context"

	"github.com/cardil/ghet/pkg/output"
	slog "github.com/go-eden/slf4go"
	"github.com/pkg/errors"
)

func Action(ctx context.Context, args Args) error {
	ctx = output.EnsureLogger(ctx, slog.Fields{
		"owner": args.Owner,
		"repo":  args.Repo,
	})
	pl, err := CreatePlan(ctx, args)
	if err != nil {
		return errors.WithStack(err)
	}
	if err = pl.Download(ctx, args); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
