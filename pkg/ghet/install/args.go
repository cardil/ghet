package install

import (
	"log"

	"dario.cat/mergo"
	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/github"
)

type Args struct {
	github.Asset
	config.Site
	MultipleBinaries bool
	VerifyInArchive  bool
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
		log.Fatal(err)
	}
	if a.BaseName == "" {
		a.BaseName = a.Repo
	}
	if a.OperatingSystem == github.OSWindows && a.Extension == "" {
		a.Extension = "exe"
	}
	return a
}
