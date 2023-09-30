package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"emperror.dev/errors"
	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	"github.com/cardil/ghet/pkg/output/tui"
	slog "github.com/go-eden/slf4go"
)

const (
	executableMode = 0o750
)

type assetInfo struct {
	githubapi.Asset
	number      int
	total       int
	longestName int
}

func (p Plan) downloadAsset(ctx context.Context, asset assetInfo) error {
	l := output.LoggerFrom(ctx).WithFields(slog.Fields{
		"asset": asset.Name,
	})
	cachePath := p.cachePath(ctx, asset.Asset)

	if fileExists(l, cachePath, asset.Size) {
		l.WithFields(slog.Fields{"cachePath": cachePath}).
			Debug("Asset already downloaded")
		return nil
	}

	l.Debug("Downloading asset")
	cl := githubapi.FromContext(ctx).Client()
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
		return errors.WithStack(err)
	}
	defer out.Close()

	format := "📥 %d/%d %s"
	progress := tui.WidgetsFrom(ctx).NewProgress(ctx, asset.Size, tui.Message{
		Text: fmt.Sprintf(format, asset.number, asset.total, asset.Name),
		Size: len(fmt.Sprintf(format, asset.total, asset.total, strings.Repeat("x", asset.longestName))),
	})
	return progress.With(func(pc tui.ProgressControl) error {
		_, err = io.Copy(out, io.TeeReader(resp.Body, pc))
		if err != nil {
			pc.Error(err)
			return errors.WithStack(err)
		}
		return nil
	})
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
