package ght

import (
	"context"
	"errors"
	"path"
	"regexp"
	"strings"

	"github.com/cardil/ghet/pkg/config"
	"github.com/cardil/ghet/pkg/ghet/install"
	"github.com/cardil/ghet/pkg/github"
	"github.com/spf13/cobra"
)

var (
	errRepoNotGiven = errors.New("repository not given")
	reporRe         = regexp.MustCompile(`^([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)$`)
)

func installCmd(_ *Args) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install an artifact from GitHub release",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("install")
		},
	}
}

type installArgs struct {
	site      string
	version   string
	basename  string
	checksums string
	repo      string
}

func (ia *installArgs) defaults() installArgs {
	return installArgs{
		site:      "github.com",
		checksums: "checksums.txt",
		version:   "latest",
	}
}

func (ia *installArgs) setFlags(c *cobra.Command) {
	defs := ia.defaults()
	fl := c.Flags()
	fl.StringVar(&ia.site, "site",
		defs.site, "a site to download from")
	fl.StringVarP(&ia.version, "version", "v",
		defs.version, "a version to download")
	fl.StringVar(&ia.basename, "basename",
		defs.basename, "a basename of the artifact, "+
			"if not given a repo name will be used")
	c.Args = cobra.ExactArgs(1)
}

func (ia *installArgs) valiadate() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ia.repo = args[0]
		if !reporRe.MatchString(ia.repo) {
			cmd.SilenceUsage = false
			return errRepoNotGiven
		}
		if ia.basename == "" {
			ia.basename = path.Base(ia.repo)
		}
		return nil
	}
}

func (ia *installArgs) parse(ctx context.Context) install.Args {
	cfg := config.FromContext(ctx)
	release := github.Release{
		Tag:        ia.version,
		Repository: ia.repository(),
	}
	return install.Args{
		Asset: github.Asset{
			FileName:        ia.filename(),
			Architecture:    github.CurrentArchitecture(),
			OperatingSystem: github.CurrentOS(),
			Release:         release,
			Checksums: github.Checksums{
				FileName: ia.checksumsFilename(),
				Release:  release,
			},
		},
		Site: cfg.Site(ia.site),
	}
}

func (ia *installArgs) repository() github.Repository {
	m := reporRe.FindStringSubmatch(ia.repo)
	return github.Repository{
		Owner: m[1],
		Repo:  m[2],
	}
}

func (ia *installArgs) filename() github.FileName {
	return toFileName(ia.basename)
}

func (ia *installArgs) checksumsFilename() github.FileName {
	return toFileName(ia.checksums)
}

func toFileName(s string) github.FileName {
	basename := s
	ext := path.Ext(s)
	if ext != "" {
		basename = strings.TrimSuffix(s, ext)
	}
	return github.FileName{
		BaseName:  basename,
		Extension: ext,
	}
}
