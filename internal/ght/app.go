package ght

import (
	"github.com/spf13/cobra"
	"github.com/wavesoftware/go-commandline"
)

// Options to override the commandline for testing purposes.
var Options []commandline.Option //nolint:gochecknoglobals

// App is a main Ght application.
type App struct{}

func (a *App) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "ght",
		Short: "Ghet artifacts from GitHub releases",
	}
}

var _ commandline.CobraProvider = new(App)
