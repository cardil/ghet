package install

import (
	"strings"

	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/github"
)

// Parse parses the installation arguments from the given spec.
//
//	The spec is a string that contains the arguments in a format of
//	"owner/repo[@version][::archive-name][!!binary-name]".
func Parse(spec string) Installation {
	const expectedParts = 2
	parts := strings.SplitN(spec, "@", expectedParts)
	p2 := strings.SplitN(parts[0], "/", expectedParts)
	owner, repo := "", ""
	if len(p2) > 1 {
		owner, repo = p2[0], p2[1]
	}
	version := "latest"
	archive := ""
	binary := ""
	if len(parts) >= expectedParts {
		parts = strings.SplitN(parts[1], "::", expectedParts)
		version = parts[0]
		if len(parts) >= expectedParts {
			archive = parts[1]
		}
	}
	parts = strings.SplitN(archive, "!!", expectedParts)
	if len(parts) >= expectedParts {
		_, binary = parts[0], parts[1]
	}
	if binary == "" {
		binary = repo
	}
	ext := ""
	parts = strings.SplitN(binary, ".", expectedParts)
	if len(parts) > 1 {
		binary, ext = parts[0], parts[1]
	}
	args := Installation{
		Asset: github.Asset{
			FileName: github.FileName{
				BaseName:  binary,
				Extension: ext,
			},
			Release: github.Release{
				Tag: version,
				Repository: github.Repository{
					Owner: owner,
					Repo:  repo,
				},
			},
		},
		Site: config.Site{
			Type: config.TypeGitHub,
		},
	}
	return args.WithDefaults()
}
