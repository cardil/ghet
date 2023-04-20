package install

import (
	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/github"
	log "github.com/go-eden/slf4go"
	"github.com/imdario/mergo"
)

type Args struct {
	github.Asset
	config.Site
}

func (a Args) WithDefaults() Args {
	defs := Args{
		Asset: github.Asset{
			Architecture:    github.CurrentArchitecture(),
			OperatingSystem: github.CurrentOS(),
			Release: github.Release{
				Tag: "latest",
			},
		},
		Site: config.Site{
			Type:    config.TypeGitHub,
			Address: "github.com",
		},
	}
	if err := mergo.Merge(&a, defs); err != nil {
		log.Panic(err)
	}
	if a.Asset.FileName.BaseName == "" {
		a.Asset.FileName.BaseName = a.Repo
	}
	if a.OperatingSystem == github.OSWindows && a.Asset.FileName.Extension == "" {
		a.Asset.FileName.Extension = "exe"
	}
	return a
}
