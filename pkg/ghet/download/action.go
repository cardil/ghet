package download

import (
	"context"

	"knative.dev/client-pkg/pkg/output/logging"
)

func Action(ctx context.Context, args Args) error {
	ctx = logging.EnsureLogger(ctx, logging.Fields{
		"owner": args.Owner,
		"repo":  args.Repo,
	})
	plan, err := CreatePlan(ctx, args)
	if err != nil {
		return err
	}
	return plan.Download(ctx, args)
}
