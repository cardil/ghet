package download

import (
	"context"
	"strings"

	pkggithub "github.com/cardil/ghet/pkg/github"
	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	slog "github.com/go-eden/slf4go"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
)

var ErrNoAssetFound = errors.New("no matching asset found")

type Plan struct {
	Assets []Asset
}

type Asset struct {
	ID          int64
	Name        string
	ContentType string
	Size        int
	URL         string
}

func CreatePlan(ctx context.Context, args Args) (*Plan, error) {
	ctx = output.EnsureLogger(ctx, slog.Fields{
		"owner": args.Owner,
		"repo":  args.Repo,
	})
	log := output.LoggerFrom(ctx)
	client := githubapi.FromContext(ctx)
	var (
		rr  *github.RepositoryRelease
		r   *github.Response
		err error
	)
	if args.Tag == pkggithub.LatestTag {
		log.Debug("Getting latest release")
		rr, r, err = client.Repositories.GetLatestRelease(ctx, args.Owner, args.Repo)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		args.Tag = rr.GetTagName()
	} else {
		log.WithFields(slog.Fields{"tag": args.Tag}).
			Debug("Getting release")
		rr, r, err = client.Repositories.GetReleaseByTag(ctx,
			args.Owner, args.Repo, args.Tag)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	log.WithFields(slog.Fields{
		"response": r,
		"release":  rr,
	}).Trace("Github API response")

	assets := make([]Asset, 0, 1)
	for _, asset := range rr.Assets {
		if assetMatches(asset, args) {
			log.WithFields(slog.Fields{"asset": asset}).
				Debug("Asset matches")
			assets = append(assets, Asset{
				ID:          asset.GetID(),
				Name:        asset.GetName(),
				ContentType: asset.GetContentType(),
				Size:        asset.GetSize(),
				URL:         asset.GetBrowserDownloadURL(),
			})
		}
	}
	if len(assets) == 0 {
		return nil, errors.WithStack(ErrNoAssetFound)
	}
	return &Plan{Assets: assets}, nil
}

func (p Plan) Download(ctx context.Context, args Args) error {
	ctx = output.EnsureLogger(ctx, slog.Fields{
		"owner": args.Owner,
		"repo":  args.Repo,
	})
	for _, asset := range p.Assets {
		if err := downloadAsset(ctx, asset, args); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func assetMatches(asset *github.ReleaseAsset, args Args) bool {
	name := asset.GetName()
	return (name == args.Checksums.ToString()) ||
		(strings.Contains(name, args.Asset.BaseName) &&
			strings.Contains(name, string(args.Architecture)) &&
			strings.Contains(name, string(args.OperatingSystem)))

}
