package ght

import (
	"context"
	"os"

	"emperror.dev/errors"
	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/metadata"
	"github.com/spf13/cobra"
	"github.com/wavesoftware/go-commandline"
	"knative.dev/client-pkg/pkg/output"
	"knative.dev/client-pkg/pkg/output/logging"
)

// Options to override the commandline for testing purposes.
var Options []commandline.Option //nolint:gochecknoglobals

// App is a main Ght application.
type App struct {
	Args
}

func (a *App) Command() *cobra.Command {
	c := &cobra.Command{
		Use:          metadata.Name,
		Short:        "Gʰet artifacts from GitHub releases",
		SilenceUsage: true,
	}
	cmds := []func(*Args) *cobra.Command{
		versionCmd,
		installCmd,
		removeCmd,
		listCmd,
		downloadCmd,
	}
	for _, cmd := range cmds {
		c.AddCommand(cmd(&a.Args))
	}
	c.SetOut(os.Stdout)
	c.SetContext(logging.EnsureLogger(
		logging.EnsureLogFile(context.Background())),
	)
	a.setFlags(c)
	c.PostRunE = postRunE
	return c
}

func postRunE(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	logFile := logging.LogFileFrom(ctx)
	if err := logFile.Sync(); err != nil {
		return errors.WithStack(err)
	}
	if err := logFile.Close(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func handle(args *Args, fn func(ctx context.Context) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		ctx = output.WithContext(ctx, cmd)
		cfg, err := config.Load(ctx, args.ConfigPath)
		if err != nil {
			return err
		}
		ctx = config.WithContext(ctx, cfg)
		return fn(ctx)
	}
}

var _ commandline.CobraProvider = new(App)
