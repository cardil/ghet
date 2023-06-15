package ght

import (
	"context"
	"path"

	configdir "github.com/cardil/ghet/pkg/config/dir"
	"github.com/spf13/cobra"
)

type Args struct {
	ConfigPath string
}

func (a Args) Defaults(ctx context.Context) Args {
	configPath := configdir.Config(ctx)
	settingsPth := path.Join(configPath, "settings.yaml")

	return Args{
		ConfigPath: settingsPth,
	}
}

func (a *App) setFlags(c *cobra.Command) {
	defs := a.Defaults(c.Context())
	fl := c.PersistentFlags()
	fl.StringVarP(&a.ConfigPath, "config", "c",
		defs.ConfigPath, "path to configuration file")
}
