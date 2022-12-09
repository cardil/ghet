package install

import (
	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/github"
)

type Args struct {
	github.Asset
	config.Site
}

func (a Args) WithDefaults() Args {
	if a.Tag == "" {
		a.Tag = "latest"
	}
	if a.Site.Address == "" {
		a.Site.Address = "github.com"
	}
	if a.Site.Type == "" {
		a.Site.Type = config.TypeGitHub
	}
	if a.Asset.FileName.BaseName == "" {
		a.Asset.FileName.BaseName = a.Repo
	}
	if a.Architecture == "" {
		a.Architecture = github.CurrentArchitecture()
	}
	if a.OperatingSystem == "" {
		a.OperatingSystem = github.CurrentOS()
	}
	if a.Checksums.FileName.BaseName == "" {
		a.Checksums.FileName.BaseName = "checksums"
	}
	if a.Checksums.FileName.Extension == "" {
		a.Checksums.FileName.Extension = "txt"
	}
	if a.OperatingSystem == github.OSWindows && a.Asset.FileName.Extension == "" {
		a.Asset.FileName.Extension = "exe"
	}
	return a
}
