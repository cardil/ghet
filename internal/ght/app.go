package ght

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wavesoftware/go-commandline"
)

// Options to override the commandline for testing purposes.
var Options []commandline.Option //nolint:gochecknoglobals

// App is a main Ght application.
type App struct{}

func (a *App) Command() *cobra.Command {
	c := &cobra.Command{
		Use:          "ght",
		Short:        "Gʰet artifacts from GitHub releases",
		SilenceUsage: true,
	}
	cmds := []func() *cobra.Command{
		versionCmd,
		installCmd,
		removeCmd,
		listCmd,
	}
	for _, cmd := range cmds {
		c.AddCommand(cmd())
	}
	c.SetOut(os.Stdout)
	return c
}

var _ commandline.CobraProvider = new(App)
