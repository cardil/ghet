package ght

import (
	"context"
	"errors"
	"path"
	"regexp"

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
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("install")
		},
	}
}

type installArgs struct {
	site             string
	version          string
	basename         string
	checksums        string
	repo             string
	multipleBinaries bool
	verifyInArchive  bool
}

func (ia *installArgs) defaults() installArgs {
	defs := install.Args{}.WithDefaults()
	return installArgs{
		site:      defs.Site.Address,
		checksums: defs.Checksums.FileName.ToString(),
		version:   defs.Tag,
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
	fl.StringVar(&ia.checksums, "checksums", defs.checksums,
		"a checksums file name")
	fl.BoolVar(&ia.multipleBinaries, "multiple-binaries", defs.multipleBinaries,
		"if set, will extract all binaries from the archive")
	fl.BoolVar(&ia.verifyInArchive, "verify-in-archive", defs.verifyInArchive,
		"if set, will verify the checksums against the binaries in the archive")
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
	args := install.Args{
		Asset: github.Asset{
			FileName: ia.filename(),
			Release: github.Release{
				Tag:        ia.version,
				Repository: ia.repository(),
			},
			Checksums: github.Checksums{
				FileName: ia.checksumsFilename(),
			},
		},
		Site:             cfg.Site(ia.site),
		MultipleBinaries: ia.multipleBinaries,
		VerifyInArchive:  ia.verifyInArchive,
	}
	args = args.WithDefaults()
	return args
}

func (ia *installArgs) repository() github.Repository {
	m := reporRe.FindStringSubmatch(ia.repo)
	return github.Repository{
		Owner: m[1],
		Repo:  m[2],
	}
}

func (ia *installArgs) filename() github.FileName {
	return github.NewFileName(ia.basename)
}

func (ia *installArgs) checksumsFilename() github.FileName {
	return github.NewFileName(ia.checksums)
}
