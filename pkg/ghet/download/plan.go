package download

import (
	"context"
	"fmt"
	"os"

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
	Assets []githubapi.Asset
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

	assets := make([]githubapi.Asset, 0, 1)
	log.WithFields(slog.Fields{"assets": namesOf(rr.Assets)}).
		Debug("Checking assets")
	for _, asset := range rr.Assets {
		if args.Asset.Matches(asset.GetName()) {
			a := githubapi.Asset{
				ID:          asset.GetID(),
				Name:        asset.GetName(),
				ContentType: asset.GetContentType(),
				Size:        asset.GetSize(),
				URL:         asset.GetBrowserDownloadURL(),
			}
			log.WithFields(slog.Fields{"asset": a}).Trace("Asset matches")
			assets = append(assets, a)
		}
	}
	index := githubapi.CreateIndex(assets)
	assets = prioritizeArchives(index)
	if len(assets) == 0 {
		return nil, errors.WithStack(ErrNoAssetFound)
	}
	plan := &Plan{Assets: assets}
	log.WithFields(slog.Fields{"plan": plan}).Debug("Plan created")
	widgets.Printf(ctx, "🎉 Found %s matching assets for %s",
		color.Cyan.Sprint(len(assets)), color.Cyan.Sprintf(rr.GetTagName()))
	return plan, nil
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
		if err := p.downloadAsset(ctx, ai); err != nil {
			return err
		}
	}
	if !args.VerifyInArchive {
		if err := p.verifyChecksums(ctx); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(args.Destination, executableMode); err != nil {
		return unexpected(err)
	}
	if err := p.extractArchives(ctx, args); err != nil {
		return err
	}
	if err := p.moveBinaries(ctx, args); err != nil {
		return err
	}

	return p.cleanCache(ctx)
}

func prioritizeArchives(idx githubapi.IndexedAssets) []githubapi.Asset {
	if len(idx.Archives) > 0 && len(idx.Binaries) > 0 {
		assets := make([]githubapi.Asset, 0, len(idx.Archives)+len(idx.Checksums))
		assets = append(assets, idx.Archives...)
		return append(assets, idx.Checksums...)
	}
	assets := make([]githubapi.Asset, 0, len(idx.Archives)+len(idx.Binaries)+len(idx.Checksums))
	assets = append(assets, idx.Binaries...)
	assets = append(assets, idx.Archives...)
	assets = append(assets, idx.Checksums...)
	return assets
}

func namesOf(assets []*github.ReleaseAsset) []string {
	names := make([]string, 0, len(assets))
	for _, asset := range assets {
		names = append(names, asset.GetName())
	}
	return names
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
