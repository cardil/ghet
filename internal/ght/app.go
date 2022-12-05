package ght

import (
	"context"
	"os"

	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/metadata"
	"github.com/cardil/ghet/pkg/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wavesoftware/go-commandline"
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
	a.setFlags(c)
	return c
}

func handle(args *Args, fn func(ctx context.Context) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		setupLogging(cmd)
		ctx := cmd.Context()
		ctx = output.WithContext(ctx, cmd)
		cfg, err := config.Load(args.ConfigPath)
		if err != nil {
			return err
		}
		ctx = config.WithContext(ctx, cfg)
		return fn(ctx)
	}
}

var _ commandline.CobraProvider = new(App)

func setupLogging(outs output.StandardOutputs) {
	log.SetOutput(outs.ErrOrStderr())
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		l, err := log.ParseLevel(lvl)
		if err != nil {
			log.WithError(err).Error("Failed to parse LOG_LEVEL")
		} else {
			log.SetLevel(l)
		}
	}
}
