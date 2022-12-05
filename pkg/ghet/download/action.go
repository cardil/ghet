package download

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"net/http"

	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/metadata"
	"github.com/google/go-github/v48/github"
	"github.com/kirsle/configdir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/1set/gut/yos"
)

var ErrNoAssetFound = errors.New("no matching asset found")

func Action(ctx context.Context, args Args) error {
	l := log.
		WithField("owner", args.Owner).
		WithField("repo", args.Repo)
	client := githubapi.NewClient(ctx, args.EffectiveToken())
	var (
		rr  *github.RepositoryRelease
		r   *github.Response
		err error
	)
	if args.Tag == "latest" {
		l.Debug("Getting latest release")
		rr, r, err = client.Repositories.GetLatestRelease(ctx, args.Owner, args.Repo)
		if err != nil {
			return errors.WithStack(err)
		}
		args.Tag = *rr.TagName
	} else {
		l.WithField("Tag", args.Tag).
			Debug("Getting release")
		rr, r, err = client.Repositories.GetReleaseByTag(ctx, args.Owner, args.Repo, args.Tag)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	l.WithField("response", r).
		WithField("release", rr).
		Trace("Github API response")

	for _, asset := range rr.Assets {
		if assetMatches(asset, args) {
			l = l.WithField("asset", asset)
			l.Debug("Asset matches")
			return downloadAsset(ctx, l, asset, args)
		}
	}
	return errors.WithStack(ErrNoAssetFound)
}

func downloadAsset(
	ctx context.Context,
	l *log.Entry,
	asset *github.ReleaseAsset,
	args Args,
) error {
	cachePath := configdir.LocalCache(metadata.Name)
	if err := configdir.MakePath(cachePath); err != nil {
		return errors.WithStack(err)
	}
	cachePath = path.Join(cachePath, fmt.Sprintf("%d", asset.GetID()))

	if fileExists(l, cachePath, asset.GetSize()) {
		l.WithField("cachePath", cachePath).
			Debug("Asset already downloaded")
		return copyFile(cachePath, args)
	}

	l.Info("Downloading asset")
	cl := http.Client{}
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, asset.GetBrowserDownloadURL(), nil)
	if err != nil {
		return errors.WithStack(err)
	}
	resp, err := cl.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: unexpected status code: %d",
			ErrNoAssetFound, resp.StatusCode)
	}

	out, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}
	return copyFile(cachePath, args)
}

func copyFile(cachePath string, args Args) error {
	bin := path.Join(args.Destination, args.BaseName)
	if err := yos.MoveFile(cachePath, bin); err != nil {
		return errors.WithStack(err)
	}
	if err := os.Chmod(bin, 0o755); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func assetMatches(asset *github.ReleaseAsset, args Args) bool {
	return strings.Contains(asset.GetName(), args.BaseName) &&
		strings.Contains(asset.GetName(), string(args.Architecture)) &&
		strings.Contains(asset.GetName(), string(args.OperatingSystem))
}

func fileExists(l *log.Entry, path string, size int) bool {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.Size() == int64(size) {
			return true
		}
		l.WithField("file-info", fi).
			WithField("size", size).
			Trace("File size mismatch")
		_ = os.Remove(path)
		return false
	}
	return !os.IsNotExist(err)
}
