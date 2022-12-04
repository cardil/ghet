//go:build mage

package main

import (
	"github.com/cardil/ghet/pkg/metadata"

	// mage:import
	"github.com/wavesoftware/go-magetasks"
	"github.com/wavesoftware/go-magetasks/config"
	"github.com/wavesoftware/go-magetasks/pkg/artifact"
	"github.com/wavesoftware/go-magetasks/pkg/artifact/platform"
	"github.com/wavesoftware/go-magetasks/pkg/checks"
	"github.com/wavesoftware/go-magetasks/pkg/git"
)

// Default target is set to binary.
//
//goland:noinspection GoUnusedGlobalVariable
var Default = magetasks.Build //nolint:deadcode,gochecknoglobals

func init() { //nolint:gochecknoinits
	cli := artifact.Binary{
		Metadata: config.Metadata{
			Name: "ght",
		},
		Platforms: []artifact.Platform{
			{OS: platform.Linux, Architecture: platform.AMD64},
			{OS: platform.Linux, Architecture: platform.ARM64},
			{OS: platform.Linux, Architecture: platform.PPC64LE},
			{OS: platform.Linux, Architecture: platform.S390X},
			{OS: platform.Mac, Architecture: platform.AMD64},
			{OS: platform.Mac, Architecture: platform.ARM64},
			{OS: platform.Windows, Architecture: platform.AMD64},
		},
	}
	magetasks.Configure(config.Config{
		Version: &config.Version{
			Path:     metadata.VersionPath(),
			Resolver: git.NewVersionResolver(),
		},
		Artifacts: []config.Artifact{cli},
		Checks:    []config.Task{checks.GolangCiLint()},
	})
}
