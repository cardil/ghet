package install

import (
	"log"

	"dario.cat/mergo"
	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/github"
)

type Installation struct {
	github.Asset
	config.Site
	MultipleBinaries bool
	VerifyInArchive  bool
}

func (a Installation) WithDefaults() Installation {
	defs := Installation{
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
	if a.Asset.FileName.BaseName == "" {
		a.Asset.FileName.BaseName = a.Repo
	}
	if a.OperatingSystem == github.OSWindows && a.Asset.FileName.Extension == "" {
		a.Asset.FileName.Extension = "exe"
	}
	return a
}
