package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/1set/gut/yos"
	"github.com/cardil/ghet/pkg/metadata"
	"github.com/cardil/ghet/pkg/output"
	slog "github.com/go-eden/slf4go"
	"github.com/kirsle/configdir"
	"github.com/pkg/errors"
)

const executableMode = 0o755

func downloadAsset(ctx context.Context, asset Asset, args Args) error {
	l := output.LoggerFrom(ctx).WithFields(slog.Fields{
		"asset": asset.Name,
	})
	cachePath := configdir.LocalCache(metadata.Name)
	if err := configdir.MakePath(cachePath); err != nil {
		return errors.WithStack(err)
	}
	cachePath = path.Join(cachePath, fmt.Sprintf("%d", asset.ID))

	if fileExists(l, cachePath, asset.Size) {
		l.WithFields(slog.Fields{"cachePath": cachePath}).
			Debug("Asset already downloaded")
		return copyFile(cachePath, asset, args)
	}

	l.Debug("Downloading asset")
	cl := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
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
	return copyFile(cachePath, asset, args)
}

func copyFile(cachePath string, asset Asset, args Args) error {
	bin := path.Join(args.Destination, args.Asset.FileName.ToString())
	if err := yos.MoveFile(cachePath, bin); err != nil {
		return errors.WithStack(err)
	}
	if asset.ContentType == "application/octet-stream" {
		if err := os.Chmod(bin, executableMode); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func fileExists(l slog.Logger, path string, size int) bool {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.Size() == int64(size) {
			return true
		}
		l.WithFields(slog.Fields{
			"file-info": fi,
			"size":      size,
		}).Trace("File size mismatch")
		_ = os.Remove(path)
		return false
	}
	return !os.IsNotExist(err)
}
