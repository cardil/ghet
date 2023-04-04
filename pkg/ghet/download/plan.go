package download

import (
	"context"
	"fmt"
	"strings"

	pkggithub "github.com/cardil/ghet/pkg/github"
	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	"github.com/cardil/ghet/pkg/output/tui"
	slog "github.com/go-eden/slf4go"
	"github.com/google/go-github/v48/github"
	"github.com/gookit/color"
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
	widgets := tui.WidgetsFrom(ctx)
	spin := widgets.NewSpinner(ctx,
		fmt.Sprintf("⛳️ Getting information about %s release",
			color.Cyan.Sprintf(args.Tag)),
	)
	if err = spin.With(func(spinner tui.Spinner) error {
		rr, r, err = fetchRelease(ctx, args, client)
		return err
	}); err != nil {
		return nil, err
	}

	log.WithFields(slog.Fields{
		"response": r,
		"release":  rr,
	}).Trace("Github API response")

	assets := make([]Asset, 0, 1)
	for _, asset := range rr.Assets {
		if assetMatches(asset, args) {
			log.WithFields(slog.Fields{"asset": asset}).
				Trace("Asset matches")
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
	log.WithFields(slog.Fields{"assets": len(assets)}).
		Debug("Plan created")
	widgets.Printf(ctx, "🎉 Found %s matching assets", color.Cyan.Sprint(len(assets)))
	return &Plan{Assets: assets}, nil
}

func (p Plan) Download(ctx context.Context, args Args) error {
	ctx = output.EnsureLogger(ctx, slog.Fields{
		"owner": args.Owner,
		"repo":  args.Repo,
	})
	longestName := 0
	for _, asset := range p.Assets {
		nameLen := len(asset.Name)
		if nameLen > longestName {
			longestName = nameLen
		}
	}
	for i, asset := range p.Assets {
		ai := assetInfo{
			Asset:       asset,
			number:      i + 1,
			total:       len(p.Assets),
			longestName: longestName,
		}
		if err := downloadAsset(ctx, ai, args); err != nil {
			return err
		}
	}
	return nil
}

func fetchRelease(
	ctx context.Context, args Args,
	client *github.Client,
) (*github.RepositoryRelease, *github.Response, error) {
	var (
		err error
		rr  *github.RepositoryRelease
		r   *github.Response
	)
	log := output.LoggerFrom(ctx)
	if args.Tag == pkggithub.LatestTag {
		log.Debug("Getting latest release")
		if rr, r, err = client.Repositories.GetLatestRelease(ctx, args.Owner, args.Repo); err != nil {
			return nil, nil, errors.WithStack(err)
		}
		args.Tag = rr.GetTagName()
	} else {
		log.WithFields(slog.Fields{"tag": args.Tag}).
			Debug("Getting release")
		if rr, r, err = client.Repositories.GetReleaseByTag(ctx,
			args.Owner, args.Repo, args.Tag); err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}
	return rr, r, nil
}

func assetMatches(asset *github.ReleaseAsset, args Args) bool {
	name := asset.GetName()
	return name == args.Checksums.ToString() ||
		(strings.Contains(name, args.Asset.BaseName) &&
			args.Architecture.Matches(name) &&
			args.OperatingSystem.Matches(name))

}