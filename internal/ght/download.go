package ght

import (
	"context"
	"os"

	"github.com/cardil/ghet/pkg/ghet/download"
	"github.com/spf13/cobra"
)

type downloadArgs struct {
	installArgs
	destination string
}

func downloadCmd(args *Args) *cobra.Command {
	da := &downloadArgs{}
	c := &cobra.Command{
		Use:               "download [flags] <owner>/<repo>",
		Short:             "Download an artifact from GitHub release",
		PersistentPreRunE: da.valiadate(),
		RunE:              handle(args, downloadAction(da)),
		Example:           "\n * ght download -v 0.1.0 -t /tmp cardil/ghet",
	}
	da.setFlags(c)
	return c
}

func (da *downloadArgs) setFlags(c *cobra.Command) {
	fl := c.Flags()
	wd, _ := os.Getwd()
	fl.StringVarP(&da.destination, "destination", "d",
		wd, "a destination directory to download asset to")
	da.installArgs.setFlags(c)
}

func downloadAction(da *downloadArgs) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return download.Action(ctx, da.parse(ctx))
	}
}

func (da *downloadArgs) parse(ctx context.Context) download.Args {
	return download.Args{
		Args:        da.installArgs.parse(ctx),
		Destination: da.destination,
	}
}
