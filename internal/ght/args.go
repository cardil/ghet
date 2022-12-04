package ght

import (
	"path"

	"github.com/cardil/ghet/pkg/metadata"
	"github.com/kirsle/configdir"
	"github.com/spf13/cobra"
)

type Args struct {
	ConfigPath string
}

func (a Args) Defaults() Args {
	configPath := configdir.LocalConfig(metadata.Name)
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		panic(err)
	}
	settingsPth := path.Join(configPath, "settings.yaml")

	return Args{
		ConfigPath: settingsPth,
	}
}

func (a *App) setFlags(c *cobra.Command) {
	defs := a.Defaults()
	fl := c.PersistentFlags()
	fl.StringVarP(&a.ConfigPath, "config", "c",
		defs.ConfigPath, "path to configuration file")
}
