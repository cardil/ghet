package download

import (
	"context"

	"knative.dev/client/pkg/output/logging"
)

func Action(ctx context.Context, download Download) error {
	ctx = logging.EnsureLogger(ctx, logging.Fields{
		"owner": download.Owner,
		"repo":  download.Repo,
	})
	plan, err := CreatePlan(ctx, download)
	if err != nil {
		return err
	}
	return plan.Download(ctx, download)
}
